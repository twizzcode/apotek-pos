package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/twizzcode/apotek-pos/internal/middleware"
	"github.com/twizzcode/apotek-pos/internal/service"
)

type AuthHandler struct {
	Svc  *service.UserService
	Auth *middleware.Auth
	View *Renderer
}

func (h *AuthHandler) FormLogin(w http.ResponseWriter, r *http.Request) {
	if u, ok := middleware.User(r); ok {
		h.redirectMasuk(w, r, u.Role)
		return
	}
	h.View.Tampil(w, r, "login.html", PageData{Title: "Masuk"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	u, err := h.Svc.Login(r.Context(), r.FormValue("username"), r.FormValue("password"))
	if errors.Is(err, service.ErrLogin) {
		h.View.Tampil(w, r, "login.html", PageData{Title: "Masuk", Error: err.Error()})
		return
	}
	if err != nil {
		log.Printf("login: %v", err)
		http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
		return
	}
	h.Auth.SetSession(w, u.ID)
	h.redirectMasuk(w, r, u.Role)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	h.Auth.ClearSession(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *AuthHandler) redirectMasuk(w http.ResponseWriter, r *http.Request, role string) {
	if role == "admin" || role == "apoteker" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/akses-ditolak", http.StatusSeeOther)
}
