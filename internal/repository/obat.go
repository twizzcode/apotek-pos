package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/twizzcode/apotek-pos/internal/model"
)

type ObatRepo struct {
	DB *sql.DB
}

// Semua mengambil seluruh data obat beserta ringkasan stok.

// Semua mengambil daftar obat berdasarkan pencarian dan status.
func (r *ObatRepo) Semua(
	ctx context.Context,
	pencarian string,
	status string,
) ([]model.Obat, error) {

	// Menyusun kondisi pencarian.
	kondisi := "WHERE o.deleted_at IS NULL"
	args := []any{}

	// Pencarian berdasarkan sebagian nama obat.
	pencarian = strings.TrimSpace(pencarian)

	if pencarian != "" {
		kondisi += " AND LOWER(o.nama) LIKE LOWER(?)"
		args = append(args, "%"+pencarian+"%")
	}

	// Filter berdasarkan status obat.
	switch status {
	case "aktif":
		kondisi += " AND o.aktif = ?"
		args = append(args, true)

	case "nonaktif":
		kondisi += " AND o.aktif = ?"
		args = append(args, false)
	}

	query := `
        SELECT
            o.id,
            o.nama,
            o.satuan_dasar_id,
            s.nama,
            o.aktif,
            o.created_at,
            o.updated_at,
            COALESCE(SUM(b.stok_dasar), 0),
            COALESCE(SUM(
                CASE
                    WHEN b.stok_dasar > 0
                    AND b.expired_at > DATE_ADD(CURDATE(), INTERVAL 30 DAY)
                    THEN b.stok_dasar
                    ELSE 0
                END
            ), 0)
        FROM obat o
        JOIN satuan s ON s.id = o.satuan_dasar_id
        LEFT JOIN batch b ON b.obat_id = o.id
        ` + kondisi + `
        GROUP BY o.id, o.nama, o.satuan_dasar_id,
            s.nama, o.aktif, o.created_at, o.updated_at
        ORDER BY o.nama
    `

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("daftar obat: %w", err)
	}

	var daftar []model.Obat

	for rows.Next() {
		var o model.Obat

		err := rows.Scan(
			&o.ID,
			&o.Nama,
			&o.SatuanDasarID,
			&o.SatuanDasar,
			&o.Aktif,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.StokTotal,
			&o.StokLayak,
		)
		if err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan obat: %w", err)
		}

		o.StokBlokir = o.StokTotal - o.StokLayak
		daftar = append(daftar, o)
	}

	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("baca daftar obat: %w", err)
	}

	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("tutup daftar obat: %w", err)
	}

	// Mengambil seluruh konversi dan harga jual obat.
	querySatuan := `
        SELECT
            os.obat_id,
            os.id,
            os.satuan_id,
            s.nama,
            os.isi,
            os.harga_jual
        FROM obat_satuan os
        JOIN satuan s ON s.id = os.satuan_id
        JOIN obat o ON o.id = os.obat_id
        WHERE o.deleted_at IS NULL
        ORDER BY os.obat_id, os.isi
    `

	satuanRows, err := r.DB.QueryContext(ctx, querySatuan)
	if err != nil {
		return nil, fmt.Errorf("daftar konversi obat: %w", err)
	}
	defer satuanRows.Close()

	// Memetakan ID obat ke posisi dalam daftar.
	indeks := make(map[int64]int)

	for i, o := range daftar {
		indeks[o.ID] = i
	}

	for satuanRows.Next() {
		var os model.ObatSatuan
		var obatID int64

		err := satuanRows.Scan(
			&obatID,
			&os.ID,
			&os.SatuanID,
			&os.Nama,
			&os.Isi,
			&os.HargaJual,
		)
		if err != nil {
			return nil, fmt.Errorf("scan konversi obat: %w", err)
		}

		if i, ok := indeks[obatID]; ok {
			daftar[i].Satuan = append(daftar[i].Satuan, os)
		}
	}

	if err := satuanRows.Err(); err != nil {
		return nil, fmt.Errorf("baca konversi obat: %w", err)
	}

	return daftar, nil
}


// Katalog mengambil obat aktif yang memiliki stok layak jual.
func (r *ObatRepo) Katalog(
    ctx context.Context,
) ([]model.Obat, error) {

    daftar, err := r.Semua(ctx, "", "aktif")
    if err != nil {
        return nil, err
    }

    var katalog []model.Obat

    for _, obat := range daftar {
        if obat.StokLayak <= 0 {
            continue
        }

        var satuanTersedia []model.ObatSatuan

        for _, satuan := range obat.Satuan {
            if satuan.Isi <= 0 {
                continue
            }

            satuan.StokTersedia = obat.StokLayak / satuan.Isi

            if satuan.StokTersedia > 0 {
                satuanTersedia = append(
                    satuanTersedia,
                    satuan,
                )
            }
        }

        obat.Satuan = satuanTersedia

        if len(obat.Satuan) > 0 {
            katalog = append(katalog, obat)
        }
    }

    return katalog, nil
}

// ByID mengambil informasi utama obat berdasarkan ID.
func (r *ObatRepo) ByID(ctx context.Context, id int64) (model.Obat, error) {
	var o model.Obat

	query := `
		SELECT
			o.id,
			o.nama,
			o.satuan_dasar_id,
			s.nama,
			o.aktif,
			o.created_at,
			o.updated_at,
			COALESCE(SUM(b.stok_dasar), 0),
			COALESCE(SUM(
				CASE
					WHEN b.expired_at > DATE_ADD(CURDATE(), INTERVAL 30 DAY)
					THEN b.stok_dasar
					ELSE 0
				END
			), 0)
		FROM obat o
		JOIN satuan s ON s.id = o.satuan_dasar_id
		LEFT JOIN batch b ON b.obat_id = o.id
		WHERE o.id = ? AND o.deleted_at IS NULL
		GROUP BY o.id, o.nama, o.satuan_dasar_id,
			s.nama, o.aktif, o.created_at, o.updated_at
	`

	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&o.ID,
		&o.Nama,
		&o.SatuanDasarID,
		&o.SatuanDasar,
		&o.Aktif,
		&o.CreatedAt,
		&o.UpdatedAt,
		&o.StokTotal,
		&o.StokLayak,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return o, model.ErrNotFound
	}
	if err != nil {
		return o, fmt.Errorf("cari obat: %w", err)
	}

	o.StokBlokir = o.StokTotal - o.StokLayak

	return o, nil
}

// Satuan mengambil seluruh pilihan satuan dari database.
func (r *ObatRepo) Satuan(ctx context.Context) ([]model.Satuan, error) {
	rows, err := r.DB.QueryContext(
		ctx,
		"SELECT id, nama FROM satuan ORDER BY nama",
	)
	if err != nil {
		return nil, fmt.Errorf("daftar satuan: %w", err)
	}
	defer rows.Close()

	var daftar []model.Satuan

	for rows.Next() {
		var s model.Satuan

		if err := rows.Scan(&s.ID, &s.Nama); err != nil {
			return nil, fmt.Errorf("scan satuan: %w", err)
		}

		daftar = append(daftar, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar satuan: %w", err)
	}

	return daftar, nil
}

// Detail mengambil satuan kemasan dan stok setiap batch.
func (r *ObatRepo) Detail(ctx context.Context, id int64) (model.DetailObat, error) {
	var detail model.DetailObat

	o, err := r.ByID(ctx, id)
	if err != nil {
		return detail, err
	}

	detail.Obat = o

	// Mengambil konversi dan harga jual.
	querySatuan := `
		SELECT
			os.id,
			os.obat_id,
			os.satuan_id,
			s.nama,
			os.isi,
			os.harga_jual
		FROM obat_satuan os
		JOIN satuan s ON s.id = os.satuan_id
		WHERE os.obat_id = ?
		ORDER BY os.isi
	`

	rows, err := r.DB.QueryContext(ctx, querySatuan, id)
	if err != nil {
		return detail, fmt.Errorf("detail satuan obat: %w", err)
	}

	for rows.Next() {
		var os model.ObatSatuan

		err := rows.Scan(
			&os.ID,
			&os.ObatID,
			&os.SatuanID,
			&os.Nama,
			&os.Isi,
			&os.HargaJual,
		)
		if err != nil {
			rows.Close()
			return detail, fmt.Errorf("scan satuan obat: %w", err)
		}

		detail.Satuan = append(detail.Satuan, os)
	}

	err = rows.Err()
	rows.Close()
	if err != nil {
		return detail, fmt.Errorf("baca satuan obat: %w", err)
	}

	// Mengambil stok setiap batch.
	queryBatch := `
		SELECT
			id,
			no_batch,
			expired_at,
			harga_beli_dasar,
			stok_dasar,
			CASE
				WHEN expired_at > DATE_ADD(CURDATE(), INTERVAL 30 DAY)
				THEN 'Layak jual'
				ELSE 'Blokir jual'
			END
		FROM batch
		WHERE obat_id = ?
		ORDER BY expired_at ASC
	`

	batches, err := r.DB.QueryContext(ctx, queryBatch, id)
	if err != nil {
		return detail, fmt.Errorf("daftar batch obat: %w", err)
	}
	defer batches.Close()

	for batches.Next() {
		var b model.BatchObat

		err := batches.Scan(
			&b.ID,
			&b.NoBatch,
			&b.ExpiredAt,
			&b.HargaBeliDasar,
			&b.Stok,
			&b.Status,
		)
		if err != nil {
			return detail, fmt.Errorf("scan batch obat: %w", err)
		}

		detail.Batch = append(detail.Batch, b)
	}

	if err := batches.Err(); err != nil {
		return detail, fmt.Errorf("baca batch obat: %w", err)
	}

	return detail, nil
}

// Buat menyimpan obat beserta konversi dan harga jualnya.
func (r *ObatRepo) Buat(ctx context.Context, obat model.Obat, satuan []model.ObatSatuan) error {
	if len(satuan) == 0 {
		return fmt.Errorf("minimal harus ada satu satuan obat")
	}

	// Pastikan satuan dasar memiliki konversi isi 1.
	adaSatuanDasar := false

	for _, s := range satuan {
		if s.SatuanID <= 0 || s.Isi <= 0 || s.HargaJual < 0 {
			return fmt.Errorf("data satuan atau harga tidak valid")
		}

		if s.SatuanID == obat.SatuanDasarID && s.Isi == 1 {
			adaSatuanDasar = true
		}
	}

	if !adaSatuanDasar {
		return fmt.Errorf("satuan dasar harus memiliki konversi isi 1")
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mulai transaksi obat: %w", err)
	}
	defer tx.Rollback()

	// Simpan data utama obat.
	result, err := tx.ExecContext(ctx, `
		INSERT INTO obat (nama, satuan_dasar_id, aktif)
		VALUES (?, ?, ?)
	`, obat.Nama, obat.SatuanDasarID, obat.Aktif)

	if err != nil {
		return fmt.Errorf("simpan obat: %w", err)
	}

	obatID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("ambil ID obat: %w", err)
	}

	// Simpan konversi dan harga jual.
	for _, s := range satuan {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO obat_satuan
				(obat_id, satuan_id, isi, harga_jual)
			VALUES (?, ?, ?, ?)
		`, obatID, s.SatuanID, s.Isi, s.HargaJual)

		if err != nil {
			return fmt.Errorf("simpan konversi obat: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit obat: %w", err)
	}

	return nil
}

// Update memperbarui data obat beserta konversi dan harga jualnya.
func (r *ObatRepo) Update(
	ctx context.Context,
	id int64,
	obat model.Obat,
	satuan []model.ObatSatuan) error {
	if len(satuan) == 0 {
		return fmt.Errorf("minimal harus ada satu satuan obat")
	}

	// Validasi satuan dasar dan data konversi.
	adaSatuanDasar := false
	satuanTerpakai := make(map[int64]bool)

	for _, s := range satuan {
		if s.SatuanID <= 0 || s.Isi <= 0 || s.HargaJual < 0 {
			return fmt.Errorf("data satuan atau harga tidak valid")
		}

		if satuanTerpakai[s.SatuanID] {
			return fmt.Errorf("satuan tidak boleh duplikat")
		}

		satuanTerpakai[s.SatuanID] = true

		if s.SatuanID == obat.SatuanDasarID && s.Isi == 1 {
			adaSatuanDasar = true
		}
	}

	if !adaSatuanDasar {
		return fmt.Errorf("satuan dasar harus memiliki konversi isi 1")
	}

	// Memulai transaksi database.
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mulai transaksi update obat: %w", err)
	}
	defer tx.Rollback()

	// Memastikan obat tersedia dan mengunci barisnya.
	var obatID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM obat
		WHERE id = ? AND deleted_at IS NULL
		FOR UPDATE
	`, id).Scan(&obatID)

	if errors.Is(err, sql.ErrNoRows) {
		return model.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("cari obat: %w", err)
	}

	// Memperbarui informasi utama obat.
	_, err = tx.ExecContext(ctx, `
		UPDATE obat
		SET nama = ?,
		    satuan_dasar_id = ?,
		    aktif = ?
		WHERE id = ?
	`, obat.Nama, obat.SatuanDasarID, obat.Aktif, id)

	if err != nil {
		return fmt.Errorf("update obat: %w", err)
	}

	// Memperbarui konversi yang sudah ada
	// atau menambahkan konversi baru.
	for _, s := range satuan {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO obat_satuan
				(obat_id, satuan_id, isi, harga_jual)
			VALUES (?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				isi = VALUES(isi),
				harga_jual = VALUES(harga_jual)
		`, id, s.SatuanID, s.Isi, s.HargaJual)

		if err != nil {
			return fmt.Errorf("update konversi obat: %w", err)
		}
	}

	// Menyimpan seluruh perubahan.
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit update obat: %w", err)
	}

	return nil
}

// UbahStatus mengaktifkan atau menonaktifkan obat.
func (r *ObatRepo) UbahStatus(
	ctx context.Context,
	id int64,
	aktif bool,
) error {
	result, err := r.DB.ExecContext(ctx, `
		UPDATE obat
		SET aktif = ?
		WHERE id = ? AND deleted_at IS NULL
	`, aktif, id)

	if err != nil {
		return fmt.Errorf("ubah status obat: %w", err)
	}

	jumlah, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("periksa status obat: %w", err)
	}

	if jumlah == 0 {
		var ada bool

		err := r.DB.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1
				FROM obat
				WHERE id = ? AND deleted_at IS NULL
			)
		`, id).Scan(&ada)

		if err != nil {
			return fmt.Errorf("periksa keberadaan obat: %w", err)
		}

		if !ada {
			return model.ErrNotFound
		}
	}

	return nil
}

// Hapus melakukan soft delete pada obat.
func (r *ObatRepo) Hapus(
	ctx context.Context,
	id int64,
) error {
	result, err := r.DB.ExecContext(ctx, `
        UPDATE obat
        SET deleted_at = CURRENT_TIMESTAMP
        WHERE id = ? AND deleted_at IS NULL
    `, id)

	if err != nil {
		return fmt.Errorf("hapus obat: %w", err)
	}

	jumlah, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("periksa penghapusan obat: %w", err)
	}

	if jumlah == 0 {
		return model.ErrNotFound
	}

	return nil
}
