package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/twizzcode/apotek-pos/internal/model"
	"github.com/twizzcode/apotek-pos/internal/service"
)

type ObatHandler struct {
	Svc  *service.ObatService
	View *Renderer
}

// Data untuk halaman form obat.
type ObatFormData struct {
	Obat   model.DetailObat
	Satuan []model.Satuan
}

// Daftar menampilkan seluruh obat dengan pencarian dan filter status.
func (h *ObatHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	status := r.URL.Query().Get("status")

	// Validasi pilihan status.
	if status != "aktif" && status != "nonaktif" {
		status = ""
	}

	obat, err := h.Svc.Daftar(r.Context(), query, status)
	if err != nil {
		gagal(w, err)
		return
	}

	h.View.Tampil(w, r, "obat/list.html", PageData{
		Title:  "Obat & Satuan",
		Pesan:  r.URL.Query().Get("pesan"),
		Data:   obat,
		Query:  query,
		Status: status,
	})
}

// FormTambah menampilkan form penambahan obat.
func (h *ObatHandler) FormTambah(w http.ResponseWriter, r *http.Request) {
	satuan, err := h.Svc.DaftarSatuan(r.Context())
	if err != nil {
		gagal(w, err)
		return
	}

	h.View.Tampil(w, r, "obat/form.html", PageData{
		Title: "Tambah Obat",
		Data: ObatFormData{
			Satuan: satuan,
		},
	})
}

// Detail menampilkan informasi lengkap obat.
func (h *ObatHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID obat tidak valid", http.StatusBadRequest)
		return
	}

	obat, err := h.Svc.Detail(r.Context(), id)
	if err != nil {
		if err == model.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		gagal(w, err)
		return
	}

	h.View.Tampil(w, r, "obat/detail.html", PageData{
		Title: "Detail Obat",
		Data:  obat,
	})
}

// FormEdit menampilkan form perubahan obat.
func (h *ObatHandler) FormEdit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID obat tidak valid", http.StatusBadRequest)
		return
	}

	obat, err := h.Svc.Detail(r.Context(), id)
	if err != nil {
		if err == model.ErrNotFound {
			http.NotFound(w, r)
			return
		}
		gagal(w, err)
		return
	}

	satuan, err := h.Svc.DaftarSatuan(r.Context())
	if err != nil {
		gagal(w, err)
		return
	}

	h.View.Tampil(w, r, "obat/form.html", PageData{
		Title: "Edit Obat",
		Data: ObatFormData{
			Obat:   obat,
			Satuan: satuan,
		},
	})
}

// Update memproses perubahan data obat.
func (h *ObatHandler) Update(w http.ResponseWriter, r *http.Request) {
	// Mengambil ID obat dari URL.
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID obat tidak valid", http.StatusBadRequest)
		return
	}

	// Membaca data form.
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form tidak valid", http.StatusBadRequest)
		return
	}

	nama := r.FormValue("nama")

	satuanDasarID, err := strconv.ParseInt(
		r.FormValue("satuan_dasar_id"), 10, 64,
	)
	if err != nil {
		http.Error(w, "Satuan dasar tidak valid", http.StatusBadRequest)
		return
	}

	aktif := r.FormValue("aktif") == "1"

	// Mengambil daftar konversi satuan.
	satuanIDs := r.Form["satuan_id"]
	isis := r.Form["isi"]
	hargas := r.Form["harga_jual"]

	if len(satuanIDs) != len(isis) ||
		len(satuanIDs) != len(hargas) {
		http.Error(w, "Data konversi tidak lengkap", http.StatusBadRequest)
		return
	}

	var satuan []model.ObatSatuan

	for i := range satuanIDs {
		satuanID, err := strconv.ParseInt(satuanIDs[i], 10, 64)
		if err != nil || satuanID <= 0 {
			http.Error(w, "Satuan tidak valid", http.StatusBadRequest)
			return
		}

		isi, err := strconv.ParseInt(isis[i], 10, 64)
		if err != nil || isi <= 0 {
			http.Error(w, "Isi konversi tidak valid", http.StatusBadRequest)
			return
		}

		harga, err := strconv.ParseInt(hargas[i], 10, 64)
		if err != nil || harga < 0 {
			http.Error(w, "Harga jual tidak valid", http.StatusBadRequest)
			return
		}

		satuan = append(satuan, model.ObatSatuan{
			SatuanID:  satuanID,
			Isi:       isi,
			HargaJual: harga,
		})
	}

	// Memanggil service untuk memperbarui obat.
	err = h.Svc.Update(
		r.Context(),
		id,
		nama,
		satuanDasarID,
		aktif,
		satuan,
	)

	if err != nil {
		gagal(w, err)
		return
	}

	// Kembali ke halaman detail obat.
	http.Redirect(
		w,
		r,
		"/obat/"+strconv.FormatInt(id, 10)+"?pesan=Obat+berhasil+diperbarui",
		http.StatusSeeOther,
	)
}

// Buat memproses penambahan obat baru.
func (h *ObatHandler) Buat(w http.ResponseWriter, r *http.Request) {
	// Membaca data form.
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form tidak valid", http.StatusBadRequest)
		return
	}

	nama := r.FormValue("nama")

	// Mengambil satuan dasar.
	satuanDasarID, err := strconv.ParseInt(
		r.FormValue("satuan_dasar_id"), 10, 64,
	)
	if err != nil || satuanDasarID <= 0 {
		http.Error(w, "Satuan dasar tidak valid", http.StatusBadRequest)
		return
	}

	aktif := r.FormValue("aktif") == "1"

	// Mengambil data konversi satuan.
	satuanIDs := r.Form["satuan_id"]
	isis := r.Form["isi"]
	hargas := r.Form["harga_jual"]

	if len(satuanIDs) != len(isis) ||
		len(satuanIDs) != len(hargas) {
		http.Error(w, "Data konversi tidak lengkap", http.StatusBadRequest)
		return
	}

	var satuan []model.ObatSatuan

	for i := range satuanIDs {
		satuanID, err := strconv.ParseInt(satuanIDs[i], 10, 64)
		if err != nil || satuanID <= 0 {
			http.Error(w, "Satuan tidak valid", http.StatusBadRequest)
			return
		}

		isi, err := strconv.ParseInt(isis[i], 10, 64)
		if err != nil || isi <= 0 {
			http.Error(w, "Isi konversi tidak valid", http.StatusBadRequest)
			return
		}

		harga, err := strconv.ParseInt(hargas[i], 10, 64)
		if err != nil || harga < 0 {
			http.Error(w, "Harga jual tidak valid", http.StatusBadRequest)
			return
		}

		satuan = append(satuan, model.ObatSatuan{
			SatuanID:  satuanID,
			Isi:       isi,
			HargaJual: harga,
		})
	}

	// Menyimpan obat melalui service.
	err = h.Svc.Buat(
		r.Context(),
		nama,
		satuanDasarID,
		aktif,
		satuan,
	)

	if err != nil {
		var validasi *service.ValidasiError

		if errors.As(err, &validasi) {
			pilihanSatuan, listErr := h.Svc.DaftarSatuan(r.Context())
			if listErr != nil {
				gagal(w, listErr)
				return
			}

			w.WriteHeader(http.StatusBadRequest)
			h.View.Tampil(w, r, "obat/form.html", PageData{
				Title: "Tambah Obat",
				Error: validasi.Error(),
				Data: ObatFormData{
					Satuan: pilihanSatuan,
				},
			})
			return
		}

		gagal(w, err)
		return
	}

	// Kembali ke daftar obat setelah berhasil.
	http.Redirect(
		w,
		r,
		"/obat?pesan=Obat+berhasil+ditambahkan",
		http.StatusSeeOther,
	)
}

// Status memproses perubahan status aktif obat.
func (h *ObatHandler) Status(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID obat tidak valid", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form tidak valid", http.StatusBadRequest)
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

	err = h.Svc.UbahStatus(r.Context(), id, aktif)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		gagal(w, err)
		return
	}

	pesan := "Obat berhasil dinonaktifkan"
	if aktif {
		pesan = "Obat berhasil diaktifkan"
	}

	http.Redirect(
		w,
		r,
		"/obat/"+strconv.FormatInt(id, 10)+"?pesan="+
			strings.ReplaceAll(pesan, " ", "+"),
		http.StatusSeeOther,
	)
}

// Hapus memproses soft delete obat.
func (h *ObatHandler) Hapus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID obat tidak valid", http.StatusBadRequest)
		return
	}

	err = h.Svc.Hapus(r.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			http.NotFound(w, r)
			return
		}

		gagal(w, err)
		return
	}

	http.Redirect(
		w,
		r,
		"/obat?pesan=Obat+berhasil+dihapus",
		http.StatusSeeOther,
	)
}
