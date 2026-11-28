package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/twizzcode/apotek-pos/internal/model"
)

type PBFRepository struct{ DB *sql.DB }

const kolomPBF = "id, nama, telepon, alamat, aktif, created_at"

func (r *PBFRepository) Semua(ctx context.Context) ([]model.PBF, error) {
	return r.daftar(ctx, "SELECT "+kolomPBF+" FROM pbf ORDER BY nama")
}

func (r *PBFRepository) Aktif(ctx context.Context) ([]model.PBF, error) {
	return r.daftar(ctx, "SELECT "+kolomPBF+" FROM pbf WHERE aktif = TRUE ORDER BY nama")
}

func (r *PBFRepository) daftar(ctx context.Context, query string) ([]model.PBF, error) {
	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("daftar PBF: %w", err)
	}
	defer rows.Close()

	pbfs := make([]model.PBF, 0)
	for rows.Next() {
		pbf, err := scanPBF(rows)
		if err != nil {
			return nil, fmt.Errorf("scan PBF: %w", err)
		}
		pbfs = append(pbfs, pbf)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar PBF: %w", err)
	}
	return pbfs, nil
}

func (r *PBFRepository) ByID(ctx context.Context, id int64) (model.PBF, error) {
	pbf, err := scanPBF(r.DB.QueryRowContext(ctx, "SELECT "+kolomPBF+" FROM pbf WHERE id = ?", id))
	if errors.Is(err, sql.ErrNoRows) {
		return pbf, model.ErrNotFound
	}
	if err != nil {
		return pbf, fmt.Errorf("cari PBF: %w", err)
	}
	return pbf, nil
}

func scanPBF(row interface{ Scan(...any) error }) (model.PBF, error) {
	var pbf model.PBF
	var telepon, alamat sql.NullString
	err := row.Scan(&pbf.ID, &pbf.Nama, &telepon, &alamat, &pbf.Aktif, &pbf.CreatedAt)
	if err != nil {
		return pbf, err
	}
	if telepon.Valid {
		pbf.Telepon = telepon.String
	}
	if alamat.Valid {
		pbf.Alamat = alamat.String
	}
	return pbf, nil
}

func (r *PBFRepository) Buat(ctx context.Context, pbf model.PBF) error {
	_, err := r.DB.ExecContext(ctx,
		"INSERT INTO pbf (nama, telepon, alamat) VALUES (?, ?, ?)",
		pbf.Nama, nullableString(pbf.Telepon), nullableString(pbf.Alamat))
	if err != nil {
		return fmt.Errorf("buat PBF: %w", err)
	}
	return nil
}

func (r *PBFRepository) Ubah(ctx context.Context, pbf model.PBF) error {
	result, err := r.DB.ExecContext(ctx,
		"UPDATE pbf SET nama = ?, telepon = ?, alamat = ? WHERE id = ?",
		pbf.Nama, nullableString(pbf.Telepon), nullableString(pbf.Alamat), pbf.ID)
	if err != nil {
		return fmt.Errorf("ubah PBF: %w", err)
	}
	return r.pastikanAda(ctx, result, pbf.ID)
}

func (r *PBFRepository) SetAktif(ctx context.Context, id int64, aktif bool) error {
	result, err := r.DB.ExecContext(ctx, "UPDATE pbf SET aktif = ? WHERE id = ?", aktif, id)
	if err != nil {
		return fmt.Errorf("ubah status PBF: %w", err)
	}
	return r.pastikanAda(ctx, result, id)
}

func (r *PBFRepository) pastikanAda(ctx context.Context, result sql.Result, id int64) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("hitung perubahan PBF: %w", err)
	}
	if affected > 0 {
		return nil
	}

	var exists bool
	if err := r.DB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM pbf WHERE id = ?)", id).Scan(&exists); err != nil {
		return fmt.Errorf("periksa PBF: %w", err)
	}
	if !exists {
		return model.ErrNotFound
	}
	return nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
