package handler

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/twizzcode/apotek-pos/internal/model"
	"github.com/twizzcode/apotek-pos/internal/service"
)

type PBFHandler struct {
	Svc  *service.PBFService
	View *Renderer
}

type PBFPageData struct {
	Daftar []model.PBF
	Form   model.PBF
	Edit   bool
}

func (h *PBFHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	pbfs, err := h.Svc.Daftar(r.Context())
	if err != nil {
		gagal(w, err)
		return
	}
	h.render(w, r, PageData{
		Title: "PBF",
		Pesan: r.URL.Query().Get("pesan"),
		Data:  PBFPageData{Daftar: pbfs},
	})
}

func (h *PBFHandler) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}
	pbf, err := h.Svc.ByID(r.Context(), id)
	if err != nil {
		h.tanganiError(w, err)
		return
	}
	pbfs, err := h.Svc.Daftar(r.Context())
	if err != nil {
		gagal(w, err)
		return
	}
	h.render(w, r, PageData{
		Title: "Edit PBF",
		Pesan: r.URL.Query().Get("pesan"),
		Data:  PBFPageData{Daftar: pbfs, Form: pbf, Edit: true},
	})
}

func (h *PBFHandler) Buat(w http.ResponseWriter, r *http.Request) {
	nama, telepon, alamat := r.FormValue("nama"), r.FormValue("telepon"), r.FormValue("alamat")
	if err := h.Svc.Buat(r.Context(), nama, telepon, alamat); err != nil {
		if isValidasi(err) {
			h.tanganiFormError(w, r, model.PBF{Nama: nama, Telepon: telepon, Alamat: alamat}, err)
		} else {
			h.tanganiError(w, err)
		}
		return
	}
	redirectPBF(w, r, "PBF berhasil ditambahkan")
}

func (h *PBFHandler) Ubah(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}
	nama, telepon, alamat := r.FormValue("nama"), r.FormValue("telepon"), r.FormValue("alamat")
	if err := h.Svc.Ubah(r.Context(), id, nama, telepon, alamat); err != nil {
		if isValidasi(err) {
			pbf := model.PBF{ID: id, Nama: nama, Telepon: telepon, Alamat: alamat}
			h.tanganiFormError(w, r, pbf, err)
			return
		}
		h.tanganiError(w, err)
		return
	}
	redirectPBF(w, r, "PBF berhasil diperbarui")
}

func (h *PBFHandler) Status(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}
	var aktif bool
	switch r.FormValue("aktif") {
	case "1":
		aktif = true
	case "0":
		aktif = false
	default:
		http.Error(w, "Status tidak valid", http.StatusBadRequest)
		return
	}
	if err := h.Svc.SetAktif(r.Context(), id, aktif); err != nil {
		h.tanganiError(w, err)
		return
	}
	redirectPBF(w, r, "Status PBF berhasil diubah")
}

func (h *PBFHandler) tanganiFormError(w http.ResponseWriter, r *http.Request, form model.PBF, validationErr error) {
	pbfs, listErr := h.Svc.Daftar(r.Context())
	if listErr != nil {
		gagal(w, listErr)
		return
	}
	var validasi *service.ValidasiError
	if !errors.As(validationErr, &validasi) {
		gagal(w, validationErr)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	title := "PBF"
	if form.ID > 0 {
		title = "Edit PBF"
	}
	h.render(w, r, PageData{
		Title: title,
		Error: validasi.Error(),
		Data:  PBFPageData{Daftar: pbfs, Form: form, Edit: form.ID > 0},
	})
}

func (h *PBFHandler) tanganiError(w http.ResponseWriter, err error) {
	if errors.Is(err, model.ErrNotFound) {
		http.Error(w, "PBF tidak ditemukan", http.StatusNotFound)
		return
	}
	var validasi *service.ValidasiError
	if errors.As(err, &validasi) {
		http.Error(w, validasi.Error(), http.StatusBadRequest)
		return
	}
	gagal(w, err)
}

func (h *PBFHandler) render(w http.ResponseWriter, r *http.Request, data PageData) {
	h.View.Tampil(w, r, "pbf/list.html", data)
}

func isValidasi(err error) bool {
	var validasi *service.ValidasiError
	return errors.As(err, &validasi)
}

func redirectPBF(w http.ResponseWriter, r *http.Request, message string) {
	query := url.Values{"pesan": []string{message}}
	http.Redirect(w, r, "/pbf?"+query.Encode(), http.StatusSeeOther)
}
