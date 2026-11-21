# Pembagian Tugas

| Anggota | Modul utama | Issue |
|---|---|---|
| Yafi | Fondasi, database, autentikasi, layout, dashboard, frontend bersama | `anggota:yafi` |
| Selma | Master obat, satuan, konversi, harga jual | `anggota:selma` |
| Ezra | PBF dan penerimaan barang per batch | `anggota:ezra` |
| Raihan | Kasir, FEFO, transaksi bersamaan, struk | `anggota:raihan` |
| Maratova | Retur penjualan dan stock opname | `anggota:maratova` |
| Nabila | Laporan stok, kedaluwarsa, dan laba kotor | `anggota:nabila` |

## Aturan kerja Git

1. Ambil satu issue, lalu buat branch `fitur/<nomor>-<nama-singkat>`.
2. Jangan mengubah `db/schema.sql` tanpa diskusi kelompok karena semua modul bergantung padanya.
3. Tambahkan route di `cmd/web/main.go`; pertahankan alur `handler -> service -> repository`.
4. Validasi input dilakukan di service. Query SQL hanya di repository.
5. Non-trivial logic wajib punya satu test kecil dengan package `testing` bawaan Go.
6. Pull request mencantumkan `Closes #<nomor>` dan langkah uji manual.
7. Frontend seluruh modul sudah disediakan Yafi. Hubungkan backend ke template sesuai [`frontend-contract.md`](frontend-contract.md); jangan membuat desain pengganti tanpa koordinasi.

## Urutan pengerjaan

1. Fondasi Yafi selesai lebih dulu.
2. Selma dan Ezra dapat berjalan paralel.
3. Raihan mulai setelah data obat, konversi, dan batch tersedia.
4. Maratova mulai setelah detail transaksi dan alokasi batch tersedia.
5. Nabila dapat mulai query laporan setelah schema stabil; validasi akhir menunggu transaksi contoh.
