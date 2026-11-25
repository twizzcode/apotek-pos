package model

import "time"

// Obat menyimpan informasi utama obat.
type Obat struct {
	ID            int64
	Nama          string
	SatuanDasarID int64
	SatuanDasar   string
	Aktif         bool
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Daftar satuan dan harga jual obat.
	Satuan []ObatSatuan

	// Ringkasan stok dalam satuan dasar.
	StokTotal    int64
	StokLayak    int64
	StokBlokir   int64
	StokTersedia int64
}

// Satuan menyimpan daftar satuan yang tersedia.
type Satuan struct {
	ID   int64
	Nama string
}

// ObatSatuan menyimpan konversi kemasan
// dan harga jual untuk setiap obat.
type ObatSatuan struct {
	ID        int64
	ObatID    int64
	SatuanID  int64
	Nama      string
	Isi       int64
	HargaJual int64
	StokTersedia int64
}

// BatchObat menyimpan informasi stok per batch.
type BatchObat struct {
	ID             int64
	NoBatch        string
	ExpiredAt      time.Time
	HargaBeliDasar float64
	Stok           int64
	Status         string
}

// DetailObat menyimpan informasi lengkap
// untuk halaman detail obat.
type DetailObat struct {
	Obat
	Batch []BatchObat
}
