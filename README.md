# POS Apotek Sumber Waras

Sistem stok apotek berbasis batch: penerimaan PBF, penjualan FEFO, retur ke batch asal, stock opname, dan laporan laba kotor.

## Teknologi

- Go: `net/http`, `html/template`, `database/sql`
- MySQL 8.0 melalui Docker Compose
- HTML, CSS, dan JavaScript native
- Arsitektur `handler -> service -> repository -> model`

## Menjalankan aplikasi

Syarat: Go 1.24+, Docker, dan Docker Compose.

```bash
cp .env.example .env
docker compose up -d
go mod download
go run ./cmd/web
```

Buka <http://localhost:8080>.

| Role | Username | Password | Akses |
|---|---|---|---|
| Admin | `admin` | `password` | Semua menu dan kelola pengguna |
| Apoteker | `apoteker` | `password` | Operasional POS |
| User | `user` | `password` | Halaman tidak ada akses dan logout |

> Akun seed hanya untuk demo. Ganti password sebelum penggunaan nyata.

## Database

MySQL memakai port host `3308`; port `3306` tetap bebas untuk proyek lain.

```bash
make db-up       # hidupkan database
make db-down     # hentikan database
make db-reset    # hapus volume lalu jalankan schema dan seed ulang
```

Stok selalu disimpan sebagai satuan dasar. Contoh: satu box paracetamol berisi 10 strip, satu strip berisi 10 tablet, maka satu box bernilai 100 tablet.

## Aturan bisnis

- Batch layak jual dipilih dengan FEFO: kedaluwarsa terdekat keluar dulu.
- Batch dengan masa kedaluwarsa 30 hari atau kurang tidak boleh dijual.
- Transaksi penjualan mengunci batch memakai transaksi DB dan `SELECT ... FOR UPDATE`.
- Harga jual dan HPP disalin saat transaksi agar perubahan harga tidak mengubah riwayat.
- Retur kondisi baik kembali ke batch asal; kondisi rusak tidak menambah stok.
- Stock opname mencatat stok sistem, stok fisik, selisih, dan alasan per batch.
- Daftar mendekati kedaluwarsa menampilkan batch dengan sisa waktu maksimal 90 hari.

## Pengujian

```bash
go test ./...
```

Pembagian kerja: [`docs/pembagian-tugas.md`](docs/pembagian-tugas.md). Diagram database: [`docs/ERD.md`](docs/ERD.md).
