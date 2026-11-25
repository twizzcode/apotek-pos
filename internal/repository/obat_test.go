package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/twizzcode/apotek-pos/internal/model"
)

// Menguji validasi data sebelum transaksi database dimulai.
func TestObatRepoBuatValidasi(t *testing.T) {
	tests := []struct {
		nama   string
		satuan []model.ObatSatuan
	}{
		{
			nama:   "satuan kosong",
			satuan: nil,
		},
		{
			nama: "ID satuan tidak valid",
			satuan: []model.ObatSatuan{
				{SatuanID: 0, Isi: 1, HargaJual: 1000},
			},
		},
		{
			nama: "isi konversi nol",
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 0, HargaJual: 1000},
			},
		},
		{
			nama: "harga jual negatif",
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 1, HargaJual: -1000},
			},
		},
		{
			nama: "satuan dasar tidak memiliki isi satu",
			satuan: []model.ObatSatuan{
				{SatuanID: 2, Isi: 10, HargaJual: 10000},
			},
		},
	}

	repo := &ObatRepo{}

	for _, tt := range tests {
		t.Run(tt.nama, func(t *testing.T) {
			obat := model.Obat{
				Nama:          "Paracetamol",
				SatuanDasarID: 1,
				Aktif:         true,
			}

			err := repo.Buat(
				context.Background(),
				obat,
				tt.satuan,
			)

			if err == nil {
				t.Fatal("seharusnya menghasilkan error validasi")
			}
		})
	}
}

// Menguji penyimpanan obat dan konversi berhasil.
func TestObatRepoBuatBerhasil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	obat := model.Obat{
		Nama:          "Paracetamol",
		SatuanDasarID: 1,
		Aktif:         true,
	}

	satuan := []model.ObatSatuan{
		{SatuanID: 1, Isi: 1, HargaJual: 2500},
		{SatuanID: 2, Isi: 10, HargaJual: 12000},
		{SatuanID: 3, Isi: 100, HargaJual: 70000},
	}

	mock.ExpectBegin()

	mock.ExpectExec("INSERT INTO obat").
		WithArgs("Paracetamol", int64(1), true).
		WillReturnResult(sqlmock.NewResult(10, 1))

	mock.ExpectExec("INSERT INTO obat_satuan").
		WithArgs(int64(10), int64(1), int64(1), int64(2500)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO obat_satuan").
		WithArgs(int64(10), int64(2), int64(10), int64(12000)).
		WillReturnResult(sqlmock.NewResult(2, 1))

	mock.ExpectExec("INSERT INTO obat_satuan").
		WithArgs(int64(10), int64(3), int64(100), int64(70000)).
		WillReturnResult(sqlmock.NewResult(3, 1))

	mock.ExpectCommit()

	err = repo.Buat(context.Background(), obat, satuan)
	if err != nil {
		t.Fatalf("gagal menyimpan obat: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

// Menguji rollback ketika penyimpanan konversi gagal.
func TestObatRepoBuatRollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	obat := model.Obat{
		Nama:          "Paracetamol",
		SatuanDasarID: 1,
		Aktif:         true,
	}

	satuan := []model.ObatSatuan{
		{SatuanID: 1, Isi: 1, HargaJual: 2500},
		{SatuanID: 2, Isi: 10, HargaJual: 12000},
	}

	mock.ExpectBegin()

	mock.ExpectExec("INSERT INTO obat").
		WithArgs("Paracetamol", int64(1), true).
		WillReturnResult(sqlmock.NewResult(10, 1))

	mock.ExpectExec("INSERT INTO obat_satuan").
		WithArgs(int64(10), int64(1), int64(1), int64(2500)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectExec("INSERT INTO obat_satuan").
		WithArgs(int64(10), int64(2), int64(10), int64(12000)).
		WillReturnError(errors.New("database error"))

	mock.ExpectRollback()

	err = repo.Buat(context.Background(), obat, satuan)
	if err == nil {
		t.Fatal("seharusnya menghasilkan error ketika penyimpanan gagal")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoUpdateBerhasil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	obat := model.Obat{
		ID:            10,
		Nama:          "Paracetamol 500 mg",
		SatuanDasarID: 1,
		Aktif:         true,
	}

	satuan := []model.ObatSatuan{
		{SatuanID: 1, Isi: 1, HargaJual: 2500},
		{SatuanID: 2, Isi: 10, HargaJual: 12000},
	}

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT id").
		WithArgs(int64(10)).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(10),
		)

	mock.ExpectExec("UPDATE obat").
		WithArgs(
			"Paracetamol 500 mg",
			int64(1),
			true,
			int64(10),
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("INSERT INTO obat_satuan").
		WithArgs(int64(10), int64(1), int64(1), int64(2500)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("INSERT INTO obat_satuan").
		WithArgs(int64(10), int64(2), int64(10), int64(12000)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectCommit()

	err = repo.Update(context.Background(), 10, obat, satuan)
	if err != nil {
		t.Fatalf("update obat gagal: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoUpdateTidakDitemukan(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	obat := model.Obat{
		ID:            99,
		Nama:          "Paracetamol",
		SatuanDasarID: 1,
		Aktif:         true,
	}

	satuan := []model.ObatSatuan{
		{SatuanID: 1, Isi: 1, HargaJual: 2500},
	}

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT id").
		WithArgs(int64(99)).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}),
		)

	mock.ExpectRollback()

	err = repo.Update(context.Background(), 99, obat, satuan)

	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("seharusnya menghasilkan ErrNotFound, dapat: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoUpdateRollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	obat := model.Obat{
		ID:            10,
		Nama:          "Paracetamol",
		SatuanDasarID: 1,
		Aktif:         true,
	}

	satuan := []model.ObatSatuan{
		{SatuanID: 1, Isi: 1, HargaJual: 2500},
		{SatuanID: 2, Isi: 10, HargaJual: 12000},
	}

	mock.ExpectBegin()

	mock.ExpectQuery("SELECT id").
		WithArgs(int64(10)).
		WillReturnRows(
			sqlmock.NewRows([]string{"id"}).AddRow(10),
		)

	mock.ExpectExec("UPDATE obat").
		WithArgs("Paracetamol", int64(1), true, int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("INSERT INTO obat_satuan").
		WithArgs(int64(10), int64(1), int64(1), int64(2500)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	mock.ExpectExec("INSERT INTO obat_satuan").
		WithArgs(int64(10), int64(2), int64(10), int64(12000)).
		WillReturnError(errors.New("database error"))

	mock.ExpectRollback()

	err = repo.Update(context.Background(), 10, obat, satuan)

	if err == nil {
		t.Fatal("seharusnya menghasilkan error ketika update gagal")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoUbahStatusBerhasil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	mock.ExpectExec("UPDATE obat").
		WithArgs(false, int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.UbahStatus(context.Background(), 10, false)
	if err != nil {
		t.Fatalf("gagal mengubah status obat: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoUbahStatusTidakDitemukan(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	mock.ExpectExec("UPDATE obat").
		WithArgs(false, int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectQuery("SELECT EXISTS").
		WithArgs(int64(99)).
		WillReturnRows(
			sqlmock.NewRows([]string{"EXISTS"}).AddRow(false),
		)

	err = repo.UbahStatus(context.Background(), 99, false)

	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("seharusnya ErrNotFound, dapat: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoHapusBerhasil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	mock.ExpectExec("UPDATE obat").
		WithArgs(int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = repo.Hapus(context.Background(), 10)
	if err != nil {
		t.Fatalf("gagal menghapus obat: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoHapusTidakDitemukan(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	mock.ExpectExec("UPDATE obat").
		WithArgs(int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.Hapus(context.Background(), 99)

	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("seharusnya ErrNotFound, dapat: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoKatalogBerhasil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	waktu := time.Date(
		2026, 11, 25, 10, 0, 0, 0, time.UTC,
	)

	// Mock query daftar obat aktif.
	mock.ExpectQuery("FROM obat o").
		WithArgs(true).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "nama", "satuan_dasar_id",
				"satuan_dasar", "aktif",
				"created_at", "updated_at",
				"stok_total", "stok_layak",
			}).
				AddRow(
					1, "Paracetamol", 1, "Tablet",
					true, waktu, waktu, 100, 80,
				).
				AddRow(
					2, "Amoxicillin", 1, "Tablet",
					true, waktu, waktu, 0, 0,
				),
		)

	// Mock query konversi satuan obat.
	mock.ExpectQuery("FROM obat_satuan os").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"obat_id", "id", "satuan_id",
				"nama", "isi", "harga_jual",
			}).
				AddRow(1, 1, 1, "Tablet", 1, 2500),
		)

	daftar, err := repo.Katalog(context.Background())
	if err != nil {
		t.Fatalf("gagal mengambil katalog obat: %v", err)
	}

	// Hanya Paracetamol yang memiliki stok layak.
	if len(daftar) != 1 {
		t.Fatalf("jumlah katalog = %d, ingin 1", len(daftar))
	}

	if daftar[0].ID != 1 {
		t.Errorf("ID obat = %d, ingin 1", daftar[0].ID)
	}

	if daftar[0].Nama != "Paracetamol" {
		t.Errorf("nama obat = %q, ingin Paracetamol", daftar[0].Nama)
	}

	if daftar[0].StokLayak != 80 {
		t.Errorf("stok layak = %d, ingin 80", daftar[0].StokLayak)
	}

	if len(daftar[0].Satuan) != 1 {
		t.Fatalf(
			"jumlah satuan = %d, ingin 1",
			len(daftar[0].Satuan),
		)
	}

	if daftar[0].Satuan[0].StokTersedia != 80 {
		t.Errorf(
			"stok tersedia = %d, ingin 80",
			daftar[0].Satuan[0].StokTersedia,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

// Menguji katalog ketika tidak ada obat yang tersedia.
func TestObatRepoKatalogKosong(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	mock.ExpectQuery("SELECT").
		WithArgs(true).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"nama",
				"satuan_dasar_id",
				"satuan_dasar",
				"aktif",
				"created_at",
				"updated_at",
				"stok_total",
				"stok_layak",
			}),
		)

	mock.ExpectQuery("SELECT").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"obat_id",
				"id",
				"satuan_id",
				"nama",
				"isi",
				"harga_jual",
			}),
		)

	daftar, err := repo.Katalog(context.Background())
	if err != nil {
		t.Fatalf("tidak seharusnya menghasilkan error: %v", err)
	}

	if len(daftar) != 0 {
		t.Errorf(
			"jumlah katalog = %d, ingin 0",
			len(daftar),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

// Menguji katalog ketika database mengalami error.
func TestObatRepoKatalogDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	dbError := errors.New("database error")

	mock.ExpectQuery("SELECT").
		WithArgs(true).
		WillReturnError(dbError)

	daftar, err := repo.Katalog(context.Background())

	if err == nil {
		t.Fatal("seharusnya menghasilkan error database")
	}

	if !errors.Is(err, dbError) {
		t.Errorf("error tidak sesuai: %v", err)
	}

	if daftar != nil {
		t.Errorf("daftar seharusnya nil ketika terjadi error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoSemuaBerhasil(t *testing.T) {
	tests := []struct {
		name      string
		pencarian string
		status    string
		obatID    int64
		nama      string
		aktif     bool
	}{
		{
			name:   "tanpa filter",
			obatID: 1,
			nama:   "Paracetamol",
			aktif:  true,
		},
		{
			name:      "pencarian dan filter aktif",
			pencarian: "para",
			status:    "aktif",
			obatID:    1,
			nama:      "Paracetamol",
			aktif:     true,
		},
		{
			name:   "filter nonaktif",
			status: "nonaktif",
			obatID: 2,
			nama:   "Amoxicillin",
			aktif:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			repo := &ObatRepo{DB: db}

			waktu := time.Date(
				2026, 11, 25, 10, 0, 0, 0, time.UTC,
			)

			// Ekspektasi query daftar obat.
			queryObat := mock.ExpectQuery("SELECT")

			if tt.pencarian != "" && tt.status == "aktif" {
				queryObat.WithArgs("%"+tt.pencarian+"%", true)
			} else if tt.status == "aktif" {
				queryObat.WithArgs(true)
			} else if tt.status == "nonaktif" {
				queryObat.WithArgs(false)
			}

			queryObat.WillReturnRows(
				sqlmock.NewRows([]string{
					"id", "nama", "satuan_dasar_id",
					"satuan_dasar", "aktif",
					"created_at", "updated_at",
					"stok_total", "stok_layak",
				}).AddRow(
					tt.obatID,
					tt.nama,
					1,
					"Tablet",
					tt.aktif,
					waktu,
					waktu,
					100,
					80,
				),
			)

			// Ekspektasi query konversi satuan.
			mock.ExpectQuery("SELECT").
				WillReturnRows(
					sqlmock.NewRows([]string{
						"obat_id", "id", "satuan_id",
						"nama", "isi", "harga_jual",
					}).AddRow(
						tt.obatID,
						1,
						1,
						"Tablet",
						1,
						2500,
					),
				)

			daftar, err := repo.Semua(
				context.Background(),
				tt.pencarian,
				tt.status,
			)

			if err != nil {
				t.Fatalf("gagal mengambil daftar obat: %v", err)
			}

			if len(daftar) != 1 {
				t.Fatalf(
					"jumlah obat = %d, ingin 1",
					len(daftar),
				)
			}

			if daftar[0].ID != tt.obatID {
				t.Errorf(
					"ID obat = %d, ingin %d",
					daftar[0].ID,
					tt.obatID,
				)
			}

			if daftar[0].Nama != tt.nama {
				t.Errorf(
					"nama obat = %q, ingin %q",
					daftar[0].Nama,
					tt.nama,
				)
			}

			if daftar[0].Aktif != tt.aktif {
				t.Errorf(
					"status aktif = %v, ingin %v",
					daftar[0].Aktif,
					tt.aktif,
				)
			}

			if daftar[0].StokBlokir != 20 {
				t.Errorf(
					"stok blokir = %d, ingin 20",
					daftar[0].StokBlokir,
				)
			}

			if len(daftar[0].Satuan) != 1 {
				t.Fatalf(
					"jumlah satuan = %d, ingin 1",
					len(daftar[0].Satuan),
				)
			}

			if daftar[0].Satuan[0].Nama != "Tablet" {
				t.Errorf(
					"nama satuan = %q, ingin Tablet",
					daftar[0].Satuan[0].Nama,
				)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf(
					"ekspektasi SQL tidak terpenuhi: %v",
					err,
				)
			}
		})
	}
}

// Menguji daftar obat ketika database tidak memiliki data.
func TestObatRepoSemuaKosong(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	mock.ExpectQuery("SELECT").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "nama", "satuan_dasar_id",
				"satuan_dasar", "aktif",
				"created_at", "updated_at",
				"stok_total", "stok_layak",
			}),
		)

	mock.ExpectQuery("SELECT").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"obat_id", "id", "satuan_id",
				"nama", "isi", "harga_jual",
			}),
		)

	daftar, err := repo.Semua(
		context.Background(),
		"",
		"",
	)

	if err != nil {
		t.Fatalf("seharusnya tidak menghasilkan error: %v", err)
	}

	if len(daftar) != 0 {
		t.Errorf(
			"jumlah obat = %d, ingin 0",
			len(daftar),
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

// Menguji error ketika query daftar obat gagal.
func TestObatRepoSemuaDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	mock.ExpectQuery("SELECT").
		WillReturnError(errors.New("database error"))

	_, err = repo.Semua(
		context.Background(),
		"",
		"",
	)

	if err == nil {
		t.Fatal("seharusnya menghasilkan error database")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

// Menguji error ketika query konversi satuan gagal.
func TestObatRepoSemuaKonversiDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	waktu := time.Date(
		2026, 11, 25, 10, 0, 0, 0, time.UTC,
	)

	mock.ExpectQuery("SELECT").
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "nama", "satuan_dasar_id",
				"satuan_dasar", "aktif",
				"created_at", "updated_at",
				"stok_total", "stok_layak",
			}).AddRow(
				1, "Paracetamol", 1, "Tablet",
				true, waktu, waktu, 100, 80,
			),
		)

	mock.ExpectQuery("SELECT").
		WillReturnError(errors.New("database error"))

	_, err = repo.Semua(
		context.Background(),
		"",
		"",
	)

	if err == nil {
		t.Fatal("seharusnya menghasilkan error konversi")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoByIDBerhasil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	waktu := time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)

	mock.ExpectQuery("SELECT").
		WithArgs(int64(10)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"nama",
				"satuan_dasar_id",
				"satuan",
				"aktif",
				"created_at",
				"updated_at",
				"stok_total",
				"stok_layak",
			}).AddRow(
				10,
				"Paracetamol 500 mg",
				1,
				"Tablet",
				true,
				waktu,
				waktu,
				100,
				80,
			),
		)

	obat, err := repo.ByID(context.Background(), 10)
	if err != nil {
		t.Fatalf("gagal mengambil obat: %v", err)
	}

	if obat.ID != 10 {
		t.Errorf("ID = %d, ingin 10", obat.ID)
	}

	if obat.Nama != "Paracetamol 500 mg" {
		t.Errorf("nama = %q, tidak sesuai", obat.Nama)
	}

	if obat.StokTotal != 100 {
		t.Errorf("stok total = %d, ingin 100", obat.StokTotal)
	}

	if obat.StokLayak != 80 {
		t.Errorf("stok layak = %d, ingin 80", obat.StokLayak)
	}

	if obat.StokBlokir != 20 {
		t.Errorf("stok blokir = %d, ingin 20", obat.StokBlokir)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoByIDTidakDitemukan(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	mock.ExpectQuery("SELECT").
		WithArgs(int64(99)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id",
				"nama",
				"satuan_dasar_id",
				"satuan",
				"aktif",
				"created_at",
				"updated_at",
				"stok_total",
				"stok_layak",
			}),
		)

	_, err = repo.ByID(context.Background(), 99)

	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf(
			"seharusnya ErrNotFound, dapat: %v",
			err,
		)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoByIDDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	dbError := errors.New("database error")

	mock.ExpectQuery("SELECT").
		WithArgs(int64(10)).
		WillReturnError(dbError)

	_, err = repo.ByID(context.Background(), 10)

	if err == nil {
		t.Fatal("seharusnya menghasilkan database error")
	}

	if !errors.Is(err, dbError) {
		t.Errorf("error tidak sesuai: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoSatuanBerhasil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	mock.ExpectQuery("SELECT id, nama FROM satuan ORDER BY nama").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "nama"}).
				AddRow(1, "Botol").
				AddRow(2, "Kaplet").
				AddRow(3, "Tablet"),
		)

	daftar, err := repo.Satuan(context.Background())
	if err != nil {
		t.Fatalf("gagal mengambil satuan: %v", err)
	}

	if len(daftar) != 3 {
		t.Fatalf("jumlah satuan = %d, ingin 3", len(daftar))
	}

	if daftar[0].Nama != "Botol" {
		t.Errorf("nama satuan pertama = %q, ingin Botol", daftar[0].Nama)
	}

	if daftar[1].Nama != "Kaplet" {
		t.Errorf("nama satuan kedua = %q, ingin Kaplet", daftar[1].Nama)
	}

	if daftar[2].Nama != "Tablet" {
		t.Errorf("nama satuan ketiga = %q, ingin Tablet", daftar[2].Nama)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoSatuanKosong(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	mock.ExpectQuery("SELECT id, nama FROM satuan ORDER BY nama").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "nama"}),
		)

	daftar, err := repo.Satuan(context.Background())
	if err != nil {
		t.Fatalf("tidak seharusnya menghasilkan error: %v", err)
	}

	if len(daftar) != 0 {
		t.Errorf("jumlah satuan = %d, ingin 0", len(daftar))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoSatuanDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	mock.ExpectQuery("SELECT id, nama FROM satuan ORDER BY nama").
		WillReturnError(errors.New("database error"))

	_, err = repo.Satuan(context.Background())

	if err == nil {
		t.Fatal("seharusnya menghasilkan database error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoSatuanScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	mock.ExpectQuery("SELECT id, nama FROM satuan ORDER BY nama").
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "nama"}).
				AddRow("bukan-angka", "Tablet"),
		)

	_, err = repo.Satuan(context.Background())

	if err == nil {
		t.Fatal("seharusnya menghasilkan scan error")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoDetailBerhasil(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	waktu := time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)

	// Mengambil informasi utama obat.
	mock.ExpectQuery("SELECT").
		WithArgs(int64(10)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "nama", "satuan_dasar_id",
				"satuan", "aktif", "created_at",
				"updated_at", "stok_total", "stok_layak",
			}).AddRow(
				10, "Paracetamol", 1,
				"Tablet", true, waktu, waktu, 100, 80,
			),
		)

	// Mengambil konversi satuan.
	mock.ExpectQuery("SELECT").
		WithArgs(int64(10)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "obat_id", "satuan_id",
				"nama", "isi", "harga_jual",
			}).
				AddRow(1, 10, 1, "Tablet", 1, 2500).
				AddRow(2, 10, 2, "Strip", 10, 20000),
		)

	// Mengambil data batch.
	mock.ExpectQuery("SELECT").
		WithArgs(int64(10)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "no_batch", "expired_at",
				"harga_beli_dasar", "stok_dasar", "status",
			}).
				AddRow(1, "BATCH001", waktu, 1000.0, 50, "Layak jual").
				AddRow(2, "BATCH002", waktu, 1200.0, 30, "Layak jual"),
		)

	detail, err := repo.Detail(context.Background(), 10)
	if err != nil {
		t.Fatalf("gagal mengambil detail obat: %v", err)
	}

	if detail.Obat.ID != 10 {
		t.Errorf("ID obat = %d, ingin 10", detail.Obat.ID)
	}

	if detail.Obat.Nama != "Paracetamol" {
		t.Errorf("nama obat = %q, tidak sesuai", detail.Obat.Nama)
	}

	if detail.Obat.StokTotal != 100 {
		t.Errorf("stok total = %d, ingin 100", detail.Obat.StokTotal)
	}

	if len(detail.Satuan) != 2 {
		t.Errorf("jumlah satuan = %d, ingin 2", len(detail.Satuan))
	}

	if len(detail.Batch) != 2 {
		t.Errorf("jumlah batch = %d, ingin 2", len(detail.Batch))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoDetailTidakDitemukan(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}

	mock.ExpectQuery("SELECT").
		WithArgs(int64(99)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "nama", "satuan_dasar_id",
				"satuan", "aktif", "created_at",
				"updated_at", "stok_total", "stok_layak",
			}),
		)

	_, err = repo.Detail(context.Background(), 99)

	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("seharusnya ErrNotFound, dapat: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoDetailSatuanDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}
	waktu := time.Now()

	mock.ExpectQuery("SELECT").
		WithArgs(int64(10)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "nama", "satuan_dasar_id",
				"satuan", "aktif", "created_at",
				"updated_at", "stok_total", "stok_layak",
			}).AddRow(
				10, "Paracetamol", 1,
				"Tablet", true, waktu, waktu, 100, 80,
			),
		)

	mock.ExpectQuery("SELECT").
		WithArgs(int64(10)).
		WillReturnError(errors.New("database error"))

	_, err = repo.Detail(context.Background(), 10)

	if err == nil {
		t.Fatal("seharusnya menghasilkan error database satuan")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}

func TestObatRepoDetailBatchDatabaseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := &ObatRepo{DB: db}
	waktu := time.Now()

	mock.ExpectQuery("SELECT").
		WithArgs(int64(10)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "nama", "satuan_dasar_id",
				"satuan", "aktif", "created_at",
				"updated_at", "stok_total", "stok_layak",
			}).AddRow(
				10, "Paracetamol", 1,
				"Tablet", true, waktu, waktu, 100, 80,
			),
		)

	mock.ExpectQuery("SELECT").
		WithArgs(int64(10)).
		WillReturnRows(
			sqlmock.NewRows([]string{
				"id", "obat_id", "satuan_id",
				"nama", "isi", "harga_jual",
			}).AddRow(1, 10, 1, "Tablet", 1, 2500),
		)

	mock.ExpectQuery("SELECT").
		WithArgs(int64(10)).
		WillReturnError(errors.New("database error"))

	_, err = repo.Detail(context.Background(), 10)

	if err == nil {
		t.Fatal("seharusnya menghasilkan error database batch")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("ekspektasi SQL tidak terpenuhi: %v", err)
	}
}
