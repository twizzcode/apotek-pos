package handler

import "net/http"

// PreviewHandler menyediakan kontrak halaman sebelum backend tiap modul terhubung.
// Handler fitur mengganti route ini tanpa mengubah template dan navigasi.
type PreviewHandler struct{ View *Renderer }

func (h *PreviewHandler) Page(file, title string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.View.Tampil(w, r, file, PageData{Title: title, Pesan: r.URL.Query().Get("pesan")})
	}
}
