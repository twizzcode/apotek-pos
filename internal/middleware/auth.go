package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/twizzcode/apotek-pos/internal/model"
	"github.com/twizzcode/apotek-pos/internal/repository"
)

const sessionCookie = "apotek_session"

type userKey struct{}

type Auth struct {
	Repo   *repository.UserRepo
	Secret []byte
	Secure bool
}

func User(r *http.Request) (model.User, bool) {
	u, ok := r.Context().Value(userKey{}).(model.User)
	return u, ok
}

func (a *Auth) SetSession(w http.ResponseWriter, userID int64) {
	exp := time.Now().Add(8 * time.Hour).Unix()
	payload := fmt.Sprintf("%d|%d", userID, exp)
	value := payload + "|" + a.sign(payload)
	http.SetCookie(w, &http.Cookie{
		Name: sessionCookie, Value: base64.RawURLEncoding.EncodeToString([]byte(value)), Path: "/",
		Expires: time.Unix(exp, 0), MaxAge: int((8 * time.Hour).Seconds()), HttpOnly: true,
		Secure: a.Secure, SameSite: http.SameSiteStrictMode,
	})
}

func (a *Auth) ClearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: a.Secure, SameSite: http.SameSiteStrictMode})
}

func (a *Auth) LoadUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := a.readSession(r)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		u, err := a.Repo.ByID(r.Context(), id)
		if err != nil || !u.Aktif {
			a.ClearSession(w)
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, u)))
	})
}

func (a *Auth) RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := User(r); !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *Auth) RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, role := range roles {
		allowed[role] = true
	}
	return func(next http.Handler) http.Handler {
		return a.RequireLogin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, _ := User(r)
			if !allowed[u.Role] {
				http.Redirect(w, r, "/akses-ditolak", http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		}))
	}
}

func (a *Auth) readSession(r *http.Request) (int64, bool) {
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return 0, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return 0, false
	}
	parts := strings.Split(string(raw), "|")
	if len(parts) != 3 {
		return 0, false
	}
	payload := parts[0] + "|" + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(a.sign(payload))) {
		return 0, false
	}
	id, errID := strconv.ParseInt(parts[0], 10, 64)
	exp, errExp := strconv.ParseInt(parts[1], 10, 64)
	return id, errID == nil && errExp == nil && time.Now().Unix() < exp
}

func (a *Auth) sign(payload string) string {
	h := hmac.New(sha256.New, a.Secret)
	h.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
