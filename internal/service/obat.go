package service

import (
	"context"
	"strings"

	"github.com/twizzcode/apotek-pos/internal/model"
	"github.com/twizzcode/apotek-pos/internal/repository"
)

type ObatService struct {
	Repo *repository.ObatRepo
}

// validasiKonversi memeriksa aturan konversi satuan obat.
func validasiKonversi(
	satuanDasarID int64,
	satuan []model.ObatSatuan,
) error {
	if len(satuan) == 0 {
		return &ValidasiError{
			"Minimal harus ada satu satuan obat",
		}
	}

	adaSatuanDasar := false
	satuanTerpakai := make(map[int64]bool)

	for _, item := range satuan {
		if item.SatuanID <= 0 ||
			item.Isi <= 0 ||
			item.HargaJual < 0 {
			return &ValidasiError{
				"Data konversi atau harga jual tidak valid",
			}
		}

		if satuanTerpakai[item.SatuanID] {
			return &ValidasiError{
				"Satuan tidak boleh dipilih lebih dari sekali",
			}
		}

		satuanTerpakai[item.SatuanID] = true

		if item.SatuanID == satuanDasarID && item.Isi == 1 {
			adaSatuanDasar = true
		}
	}

	if !adaSatuanDasar {
		return &ValidasiError{
			"Satuan dasar harus memiliki konversi isi 1",
		}
	}

	return nil
}

// Daftar mengambil daftar obat berdasarkan pencarian dan status.
func (s *ObatService) Daftar(
	ctx context.Context,
	query string,
	status string,
) ([]model.Obat, error) {
	return s.Repo.Semua(ctx, query, status)
}

// Katalog mengambil daftar obat yang tersedia untuk transaksi.
func (s *ObatService) Katalog(
	ctx context.Context,
) ([]model.Obat, error) {
	return s.Repo.Katalog(ctx)
}

// Detail mengambil informasi lengkap obat.
func (s *ObatService) Detail(ctx context.Context, id int64) (model.DetailObat, error) {
	return s.Repo.Detail(ctx, id)
}

// DaftarSatuan mengambil pilihan satuan.
func (s *ObatService) DaftarSatuan(ctx context.Context) ([]model.Satuan, error) {
	return s.Repo.Satuan(ctx)
}

func (s *ObatService) Buat(
	ctx context.Context,
	nama string,
	satuanDasarID int64,
	aktif bool,
	satuan []model.ObatSatuan,
) error {
	nama = strings.TrimSpace(nama)

	if nama == "" {
		return &ValidasiError{"Nama obat wajib diisi"}
	}

	if satuanDasarID <= 0 {
		return &ValidasiError{"Satuan dasar wajib dipilih"}
	}

	if err := validasiKonversi(satuanDasarID, satuan); err != nil {
		return err
	}

	obat := model.Obat{
		Nama:          nama,
		SatuanDasarID: satuanDasarID,
		Aktif:         aktif,
	}

	return s.Repo.Buat(ctx, obat, satuan)
}

func (s *ObatService) Update(
	ctx context.Context,
	id int64,
	nama string,
	satuanDasarID int64,
	aktif bool,
	satuan []model.ObatSatuan,
) error {
	nama = strings.TrimSpace(nama)

	if id <= 0 || nama == "" || satuanDasarID <= 0 {
		return &ValidasiError{
			"Data obat wajib diisi dengan benar",
		}
	}

	if err := validasiKonversi(satuanDasarID, satuan); err != nil {
		return err
	}

	obat := model.Obat{
		ID:            id,
		Nama:          nama,
		SatuanDasarID: satuanDasarID,
		Aktif:         aktif,
	}

	return s.Repo.Update(ctx, id, obat, satuan)
}

// UbahStatus mengubah status aktif obat.
func (s *ObatService) UbahStatus(
	ctx context.Context,
	id int64,
	aktif bool,
) error {
	if id <= 0 {
		return &ValidasiError{"ID obat tidak valid"}
	}

	return s.Repo.UbahStatus(ctx, id, aktif)
}

// Hapus melakukan soft delete pada obat.
func (s *ObatService) Hapus(
	ctx context.Context,
	id int64,
) error {
	if id <= 0 {
		return &ValidasiError{"ID obat tidak valid"}
	}

	return s.Repo.Hapus(ctx, id)
}
