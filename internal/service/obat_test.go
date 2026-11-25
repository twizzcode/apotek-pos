package service

import (
	"context"
	"testing"

	"github.com/twizzcode/apotek-pos/internal/model"
)

func TestObatServiceBuatValidasiKonversi(t *testing.T) {
	svc := &ObatService{}

	tests := []struct {
		name          string
		nama          string
		satuanDasarID int64
		satuan        []model.ObatSatuan
		wantError     string
	}{
		{
			name:          "nama obat kosong",
			nama:          "",
			satuanDasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 1, HargaJual: 1000},
			},
			wantError: "Nama obat wajib diisi",
		},
		{
			name:          "satuan dasar tidak valid",
			nama:          "Paracetamol",
			satuanDasarID: 0,
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 1, HargaJual: 1000},
			},
			wantError: "Satuan dasar wajib dipilih",
		},
		{
			name:          "satuan konversi kosong",
			nama:          "Paracetamol",
			satuanDasarID: 1,
			satuan:        nil,
			wantError:     "Minimal harus ada satu satuan obat",
		},
		{
			name:          "ID satuan tidak valid",
			nama:          "Paracetamol",
			satuanDasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 0, Isi: 1, HargaJual: 1000},
			},
			wantError: "Data konversi atau harga jual tidak valid",
		},
		{
			name:          "isi konversi nol",
			nama:          "Paracetamol",
			satuanDasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 0, HargaJual: 1000},
			},
			wantError: "Data konversi atau harga jual tidak valid",
		},
		{
			name:          "harga jual negatif",
			nama:          "Paracetamol",
			satuanDasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 1, HargaJual: -1000},
			},
			wantError: "Data konversi atau harga jual tidak valid",
		},
		{
			name:          "satuan duplikat",
			nama:          "Paracetamol",
			satuanDasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 1, HargaJual: 1000},
				{SatuanID: 1, Isi: 10, HargaJual: 10000},
			},
			wantError: "Satuan tidak boleh dipilih lebih dari sekali",
		},
		{
			name:          "satuan dasar tidak memiliki isi 1",
			nama:          "Paracetamol",
			satuanDasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 2, Isi: 10, HargaJual: 10000},
			},
			wantError: "Satuan dasar harus memiliki konversi isi 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Buat(
				context.Background(),
				tt.nama,
				tt.satuanDasarID,
				true,
				tt.satuan,
			)

			if err == nil {
				t.Fatal("seharusnya menghasilkan error validasi")
			}

			if got := err.Error(); got != tt.wantError {
				t.Errorf(
					"error = %q, ingin %q",
					got,
					tt.wantError,
				)
			}
		})
	}
}

func TestObatServiceUpdateValidasiKonversi(t *testing.T) {
	svc := &ObatService{}

	tests := []struct {
		name      string
		id        int64
		nama      string
		dasarID   int64
		satuan    []model.ObatSatuan
		wantError string
	}{
		{
			name:    "ID obat tidak valid",
			id:      0,
			nama:    "Paracetamol",
			dasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 1, HargaJual: 1000},
			},
			wantError: "Data obat wajib diisi dengan benar",
		},
		{
			name:    "nama obat kosong",
			id:      1,
			nama:    "",
			dasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 1, HargaJual: 1000},
			},
			wantError: "Data obat wajib diisi dengan benar",
		},
		{
			name:    "satuan dasar tidak valid",
			id:      1,
			nama:    "Paracetamol",
			dasarID: 0,
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 1, HargaJual: 1000},
			},
			wantError: "Data obat wajib diisi dengan benar",
		},
		{
			name:      "konversi kosong",
			id:        1,
			nama:      "Paracetamol",
			dasarID:   1,
			satuan:    nil,
			wantError: "Minimal harus ada satu satuan obat",
		},
		{
			name:    "ID satuan tidak valid",
			id:      1,
			nama:    "Paracetamol",
			dasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 0, Isi: 1, HargaJual: 1000},
			},
			wantError: "Data konversi atau harga jual tidak valid",
		},
		{
			name:    "isi konversi nol",
			id:      1,
			nama:    "Paracetamol",
			dasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 0, HargaJual: 1000},
			},
			wantError: "Data konversi atau harga jual tidak valid",
		},
		{
			name:    "harga jual negatif",
			id:      1,
			nama:    "Paracetamol",
			dasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 1, HargaJual: -1000},
			},
			wantError: "Data konversi atau harga jual tidak valid",
		},
		{
			name:    "satuan duplikat",
			id:      1,
			nama:    "Paracetamol",
			dasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 1, HargaJual: 1000},
				{SatuanID: 1, Isi: 10, HargaJual: 10000},
			},
			wantError: "Satuan tidak boleh dipilih lebih dari sekali",
		},
		{
			name:    "satuan dasar tidak memiliki isi 1",
			id:      1,
			nama:    "Paracetamol",
			dasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 2, Isi: 10, HargaJual: 10000},
			},
			wantError: "Satuan dasar harus memiliki konversi isi 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Update(
				context.Background(),
				tt.id,
				tt.nama,
				tt.dasarID,
				true,
				tt.satuan,
			)

			if err == nil {
				t.Fatal("seharusnya menghasilkan error validasi")
			}

			if got := err.Error(); got != tt.wantError {
				t.Errorf(
					"error = %q, ingin %q",
					got,
					tt.wantError,
				)
			}
		})
	}
}

func TestValidasiKonversiBerhasil(t *testing.T) {
	tests := []struct {
		name          string
		satuanDasarID int64
		satuan        []model.ObatSatuan
	}{
		{
			name:          "satu satuan dasar",
			satuanDasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 1, HargaJual: 2500},
			},
		},
		{
			name:          "satuan dasar dan kemasan",
			satuanDasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 1, HargaJual: 2500},
				{SatuanID: 2, Isi: 10, HargaJual: 12000},
				{SatuanID: 3, Isi: 100, HargaJual: 70000},
			},
		},
		{
			name:          "harga jual nol diperbolehkan",
			satuanDasarID: 1,
			satuan: []model.ObatSatuan{
				{SatuanID: 1, Isi: 1, HargaJual: 0},
				{SatuanID: 2, Isi: 10, HargaJual: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validasiKonversi(
				tt.satuanDasarID,
				tt.satuan,
			)

			if err != nil {
				t.Fatalf("seharusnya valid, tetapi mendapat error: %v", err)
			}
		})
	}
}
