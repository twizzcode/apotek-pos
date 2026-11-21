# Kontrak Frontend

Semua halaman sudah tersedia sebagai prototype di branch `fitur/21-frontend-semua-modul`. Data pada halaman fitur masih contoh sampai handler modul terhubung.

## Tanggung jawab anggota fitur

1. Pertahankan template, URL, class CSS, dan alur UI yang sudah ada.
2. Ganti `PreviewHandler` pada route milik modul dengan handler asli.
3. Kirim data backend melalui `PageData.Data`; ubah data contoh menjadi ekspresi template Go.
4. Hubungkan form ke method dan action backend. Server wajib tetap memvalidasi semua input.
5. Jika membutuhkan perubahan UI, koordinasikan dengan Yafi agar tampilan antar modul konsisten.

## Kepemilikan template

| Modul | Template | Backend |
|---|---|---|
| Obat & satuan | `templates/obat/` | Selma |
| PBF | `templates/pbf/` | Ezra |
| Penerimaan | `templates/penerimaan/` | Ezra |
| Kasir | `templates/kasir/` | Raihan |
| Penjualan & struk | `templates/penjualan/` | Raihan |
| Retur | `templates/retur/` | Maratova |
| Stock opname | `templates/opname/` | Maratova |
| Laporan | `templates/laporan/` | Nabila |

`internal/handler/preview.go` hanya jembatan sementara. Hapus route preview per modul setelah handler asli mengambil alih.
