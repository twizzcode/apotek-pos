package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/twizzcode/apotek-pos/internal/middleware"
	"github.com/twizzcode/apotek-pos/internal/service"
)

type UserHandler struct {
	Svc  *service.UserService
	View *Renderer
}

func (h *UserHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	users, err := h.Svc.Daftar(r.Context())
	if err != nil {
		gagal(w, err)
		return
	}
	h.View.Tampil(w, r, "users/list.html", PageData{Title: "Pengguna", Pesan: r.URL.Query().Get("pesan"), Data: users})
}

func (h *UserHandler) Buat(w http.ResponseWriter, r *http.Request) {
	err := h.Svc.Buat(r.Context(), r.FormValue("nama"), r.FormValue("username"), r.FormValue("password"), r.FormValue("role"))
	if err != nil {
		h.tampilError(w, r, err)
		return
	}
	http.Redirect(w, r, "/users?pesan=Pengguna+berhasil+dibuat", http.StatusSeeOther)
}

func (h *UserHandler) Status(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}
	actor, _ := middleware.User(r)
	err = h.Svc.SetAktif(r.Context(), id, actor.ID, r.FormValue("aktif") == "1")
	if err != nil {
		h.tampilError(w, r, err)
		return
	}
	http.Redirect(w, r, "/users?pesan=Status+berhasil+diubah", http.StatusSeeOther)
}

func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}
	if err := h.Svc.ResetPassword(r.Context(), id, r.FormValue("password")); err != nil {
		h.tampilError(w, r, err)
		return
	}
	http.Redirect(w, r, "/users?pesan=Password+berhasil+direset", http.StatusSeeOther)
}

func (h *UserHandler) tampilError(w http.ResponseWriter, r *http.Request, err error) {
	var validasi *service.ValidasiError
	if !errors.As(err, &validasi) {
		gagal(w, err)
		return
	}
	users, listErr := h.Svc.Daftar(r.Context())
	if listErr != nil {
		gagal(w, listErr)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	h.View.Tampil(w, r, "users/list.html", PageData{Title: "Pengguna", Error: validasi.Error(), Data: users})
}

func gagal(w http.ResponseWriter, err error) {
	log.Printf("handler: %v", err)
	http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
}
