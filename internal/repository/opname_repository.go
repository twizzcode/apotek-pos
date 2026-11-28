package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/twizzcode/apotek-pos/internal/model"
)

type OpnameRepository struct{ DB *sql.DB }

func (r *OpnameRepository) Daftar(ctx context.Context, f model.OpnameFilter) ([]model.Opname, error) {
	query := `SELECT o.id, o.kode, o.tanggal, u.nama,
		COUNT(i.id),
		COALESCE(SUM(CASE WHEN i.selisih > 0 THEN i.selisih END), 0),
		COALESCE(SUM(CASE WHEN i.selisih < 0 THEN -i.selisih END), 0)
		FROM opname o
		JOIN users u ON u.id = o.user_id
		LEFT JOIN opname_item i ON i.opname_id = o.id
		WHERE 1 = 1`
	args := []any{}
	if f.Dari != nil {
		query += " AND o.tanggal >= ?"
		args = append(args, *f.Dari)
	}
	if f.Sampai != nil {
		query += " AND o.tanggal <= ?"
		args = append(args, *f.Sampai)
	}
	if f.Kode != "" {
		query += " AND o.kode LIKE ?"
		args = append(args, "%"+f.Kode+"%")
	}
	query += " GROUP BY o.id, o.kode, o.tanggal, u.nama ORDER BY o.tanggal DESC, o.id DESC"

	rows, err := r.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("daftar opname: %w", err)
	}
	defer rows.Close()

	daftar := make([]model.Opname, 0)
	for rows.Next() {
		var o model.Opname
		if err := rows.Scan(&o.ID, &o.Kode, &o.Tanggal, &o.Petugas,
			&o.JumlahBatch, &o.SelisihPlus, &o.SelisihMinus); err != nil {
			return nil, fmt.Errorf("scan opname: %w", err)
		}
		daftar = append(daftar, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar opname: %w", err)
	}
	return daftar, nil
}

func (r *OpnameRepository) ByID(ctx context.Context, id int64) (model.Opname, []model.OpnameItem, error) {
	var o model.Opname
	var catatan sql.NullString
	err := r.DB.QueryRowContext(ctx,
		`SELECT o.id, o.kode, o.tanggal, o.user_id, u.nama, o.catatan, o.created_at
		 FROM opname o JOIN users u ON u.id = o.user_id WHERE o.id = ?`, id).
		Scan(&o.ID, &o.Kode, &o.Tanggal, &o.UserID, &o.Petugas, &catatan, &o.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return o, nil, model.ErrNotFound
	}
	if err != nil {
		return o, nil, fmt.Errorf("cari opname: %w", err)
	}
	o.Catatan = catatan.String

	rows, err := r.DB.QueryContext(ctx,
		`SELECT i.id, i.batch_id, ob.nama, b.no_batch, b.expired_at,
		        i.stok_sistem, i.stok_fisik, i.selisih, i.alasan
		 FROM opname_item i
		 JOIN batch b ON b.id = i.batch_id
		 JOIN obat ob ON ob.id = b.obat_id
		 WHERE i.opname_id = ?
		 ORDER BY ob.nama, b.expired_at, b.id`, id)
	if err != nil {
		return o, nil, fmt.Errorf("daftar item opname: %w", err)
	}
	defer rows.Close()

	items := make([]model.OpnameItem, 0)
	for rows.Next() {
		var it model.OpnameItem
		var alasan sql.NullString
		if err := rows.Scan(&it.ID, &it.BatchID, &it.NamaObat, &it.NoBatch, &it.ExpiredAt,
			&it.StokSistem, &it.StokFisik, &it.Selisih, &alasan); err != nil {
			return o, nil, fmt.Errorf("scan item opname: %w", err)
		}
		it.Alasan = alasan.String
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return o, nil, fmt.Errorf("baca item opname: %w", err)
	}
	return o, items, nil
}

func (r *OpnameRepository) BatchUntukOpname(ctx context.Context) ([]model.BatchOpname, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT b.id, o.nama, b.no_batch, b.expired_at, b.stok_dasar, s.nama
		 FROM batch b
		 JOIN obat o ON o.id = b.obat_id
		 JOIN satuan s ON s.id = o.satuan_dasar_id
		 WHERE o.deleted_at IS NULL
		 ORDER BY o.nama, b.expired_at, b.id`)
	if err != nil {
		return nil, fmt.Errorf("daftar batch opname: %w", err)
	}
	defer rows.Close()

	hasil := make([]model.BatchOpname, 0)
	for rows.Next() {
		var b model.BatchOpname
		if err := rows.Scan(&b.BatchID, &b.NamaObat, &b.NoBatch, &b.ExpiredAt, &b.StokSistem, &b.Satuan); err != nil {
			return nil, fmt.Errorf("scan batch opname: %w", err)
		}
		hasil = append(hasil, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca batch opname: %w", err)
	}
	return hasil, nil
}

func (r *OpnameRepository) Buat(ctx context.Context, in model.OpnameInput) (int64, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("mulai transaksi opname: %w", err)
	}
	defer tx.Rollback()

	prefix := "OPN-" + in.Tanggal.Format("060102") + "-"
	var urut int
	if err := tx.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM opname WHERE kode LIKE ?", prefix+"%").Scan(&urut); err != nil {
		return 0, fmt.Errorf("hitung kode opname: %w", err)
	}
	kode := fmt.Sprintf("%s%03d", prefix, urut+1)

	res, err := tx.ExecContext(ctx,
		"INSERT INTO opname (kode, tanggal, user_id, catatan) VALUES (?, ?, ?, ?)",
		kode, in.Tanggal, in.UserID, sql.NullString{String: in.Catatan, Valid: in.Catatan != ""})
	if err != nil {
		return 0, fmt.Errorf("simpan opname: %w", err)
	}
	opnameID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("id opname: %w", err)
	}

	for _, it := range in.Items {
		var sistem int
		var noBatch string
		err := tx.QueryRowContext(ctx,
			"SELECT stok_dasar, no_batch FROM batch WHERE id = ? FOR UPDATE", it.BatchID).
			Scan(&sistem, &noBatch)
		if errors.Is(err, sql.ErrNoRows) {
			return 0, model.ErrNotFound
		}
		if err != nil {
			return 0, fmt.Errorf("kunci batch: %w", err)
		}

		selisih := it.StokFisik - sistem
		if selisih != 0 && it.Alasan == "" {
			return 0, fmt.Errorf("batch %s: %w", noBatch, model.ErrAlasanWajib)
		}

		if _, err := tx.ExecContext(ctx,
			`INSERT INTO opname_item (opname_id, batch_id, stok_sistem, stok_fisik, selisih, alasan)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			opnameID, it.BatchID, sistem, it.StokFisik, selisih,
			sql.NullString{String: it.Alasan, Valid: it.Alasan != ""}); err != nil {
			return 0, fmt.Errorf("simpan item opname: %w", err)
		}
		if selisih != 0 {
			if _, err := tx.ExecContext(ctx,
				"UPDATE batch SET stok_dasar = ? WHERE id = ?", it.StokFisik, it.BatchID); err != nil {
				return 0, fmt.Errorf("sesuaikan stok batch: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit opname: %w", err)
	}
	return opnameID, nil
}
