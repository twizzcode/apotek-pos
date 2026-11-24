package service

import (
	"context"
	"testing"

	"github.com/twizzcode/apotek-pos/internal/model"
)

type pbfRepoFake struct {
	created      model.PBF
	createCalled bool
	statuses     []bool
}

func (f *pbfRepoFake) Semua(context.Context) ([]model.PBF, error) { return nil, nil }
func (f *pbfRepoFake) Aktif(context.Context) ([]model.PBF, error) { return nil, nil }
func (f *pbfRepoFake) ByID(context.Context, int64) (model.PBF, error) {
	return model.PBF{}, nil
}
func (f *pbfRepoFake) Buat(_ context.Context, pbf model.PBF) error {
	f.created = pbf
	f.createCalled = true
	return nil
}
func (f *pbfRepoFake) Ubah(context.Context, model.PBF) error { return nil }
func (f *pbfRepoFake) SetAktif(_ context.Context, _ int64, aktif bool) error {
	f.statuses = append(f.statuses, aktif)
	return nil
}

func TestPBFServiceBuatMenolakNamaKosong(t *testing.T) {
	for _, nama := range []string{"", "   \t"} {
		repo := &pbfRepoFake{}
		err := (&PBFService{Repo: repo}).Buat(context.Background(), nama, "", "")
		if _, ok := err.(*ValidasiError); !ok {
			t.Errorf("Buat(%q) error = %v, ingin ValidasiError", nama, err)
		}
		if repo.createCalled {
			t.Errorf("Buat(%q) tidak boleh memanggil repository", nama)
		}
	}
}

func TestPBFServiceBuatNamaDenganKontakKosong(t *testing.T) {
	repo := &pbfRepoFake{}
	if err := (&PBFService{Repo: repo}).Buat(context.Background(), "  PBF Sehat  ", "", ""); err != nil {
		t.Fatalf("Buat() error = %v", err)
	}
	if !repo.createCalled || repo.created.Nama != "PBF Sehat" || repo.created.Telepon != "" || repo.created.Alamat != "" || !repo.created.Aktif {
		t.Fatalf("data PBF yang dibuat tidak sesuai: %+v", repo.created)
	}
}

func TestPBFServiceSetAktif(t *testing.T) {
	repo := &pbfRepoFake{}
	svc := &PBFService{Repo: repo}
	for _, aktif := range []bool{false, true} {
		if err := svc.SetAktif(context.Background(), 1, aktif); err != nil {
			t.Fatalf("SetAktif(%t) error = %v", aktif, err)
		}
	}
	if len(repo.statuses) != 2 || repo.statuses[0] || !repo.statuses[1] {
		t.Fatalf("status yang diteruskan tidak sesuai: %v", repo.statuses)
	}
}
