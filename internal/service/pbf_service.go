package service

import (
	"context"
	"strings"

	"github.com/twizzcode/apotek-pos/internal/model"
)

type pbfRepository interface {
	Semua(context.Context) ([]model.PBF, error)
	Aktif(context.Context) ([]model.PBF, error)
	ByID(context.Context, int64) (model.PBF, error)
	Buat(context.Context, model.PBF) error
	Ubah(context.Context, model.PBF) error
	SetAktif(context.Context, int64, bool) error
}

type PBFService struct{ Repo pbfRepository }

func (s *PBFService) Daftar(ctx context.Context) ([]model.PBF, error) {
	return s.Repo.Semua(ctx)
}

func (s *PBFService) DaftarAktif(ctx context.Context) ([]model.PBF, error) {
	return s.Repo.Aktif(ctx)
}

func (s *PBFService) ByID(ctx context.Context, id int64) (model.PBF, error) {
	if id <= 0 {
		return model.PBF{}, &ValidasiError{"ID PBF tidak valid"}
	}
	return s.Repo.ByID(ctx, id)
}

func (s *PBFService) Buat(ctx context.Context, nama, telepon, alamat string) error {
	pbf, err := dataValid(nama, telepon, alamat)
	if err != nil {
		return err
	}
	pbf.Aktif = true
	return s.Repo.Buat(ctx, pbf)
}

func (s *PBFService) Ubah(ctx context.Context, id int64, nama, telepon, alamat string) error {
	if id <= 0 {
		return &ValidasiError{"ID PBF tidak valid"}
	}
	pbf, err := dataValid(nama, telepon, alamat)
	if err != nil {
		return err
	}
	pbf.ID = id
	return s.Repo.Ubah(ctx, pbf)
}

func (s *PBFService) SetAktif(ctx context.Context, id int64, aktif bool) error {
	if id <= 0 {
		return &ValidasiError{"ID PBF tidak valid"}
	}
	return s.Repo.SetAktif(ctx, id, aktif)
}

func dataValid(nama, telepon, alamat string) (model.PBF, error) {
	nama = strings.TrimSpace(nama)
	if nama == "" {
		return model.PBF{}, &ValidasiError{"Nama PBF wajib diisi"}
	}
	return model.PBF{
		Nama:    nama,
		Telepon: strings.TrimSpace(telepon),
		Alamat:  strings.TrimSpace(alamat),
	}, nil
}
