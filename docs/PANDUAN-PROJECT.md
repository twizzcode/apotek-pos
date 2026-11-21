# Panduan Global Project POS Apotek

Dokumen ini menjadi pedoman utama seluruh anggota kelompok. Baca sebelum mulai mengerjakan issue. Jika isi issue berbeda dengan dokumen ini, diskusikan dengan Yafi sebelum mengubah kode atau database.

## 1. Informasi Project

| Item | Nilai |
|---|---|
| Nama aplikasi | POS Apotek Sumber Waras |
| Repository | <https://github.com/twizzcode/apotek-pos> |
| Bahasa | Go |
| Database | MySQL 8.0 melalui Docker Compose |
| Frontend | HTML, CSS, dan JavaScript native |
| Server | `net/http` dan `html/template` |
| Arsitektur | `handler -> service -> repository -> model` |
| Branch utama | `main` |
| Reviewer dan merger | Yafi |

### Tujuan aplikasi

Aplikasi mengelola obat berdasarkan batch agar apotek dapat:

1. Mencatat obat, satuan, dan konversi kemasan.
2. Mencatat penerimaan barang dari PBF beserta batch dan tanggal kedaluwarsa.
3. Menjual obat dengan pemilihan batch otomatis menggunakan FEFO.
4. Menerima retur dan mengetahui batch asal barang.
5. Mencatat stock opname beserta selisih stok.
6. Menampilkan stok per batch, daftar kedaluwarsa, dan laba kotor.

## 2. Menjalankan Project

### Syarat

- Go 1.24 atau lebih baru.
- Docker dan Docker Compose.
- Git.

### Setup pertama kali

```bash
git clone https://github.com/twizzcode/apotek-pos.git
cd apotek-pos
cp .env.example .env
docker compose up -d --wait
go mod download
go run ./cmd/web
```

Buka <http://localhost:8080>.

### Akun demo

| Role | Username | Password | Akses |
|---|---|---|---|
| Admin | `admin` | `password` | Semua menu dan manajemen pengguna |
| Apoteker | `apoteker` | `password` | Operasional apotek dan POS |
| User | `user` | `password` | Halaman tidak ada akses dan logout |

### Perintah yang sering dipakai

```bash
make db-up       # Menyalakan MySQL
make db-down     # Menghentikan MySQL
make db-reset    # Menghapus data lalu menjalankan schema dan seed ulang
make run         # Menjalankan web
make test        # Menjalankan seluruh test
```

MySQL project memakai port host `3308`. Jangan menggantinya menjadi `3306` karena port tersebut dipakai project lain.

## 3. Struktur Folder

```text
pos-apotek/
├── cmd/web/main.go          # Wiring dependency, route, dan server
├── db/
│   ├── schema.sql           # Struktur database bersama
│   └── seed.sql             # Data dan akun demo
├── internal/
│   ├── config/              # Environment dan konfigurasi
│   ├── handler/             # HTTP request, form, response, redirect
│   ├── middleware/          # Auth, role, logging, security header
│   ├── model/               # Struct data dan error domain
│   ├── repository/          # Query SQL dan transaksi database
│   └── service/             # Validasi dan aturan bisnis
├── templates/               # Template HTML per modul
├── static/
│   ├── css/style.css        # Seluruh tampilan native CSS
│   └── js/app.js            # Interaksi frontend native
├── docs/                    # Dokumentasi project
├── compose.yaml
├── Makefile
└── README.md
```

## 4. Arsitektur dan Tanggung Jawab Layer

Alur request wajib mengikuti pola berikut:

```text
Browser -> Handler -> Service -> Repository -> MySQL
Browser <- Template <- Handler <- Service <- Repository
```

### Handler

- Membaca path, query string, dan form HTTP.
- Memanggil service.
- Memilih status HTTP, redirect, atau template.
- Tidak menulis query SQL.
- Tidak menjadi tempat aturan bisnis.

### Service

- Memvalidasi input.
- Menjalankan aturan bisnis.
- Mengatur operasi yang perlu transaction melalui repository.
- Mengembalikan error yang dapat diterjemahkan handler.
- Menjadi tempat test untuk logika non-trivial.

### Repository

- Menjalankan query SQL.
- Melakukan scan hasil query ke model.
- Menjalankan DB transaction dan locking bila diperlukan.
- Membungkus error dengan konteks menggunakan `%w`.
- Tidak merender HTML dan tidak membaca form HTTP.

### Model

- Berisi struct data, konstanta domain, dan error umum.
- Tidak bergantung pada HTTP atau template.

### Template

- Menampilkan data yang diberikan handler.
- Tidak menentukan aturan harga, stok, FEFO, atau laba.
- JavaScript hanya membantu interaksi; server tetap sumber kebenaran.

## 5. Hak Akses Role

| Fitur | Admin | Apoteker | User |
|---|---:|---:|---:|
| Dashboard | Ya | Ya | Tidak |
| Kelola pengguna | Ya | Tidak | Tidak |
| Lihat obat dan stok | Ya | Ya | Tidak |
| Ubah master obat/satuan | Ya | Tidak | Tidak |
| Kelola PBF | Ya | Lihat | Tidak |
| Penerimaan barang | Ya | Ya | Tidak |
| Kasir dan penjualan | Ya | Ya | Tidak |
| Retur penjualan | Ya | Ya | Tidak |
| Stock opname | Ya | Ya | Tidak |
| Laporan stok dan ED | Ya | Ya | Tidak |
| Laporan laba kotor | Ya | Tidak | Tidak |

Proteksi wajib dilakukan di route menggunakan `RequireLogin` atau `RequireRole`. Menyembunyikan menu saja tidak cukup.

## 6. Aturan Bisnis Bersama

### 6.1 Satuan dan konversi

- Stok selalu disimpan dalam satuan dasar.
- `obat_satuan.isi` berarti jumlah satuan dasar dalam satu satuan jual.
- Setiap obat wajib memiliki satuan dasar dengan `isi = 1`.
- Contoh paracetamol:
  - 1 tablet = 1 tablet dasar.
  - 1 strip = 10 tablet dasar.
  - 1 box = 100 tablet dasar.
- Harga jual disimpan untuk setiap satuan jual dan tidak harus proporsional.

### 6.2 Penerimaan dan batch

- Satu faktur memiliki banyak item.
- Satu item penerimaan menghasilkan satu batch.
- Nomor batch unik untuk obat yang sama.
- `stok_awal_dasar = qty penerimaan * isi satuan`.
- `harga_beli_dasar = harga beli per satuan penerimaan / isi satuan`.
- Semua item, batch, stok, dan total faktur disimpan dalam satu DB transaction.
- Jika satu item gagal, seluruh penerimaan harus rollback.

### 6.3 Penjualan dan FEFO

- FEFO berarti batch dengan tanggal kedaluwarsa terdekat dijual lebih dulu.
- Urutan batch: `expired_at ASC, id ASC`.
- Batch dengan sisa masa kedaluwarsa 30 hari atau kurang tidak boleh dijual.
- Batch terblokir tetap tercatat dan tetap muncul di laporan.
- Satu item penjualan boleh mengambil stok dari beberapa batch.
- Harga jual disalin ke `penjualan_item` saat transaksi.
- HPP dasar disalin ke `penjualan_item_batch` saat transaksi.
- Perubahan harga setelah transaksi tidak boleh mengubah riwayat.
- Checkout memakai DB transaction dan `SELECT ... FOR UPDATE`.
- Kekurangan stok membatalkan seluruh transaksi.
- Lock item dilakukan dalam urutan konsisten untuk mengurangi deadlock dua kasir.

### 6.4 Pembayaran

- Pembayaran hanya tunai.
- Uang bayar tidak boleh kurang dari total.
- Kembalian dihitung server: `bayar - total`.
- Nilai total dari browser tidak boleh langsung dipercaya; server menghitung ulang.

### 6.5 Retur

- Retur harus merujuk ke transaksi dan alokasi batch asal.
- Jumlah retur tidak boleh melebihi jumlah terjual dikurangi retur sebelumnya.
- Kondisi baik menambah stok batch asal.
- Kondisi rusak tetap tercatat tetapi tidak menambah stok.
- Restock dan penyimpanan retur dilakukan dalam satu DB transaction.
- Refund dihitung dari snapshot harga jual transaksi.

### 6.6 Stock opname

- Opname dilakukan per batch, bukan hanya per obat.
- Simpan stok sistem sebelum koreksi.
- Simpan stok fisik hasil hitung.
- `selisih = stok fisik - stok sistem`.
- Selisih bukan nol wajib memiliki alasan.
- Setelah final, stok batch diset ke stok fisik.
- Riwayat selisih tidak boleh dihapus atau ditimpa.

### 6.7 Laporan laba

- HPP berasal dari batch yang benar-benar terjual.
- HPP transaksi dihitung dari `SUM(qty_dasar * harga_beli_dasar)` pada alokasi batch.
- Laba kotor dihitung dengan `omzet bersih - HPP bersih`.
- Retur mengurangi omzet dan HPP sesuai batch yang diretur.
- Perhitungan memakai snapshot transaksi, bukan harga master terbaru.

### 6.8 Kedaluwarsa

| Kondisi | Status |
|---|---|
| `expired_at < CURDATE()` | Sudah kedaluwarsa |
| Sisa 0–30 hari | Diblokir dari penjualan |
| Sisa 31–90 hari | Peringatan |
| Lebih dari 90 hari | Normal |

## 7. Kontrak Database

Diagram lengkap tersedia di [`ERD.md`](ERD.md). Tabel utama:

| Tabel | Fungsi |
|---|---|
| `users` | Akun dan role |
| `satuan` | Master satuan |
| `obat` | Master obat dan satuan dasar |
| `obat_satuan` | Konversi dan harga jual |
| `pbf` | Data pemasok |
| `penerimaan` | Header faktur penerimaan |
| `penerimaan_item` | Item dalam faktur |
| `batch` | Nomor batch, ED, HPP, dan stok dasar |
| `penjualan` | Header transaksi kasir |
| `penjualan_item` | Obat, satuan, qty, dan snapshot harga jual |
| `penjualan_item_batch` | Alokasi stok serta snapshot HPP per batch |
| `retur` | Header retur |
| `retur_item` | Batch asal, kondisi, qty, dan refund |
| `opname` | Header stock opname |
| `opname_item` | Stok sistem, fisik, selisih, dan alasan per batch |

### Aturan perubahan schema

1. Jangan mengubah `db/schema.sql` langsung tanpa diskusi kelompok.
2. Tulis kebutuhan perubahan pada issue atau grup.
3. Periksa dampak ke model, repository, seed, ERD, dan anggota lain.
4. Yafi menyetujui perubahan sebelum branch di-merge.
5. Setelah schema berubah, jalankan:

```bash
make db-reset
go test ./...
```

Perintah `make db-reset` menghapus seluruh data lokal. Jangan jalankan jika ada data lokal penting yang belum disalin.

## 8. Route dan Pemilik Backend

| Method | Route | Halaman/Fungsi | Role | Pemilik |
|---|---|---|---|---|
| GET/POST | `/login` | Login | Publik | Yafi |
| POST | `/logout` | Logout | Login | Yafi |
| GET | `/` | Dashboard | Admin, Apoteker | Yafi |
| GET/POST | `/users...` | Manajemen pengguna | Admin | Yafi |
| GET | `/obat` | Daftar obat | Admin, Apoteker | Selma |
| GET | `/obat/tambah` | Form tambah obat | Admin | Selma |
| GET | `/obat/{id}` | Detail obat dan batch | Admin, Apoteker | Selma |
| GET | `/obat/{id}/edit` | Form edit obat | Admin | Selma |
| GET | `/pbf` | PBF | Admin, Apoteker | Ezra |
| GET | `/penerimaan` | Riwayat penerimaan | Admin, Apoteker | Ezra |
| GET | `/penerimaan/tambah` | Form penerimaan | Admin, Apoteker | Ezra |
| GET | `/penerimaan/{id}` | Detail faktur | Admin, Apoteker | Ezra |
| GET | `/kasir` | Kasir dan keranjang | Admin, Apoteker | Raihan |
| GET | `/penjualan` | Riwayat penjualan | Admin, Apoteker | Raihan |
| GET | `/penjualan/{id}` | Detail transaksi | Admin, Apoteker | Raihan |
| GET | `/penjualan/{id}/struk` | Struk | Admin, Apoteker | Raihan |
| GET | `/retur` | Riwayat retur | Admin, Apoteker | Maratova |
| GET | `/retur/tambah` | Form retur | Admin, Apoteker | Maratova |
| GET | `/retur/{id}` | Detail retur | Admin, Apoteker | Maratova |
| GET | `/opname` | Riwayat opname | Admin, Apoteker | Maratova |
| GET | `/opname/tambah` | Form opname | Admin, Apoteker | Maratova |
| GET | `/opname/{id}` | Detail opname | Admin, Apoteker | Maratova |
| GET | `/laporan/stok` | Stok per batch | Admin, Apoteker | Nabila |
| GET | `/laporan/kedaluwarsa` | Kedaluwarsa | Admin, Apoteker | Nabila |
| GET | `/laporan/laba` | Laba kotor | Admin | Nabila |

Route POST untuk simpan, ubah, hapus, checkout, retur, dan finalisasi opname ditambahkan pemilik modul dengan pola URL yang konsisten. Aksi yang mengubah data tidak boleh memakai method GET.

## 9. Kontrak Frontend dan Backend

Frontend seluruh modul sudah disediakan Yafi. Selama backend belum terhubung, route menggunakan `PreviewHandler` dan data contoh.

### Yang dilakukan pemilik backend

1. Pertahankan URL, template, class CSS, dan alur halaman.
2. Ganti route `PreviewHandler` modulnya dengan handler asli.
3. Buat struct view data yang jelas untuk template.
4. Ganti data contoh dengan `{{range .Data...}}` atau field yang sesuai.
5. Isi `method`, `action`, dan `name` form sesuai handler.
6. Tampilkan error validasi tanpa kehilangan data form pengguna.
7. Redirect setelah POST berhasil agar refresh tidak mengulang transaksi.
8. Minta Yafi mereview jika perlu mengubah layout atau JavaScript bersama.

### Kepemilikan template

| Folder | Backend |
|---|---|
| `templates/obat/` | Selma |
| `templates/pbf/` | Ezra |
| `templates/penerimaan/` | Ezra |
| `templates/kasir/` | Raihan |
| `templates/penjualan/` | Raihan |
| `templates/retur/` | Maratova |
| `templates/opname/` | Maratova |
| `templates/laporan/` | Nabila |

`static/css/style.css`, `static/js/app.js`, layout, header, sidebar, dan tampilan lintas modul dimiliki Yafi. Anggota boleh mengusulkan perubahan, tetapi jangan mengganti desain sepihak.

### Batas JavaScript

- Boleh: tambah/hapus baris form, keranjang visual, format nominal, konfirmasi, dan preview perhitungan.
- Tidak boleh menjadi satu-satunya validasi: stok, harga, role, FEFO, refund, laba, atau finalisasi transaksi.
- Semua aturan penting harus dihitung dan divalidasi ulang oleh server.

## 10. Pembagian Tugas

| Anggota | Tanggung jawab | Issue |
|---|---|---|
| Yafi | Fondasi, Docker, schema, auth, role, layout, dashboard, akun, seluruh frontend, integrasi akhir | #1–#5, #21 |
| Selma | Obat, satuan, konversi, harga jual, katalog stok layak jual | #6–#8 |
| Ezra | PBF, faktur penerimaan multi-item, batch, riwayat penerimaan | #9–#11 |
| Raihan | Kasir, keranjang, checkout FEFO, concurrency, pembayaran, struk, riwayat penjualan | #12–#14 |
| Maratova | Retur ke batch asal, stock opname, riwayat retur dan opname | #15–#17 |
| Nabila | Laporan stok batch, kedaluwarsa, dan laba kotor | #18–#20 |

Daftar issue: <https://github.com/twizzcode/apotek-pos/issues>

### Dependency pengerjaan

```text
Yafi: fondasi + frontend
        |
        +-- Selma: obat dan konversi --+
        |                              |
        +-- Ezra: PBF dan batch -------+--> Raihan: penjualan
                                               |
                                               +--> Maratova: retur
                                               |
                                               +--> Nabila: laba

Ezra: batch --> Maratova: opname
Ezra: batch --> Nabila: stok dan ED
```

- Selma dan Ezra dapat bekerja paralel.
- Raihan mengintegrasikan katalog Selma dan batch Ezra.
- Maratova membutuhkan alokasi batch penjualan dari Raihan untuk retur.
- Nabila dapat mengerjakan stok/ED setelah Ezra; laporan laba menunggu Raihan dan Maratova.

## 11. Standar Kode

### Penamaan

- Nama package: bahasa Inggris/Indonesia pendek, huruf kecil, tanpa underscore.
- Nama file: huruf kecil, misalnya `penerimaan.go` dan `penjualan_test.go`.
- Nama tipe/fungsi publik: PascalCase.
- Nama variabel/fungsi privat: camelCase.
- Gunakan istilah domain yang sudah ada: `obat`, `satuan`, `batch`, `penerimaan`, `penjualan`, `retur`, dan `opname`.
- Jangan membuat istilah kedua untuk konsep sama, misalnya memakai `produk` untuk `obat`.

### Query dan error

- Gunakan placeholder `?`; jangan menyusun input pengguna ke string SQL.
- Gunakan `QueryContext`, `QueryRowContext`, dan `ExecContext`.
- Selalu `defer rows.Close()` setelah query berhasil.
- Selalu periksa `rows.Err()`.
- Terjemahkan `sql.ErrNoRows` menjadi `model.ErrNotFound` jika sesuai.
- Bungkus error: `fmt.Errorf("simpan penerimaan: %w", err)`.
- Jangan tampilkan error SQL mentah ke pengguna.

### Uang dan jumlah

- Nominal rupiah bulat memakai integer (`BIGINT UNSIGNED` / `int64`).
- HPP per satuan dasar boleh pecahan dan memakai `DECIMAL(15,4)` di DB.
- Jangan gunakan `float64` untuk total uang jika pembulatan dapat memengaruhi hasil.
- Jumlah stok dasar tidak boleh negatif.

### Validasi trust boundary

Semua input dari form, URL, query string, dan cookie dianggap tidak tepercaya. Minimal validasi:

- ID harus angka positif.
- String wajib di-trim dan diperiksa kosong.
- Qty harus lebih dari nol.
- Harga tidak boleh negatif.
- Foreign key harus ada dan aktif bila diperlukan.
- Role harus diperiksa server.
- Tanggal harus masuk akal terhadap alur bisnis.

### Transaksi database

Gunakan pola aman:

1. `BeginTx`.
2. Lock dan baca data yang akan diubah.
3. Validasi ulang setelah lock.
4. Insert/update seluruh data terkait.
5. `Commit` hanya jika semua berhasil.
6. `Rollback` untuk semua jalur error.

## 12. Aturan Git dan Pull Request

### Larangan

- Jangan mengerjakan langsung di `main`.
- Jangan push langsung ke `main`.
- Jangan merge Pull Request sendiri.
- Jangan memakai `git push --force` pada branch bersama.
- Jangan menghapus atau menimpa pekerjaan anggota lain.
- Yafi melakukan review dan merge ke `main`.

### Mulai satu issue

Gunakan satu branch untuk satu issue:

```bash
git checkout main
git pull origin main
git checkout -b fitur/<nomor-issue>-<nama-singkat>
```

Contoh:

```bash
git checkout -b fitur/6-crud-obat
```

### Commit

Format pesan:

```text
feat: tambah penerimaan multi-item
fix: cegah stok batch menjadi negatif
test: tambah pengujian alokasi FEFO
docs: perbarui panduan integrasi
```

Commit perubahan yang satu tujuan. Jangan menggabungkan perubahan seluruh modul dalam satu commit besar.

### Sebelum push

```bash
gofmt -w cmd internal
go test ./...
git diff --check
git status
```

Pastikan `.env`, binary, file log, dan data lokal tidak ikut di-commit.

### Push dan Pull Request

```bash
git push -u origin fitur/<nomor-issue>-<nama-singkat>
```

Isi Pull Request:

```markdown
Closes #<nomor-issue>

## Perubahan
- ...

## Cara menguji
1. ...

## Screenshot
- Lampirkan jika ada perubahan tampilan.

## Catatan integrasi
- Perubahan route/schema/kontrak yang perlu diketahui anggota lain.
```

Yafi mereview. Perbaiki komentar pada branch sama, lalu push ulang. Jangan membuat PR baru untuk perbaikan PR yang sama.

### Mengambil pembaruan `main`

Sebelum melanjutkan pekerjaan lama:

```bash
git checkout main
git pull origin main
git checkout fitur/<branch-kamu>
git merge main
```

Jika ada konflik, jangan asal memilih `ours` atau `theirs`. Pahami kedua perubahan, pertahankan fungsi keduanya, lalu jalankan seluruh test.

## 13. Standar Pengujian

Gunakan package `testing` bawaan Go. Tidak perlu menambah framework test.

### Wajib diuji

- Konversi satuan menjadi stok dasar dan HPP dasar.
- FEFO satu batch dan lintas beberapa batch.
- Batch ED 30 hari atau kurang ditolak dari penjualan.
- Stok tidak cukup menyebabkan rollback.
- Dua checkout tidak dapat menjual stok yang sama.
- Retur sebagian dan retur berulang.
- Retur baik menambah batch asal; retur rusak tidak.
- Selisih opname dihitung benar dan tercatat.
- Laba memakai HPP batch yang dialokasikan.
- Harga master berubah tetapi transaksi lama tetap sama.

### Uji manual minimum

1. Login sesuai role.
2. Buka route modul.
3. Uji data valid.
4. Uji input kosong/tidak valid.
5. Uji data tidak ditemukan.
6. Uji akses role yang dilarang.
7. Periksa redirect dan pesan hasil.
8. Periksa tampilan desktop dan layar sempit.

## 14. Checklist Selesai Issue

- [ ] Scope issue selesai tanpa fitur tambahan yang tidak diminta.
- [ ] Struktur mengikuti `handler -> service -> repository -> model`.
- [ ] Route dilindungi role yang tepat.
- [ ] Semua input divalidasi server.
- [ ] Query aman dari SQL injection.
- [ ] Operasi multi-tabel memakai transaction.
- [ ] Error tidak membocorkan detail database.
- [ ] Template frontend yang tersedia sudah dihubungkan.
- [ ] Data contoh pada modul sudah diganti data backend.
- [ ] Test logika penting tersedia dan lulus.
- [ ] `gofmt -w cmd internal` sudah dijalankan.
- [ ] `go test ./...` lulus.
- [ ] `git diff --check` bersih.
- [ ] Pull Request berisi `Closes #...` dan cara uji.
- [ ] Tidak ada `.env`, password baru, log, atau binary di commit.

## 15. Checklist Integrasi Akhir

- [ ] Schema dan seed dapat berjalan dari volume kosong.
- [ ] Tiga akun demo dapat login sesuai hak akses.
- [ ] Master obat, satuan, harga, dan PBF berfungsi.
- [ ] Penerimaan membuat batch dan menambah stok dasar.
- [ ] Kasir memilih batch FEFO dan menolak batch ED ≤ 30 hari.
- [ ] Dua kasir tidak menyebabkan stok negatif.
- [ ] Struk menampilkan snapshot transaksi.
- [ ] Retur terlacak ke batch asal.
- [ ] Opname mencatat selisih sebelum koreksi stok.
- [ ] Laporan stok dan ED sesuai data batch.
- [ ] Laporan laba sesuai batch yang terjual dan retur.
- [ ] Semua menu sesuai role.
- [ ] Semua halaman tidak lagi memakai data contoh.
- [ ] `go test ./...` lulus pada branch `main`.
- [ ] Alur demo presentasi sudah dicoba dari awal sampai akhir.

## 16. Alur Demo Presentasi

Urutan demo yang disarankan:

1. Login admin dan tunjukkan manajemen pengguna.
2. Buat obat serta konversi tablet, strip, dan box.
3. Buat PBF dan penerimaan dua batch dengan ED serta harga beli berbeda.
4. Tunjukkan stok per batch dan peringatan ED.
5. Login apoteker dan lakukan penjualan.
6. Tunjukkan sistem mengambil batch FEFO secara otomatis.
7. Tunjukkan penolakan ketika stok layak jual tidak cukup.
8. Cetak struk.
9. Retur sebagian kondisi baik dan tunjukkan stok kembali ke batch asal.
10. Lakukan stock opname dan tunjukkan riwayat selisih.
11. Buka laporan stok, kedaluwarsa, dan laba kotor.
12. Login sebagai `user` dan tunjukkan bahwa POS tidak dapat diakses.

## 17. Sumber Kebenaran

Urutan acuan jika menemukan perbedaan:

1. Keputusan kelompok terbaru yang dicatat pada issue/PR.
2. Dokumen global ini.
3. Acceptance criteria pada issue.
4. ERD dan schema database.
5. Prototype frontend.

Perubahan keputusan wajib diperbarui pada dokumen atau issue terkait agar semua anggota memakai aturan yang sama.
