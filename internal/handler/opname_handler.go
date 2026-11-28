package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/twizzcode/apotek-pos/internal/middleware"
	"github.com/twizzcode/apotek-pos/internal/model"
	"github.com/twizzcode/apotek-pos/internal/service"
)

type OpnameHandler struct {
	Svc  *service.OpnameService
	View *Renderer
}

type OpnameListData struct {
	Daftar             []model.Opname
	Dari, Sampai, Kode string
}

type OpnameBarisForm struct {
	BatchID    int64
	Obat       string
	NoBatch    string
	Satuan     string
	ExpiredAt  time.Time
	StokSistem int
	Fisik      string
	Alasan     string
}

type OpnameFormData struct {
	Tanggal string
	Catatan string
	Baris   []OpnameBarisForm
}

type OpnameDetailData struct {
	Opname model.Opname
	Items  []model.OpnameItem
	Sesuai int
	Plus   int
	Minus  int
}

func (h *OpnameHandler) Daftar(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	data := OpnameListData{Dari: q.Get("dari"), Sampai: q.Get("sampai"), Kode: q.Get("kode"), Daftar: []model.Opname{}}
	page := PageData{Title: "Stock Opname", Pesan: q.Get("pesan"), Data: data}

	daftar, err := h.Svc.Daftar(r.Context(), data.Dari, data.Sampai, data.Kode)
	if err != nil {
		if msg, ok := opnamePesanValidasi(err); ok {
			page.Error = msg
			w.WriteHeader(http.StatusBadRequest)
			h.View.Tampil(w, r, "opname/list.html", page)
			return
		}
		opnameGagal(w, err)
		return
	}
	data.Daftar = daftar
	page.Data = data
	h.View.Tampil(w, r, "opname/list.html", page)
}

func (h *OpnameHandler) Form(w http.ResponseWriter, r *http.Request) {
	h.tampilForm(w, r, http.StatusOK, time.Now().Format("2006-01-02"), "", "", nil, nil)
}

func (h *OpnameHandler) Simpan(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form tidak valid", http.StatusBadRequest)
		return
	}
	user, _ := middleware.User(r)
	tanggal := r.PostFormValue("tanggal")
	catatan := r.PostFormValue("catatan")

	fisik := map[int64]string{}
	alasan := map[int64]string{}
	var items []model.OpnameItemInput
	errMsg := ""

	for _, raw := range r.PostForm["batch_id"] {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			http.Error(w, "Batch tidak valid", http.StatusBadRequest)
			return
		}
		f := strings.TrimSpace(r.PostFormValue("fisik_" + raw))
		a := strings.TrimSpace(r.PostFormValue("alasan_" + raw))
		fisik[id], alasan[id] = f, a
		if f == "" {
			continue // batch tidak dihitung
		}
		n, err := strconv.Atoi(f)
		if err != nil || n < 0 {
			errMsg = "Stok fisik harus berupa angka 0 atau lebih"
			continue
		}
		items = append(items, model.OpnameItemInput{BatchID: id, StokFisik: n, Alasan: a})
	}
	if errMsg != "" {
		h.tampilForm(w, r, http.StatusBadRequest, tanggal, catatan, errMsg, fisik, alasan)
		return
	}

	id, err := h.Svc.Buat(r.Context(), user.ID, tanggal, catatan, items)
	if err != nil {
		if msg, ok := opnamePesanValidasi(err); ok {
			h.tampilForm(w, r, http.StatusBadRequest, tanggal, catatan, msg, fisik, alasan)
			return
		}
		opnameGagal(w, err)
		return
	}
	pesan := url.QueryEscape("Opname berhasil difinalisasi")
	http.Redirect(w, r, fmt.Sprintf("/opname/%d?pesan=%s", id, pesan), http.StatusSeeOther)
}

func (h *OpnameHandler) Detail(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "ID tidak valid", http.StatusBadRequest)
		return
	}
	opname, items, err := h.Svc.ByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			http.Error(w, "Opname tidak ditemukan", http.StatusNotFound)
			return
		}
		opnameGagal(w, err)
		return
	}
	data := OpnameDetailData{Opname: opname, Items: items}
	for _, it := range items {
		switch {
		case it.Selisih > 0:
			data.Plus += it.Selisih
		case it.Selisih < 0:
			data.Minus += -it.Selisih
		default:
			data.Sesuai++
		}
	}
	h.View.Tampil(w, r, "opname/detail.html", PageData{
		Title: "Detail Opname", Pesan: r.URL.Query().Get("pesan"), Data: data,
	})
}

func (h *OpnameHandler) tampilForm(w http.ResponseWriter, r *http.Request, status int,
	tanggal, catatan, pesanError string, fisik, alasan map[int64]string) {
	batches, err := h.Svc.BatchUntukOpname(r.Context())
	if err != nil {
		opnameGagal(w, err)
		return
	}
	baris := make([]OpnameBarisForm, 0, len(batches))
	for _, b := range batches {
		baris = append(baris, OpnameBarisForm{
			BatchID: b.BatchID, Obat: b.NamaObat, NoBatch: b.NoBatch, Satuan: b.Satuan,
			ExpiredAt: b.ExpiredAt, StokSistem: b.StokSistem,
			Fisik: fisik[b.BatchID], Alasan: alasan[b.BatchID],
		})
	}
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	h.View.Tampil(w, r, "opname/form.html", PageData{
		Title: "Opname Baru", Error: pesanError,
		Data: OpnameFormData{Tanggal: tanggal, Catatan: catatan, Baris: baris},
	})
}

func opnamePesanValidasi(err error) (string, bool) {
	var v *service.ValidasiError
	if errors.As(err, &v) {
		return v.Error(), true
	}
	return "", false
}

func opnameGagal(w http.ResponseWriter, err error) {
	log.Printf("opname: %v", err)
	http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
}
