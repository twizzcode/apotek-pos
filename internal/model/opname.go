package model

import (
	"errors"
	"time"
)

var ErrAlasanWajib = errors.New("alasan wajib diisi jika ada selisih")

type Opname struct {
	ID           int64
	Kode         string
	Tanggal      time.Time
	UserID       int64
	Petugas      string
	Catatan      string
	CreatedAt    time.Time
	JumlahBatch  int
	SelisihPlus  int
	SelisihMinus int // nilai absolut
}

type OpnameItem struct {
	ID         int64
	BatchID    int64
	NamaObat   string
	NoBatch    string
	ExpiredAt  time.Time
	StokSistem int
	StokFisik  int
	Selisih    int
	Alasan     string
}

type BatchOpname struct {
	BatchID    int64
	NamaObat   string
	NoBatch    string
	ExpiredAt  time.Time
	StokSistem int
	Satuan     string
}

type OpnameItemInput struct {
	BatchID   int64
	StokFisik int
	Alasan    string
}

type OpnameInput struct {
	Tanggal time.Time
	UserID  int64
	Catatan string
	Items   []OpnameItemInput
}

type OpnameFilter struct {
	Dari   *time.Time
	Sampai *time.Time
	Kode   string
}
