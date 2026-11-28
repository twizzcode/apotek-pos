package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/twizzcode/apotek-pos/internal/model"
)

type opnameRepository interface {
	Daftar(context.Context, model.OpnameFilter) ([]model.Opname, error)
	ByID(context.Context, int64) (model.Opname, []model.OpnameItem, error)
	BatchUntukOpname(context.Context) ([]model.BatchOpname, error)
	Buat(context.Context, model.OpnameInput) (int64, error)
}

type OpnameService struct{ Repo opnameRepository }

const layoutTanggal = "2006-01-02"

func (s *OpnameService) Daftar(ctx context.Context, dari, sampai, kode string) ([]model.Opname, error) {
	f := model.OpnameFilter{Kode: strings.TrimSpace(kode)}
	if dari = strings.TrimSpace(dari); dari != "" {
		t, err := time.ParseInLocation(layoutTanggal, dari, time.Local)
		if err != nil {
			return nil, &ValidasiError{"Tanggal \"dari\" tidak valid"}
		}
		f.Dari = &t
	}
	if sampai = strings.TrimSpace(sampai); sampai != "" {
		t, err := time.ParseInLocation(layoutTanggal, sampai, time.Local)
		if err != nil {
			return nil, &ValidasiError{"Tanggal \"sampai\" tidak valid"}
		}
		f.Sampai = &t
	}
	if f.Dari != nil && f.Sampai != nil && f.Dari.After(*f.Sampai) {
		return nil, &ValidasiError{"Tanggal \"dari\" tidak boleh setelah \"sampai\""}
	}
	return s.Repo.Daftar(ctx, f)
}

func (s *OpnameService) ByID(ctx context.Context, id int64) (model.Opname, []model.OpnameItem, error) {
	if id <= 0 {
		return model.Opname{}, nil, &ValidasiError{"ID opname tidak valid"}
	}
	return s.Repo.ByID(ctx, id)
}

func (s *OpnameService) BatchUntukOpname(ctx context.Context) ([]model.BatchOpname, error) {
	return s.Repo.BatchUntukOpname(ctx)
}

func (s *OpnameService) Buat(ctx context.Context, userID int64, tanggal, catatan string, items []model.OpnameItemInput) (int64, error) {
	if userID <= 0 {
		return 0, &ValidasiError{"Pengguna tidak valid"}
	}
	tgl, err := time.ParseInLocation(layoutTanggal, strings.TrimSpace(tanggal), time.Local)
	if err != nil {
		return 0, &ValidasiError{"Tanggal opname tidak valid"}
	}
	if len(items) == 0 {
		return 0, &ValidasiError{"Isi stok fisik minimal satu batch"}
	}

	seen := make(map[int64]bool, len(items))
	bersih := make([]model.OpnameItemInput, 0, len(items))
	for _, it := range items {
		if it.BatchID <= 0 {
			return 0, &ValidasiError{"Batch tidak valid"}
		}
		if it.StokFisik < 0 {
			return 0, &ValidasiError{"Stok fisik tidak boleh negatif"}
		}
		if seen[it.BatchID] {
			return 0, &ValidasiError{"Batch tidak boleh dihitung dua kali"}
		}
		seen[it.BatchID] = true
		it.Alasan = strings.TrimSpace(it.Alasan)
		if len([]rune(it.Alasan)) > 255 {
			return 0, &ValidasiError{"Alasan maksimal 255 karakter"}
		}
		bersih = append(bersih, it)
	}

	sort.Slice(bersih, func(i, j int) bool { return bersih[i].BatchID < bersih[j].BatchID })

	id, err := s.Repo.Buat(ctx, model.OpnameInput{
		Tanggal: tgl, UserID: userID, Catatan: strings.TrimSpace(catatan), Items: bersih,
	})
	if err != nil {
		if errors.Is(err, model.ErrAlasanWajib) {
			return 0, &ValidasiError{err.Error()}
		}
		if errors.Is(err, model.ErrNotFound) {
			return 0, &ValidasiError{"Batch tidak ditemukan"}
		}
		return 0, err
	}
	return id, nil
}
