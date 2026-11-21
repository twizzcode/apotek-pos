package handler

import (
	"log"
	"net/http"

	"github.com/twizzcode/apotek-pos/internal/repository"
)

type DashboardHandler struct {
	Repo *repository.DashboardRepo
	View *Renderer
}

func (h *DashboardHandler) Home(w http.ResponseWriter, r *http.Request) {
	d, err := h.Repo.Ringkasan(r.Context())
	if err != nil {
		log.Printf("dashboard: %v", err)
		http.Error(w, "Terjadi kesalahan pada server", 500)
		return
	}
	h.View.Tampil(w, r, "dashboard.html", PageData{Title: "Dashboard", Data: d})
}

func (h *DashboardHandler) AksesDitolak(w http.ResponseWriter, r *http.Request) {
	h.View.Tampil(w, r, "akses-ditolak.html", PageData{Title: "Akses Tidak Tersedia"})
}
