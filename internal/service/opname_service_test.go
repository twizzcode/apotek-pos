package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/twizzcode/apotek-pos/internal/model"
)

type opnameRepoFake struct {
	input   model.OpnameInput
	called  bool
	buatErr error
}

func (f *opnameRepoFake) Daftar(context.Context, model.OpnameFilter) ([]model.Opname, error) {
	return nil, nil
}
func (f *opnameRepoFake) ByID(context.Context, int64) (model.Opname, []model.OpnameItem, error) {
	return model.Opname{}, nil, nil
}
func (f *opnameRepoFake) BatchUntukOpname(context.Context) ([]model.BatchOpname, error) {
	return nil, nil
}
func (f *opnameRepoFake) Buat(_ context.Context, in model.OpnameInput) (int64, error) {
	f.called = true
	f.input = in
	return 1, f.buatErr
}

func TestOpnameBuatMenolakInputTidakValid(t *testing.T) {
	kasus := map[string][]model.OpnameItemInput{
		"kosong":  {},
		"negatif": {{BatchID: 1, StokFisik: -1}},
		"ganda":   {{BatchID: 1, StokFisik: 1}, {BatchID: 1, StokFisik: 2}},
	}
	for nama, items := range kasus {
		repo := &opnameRepoFake{}
		_, err := (&OpnameService{Repo: repo}).Buat(context.Background(), 2, "2026-11-28", "", items)
		if _, ok := err.(*ValidasiError); !ok {
			t.Errorf("%s: error = %v, ingin ValidasiError", nama, err)
		}
		if repo.called {
			t.Errorf("%s: repository tidak boleh dipanggil", nama)
		}
	}
}

func TestOpnameBuatMenolakTanggalTidakValid(t *testing.T) {
	repo := &opnameRepoFake{}
	_, err := (&OpnameService{Repo: repo}).Buat(context.Background(), 2, "bukan-tanggal", "",
		[]model.OpnameItemInput{{BatchID: 1, StokFisik: 1}})
	if _, ok := err.(*ValidasiError); !ok || repo.called {
		t.Fatalf("err = %v, called = %v", err, repo.called)
	}
}

func TestOpnameBuatMengurutkanBatchDanMembersihkanAlasan(t *testing.T) {
	repo := &opnameRepoFake{}
	_, err := (&OpnameService{Repo: repo}).Buat(context.Background(), 2, "2026-11-28", "  akhir bulan ",
		[]model.OpnameItemInput{{BatchID: 3, StokFisik: 1, Alasan: "  rusak "}, {BatchID: 1, StokFisik: 5}})
	if err != nil {
		t.Fatalf("Buat() error = %v", err)
	}
	got := repo.input
	if got.Items[0].BatchID != 1 || got.Items[1].BatchID != 3 {
		t.Errorf("item tidak terurut: %+v", got.Items)
	}
	if got.Items[1].Alasan != "rusak" || got.Catatan != "akhir bulan" {
		t.Errorf("teks tidak dirapikan: %+v", got)
	}
}

func TestOpnameBuatMemetakanAlasanWajib(t *testing.T) {
	repo := &opnameRepoFake{buatErr: fmt.Errorf("batch PCT-A1: %w", model.ErrAlasanWajib)}
	_, err := (&OpnameService{Repo: repo}).Buat(context.Background(), 2, "2026-11-28", "",
		[]model.OpnameItemInput{{BatchID: 1, StokFisik: 90}})
	if _, ok := err.(*ValidasiError); !ok {
		t.Fatalf("error = %v, ingin ValidasiError", err)
	}
}
