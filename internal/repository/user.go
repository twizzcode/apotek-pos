package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/twizzcode/apotek-pos/internal/model"
)

type UserRepo struct{ DB *sql.DB }

const kolomUser = "id, nama, username, password_hash, role, aktif, created_at"

func (r *UserRepo) ByUsername(ctx context.Context, username string) (model.User, error) {
	return r.satu(ctx, "SELECT "+kolomUser+" FROM users WHERE username = ?", username)
}

func (r *UserRepo) ByID(ctx context.Context, id int64) (model.User, error) {
	return r.satu(ctx, "SELECT "+kolomUser+" FROM users WHERE id = ?", id)
}

func (r *UserRepo) satu(ctx context.Context, query string, arg any) (model.User, error) {
	var u model.User
	err := r.DB.QueryRowContext(ctx, query, arg).Scan(&u.ID, &u.Nama, &u.Username, &u.PasswordHash, &u.Role, &u.Aktif, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return u, model.ErrNotFound
	}
	if err != nil {
		return u, fmt.Errorf("cari user: %w", err)
	}
	return u, nil
}

func (r *UserRepo) Semua(ctx context.Context) ([]model.User, error) {
	rows, err := r.DB.QueryContext(ctx, "SELECT "+kolomUser+" FROM users ORDER BY nama")
	if err != nil {
		return nil, fmt.Errorf("daftar user: %w", err)
	}
	defer rows.Close()
	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Nama, &u.Username, &u.PasswordHash, &u.Role, &u.Aktif, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepo) Buat(ctx context.Context, u model.User) error {
	_, err := r.DB.ExecContext(ctx,
		"INSERT INTO users (nama, username, password_hash, role) VALUES (?, ?, ?, ?)",
		u.Nama, u.Username, u.PasswordHash, u.Role)
	if err != nil {
		return fmt.Errorf("buat user: %w", err)
	}
	return nil
}

func (r *UserRepo) SetAktif(ctx context.Context, id int64, aktif bool) error {
	return r.ubah(ctx, "UPDATE users SET aktif = ? WHERE id = ?", aktif, id)
}

func (r *UserRepo) SetPassword(ctx context.Context, id int64, hash string) error {
	return r.ubah(ctx, "UPDATE users SET password_hash = ? WHERE id = ?", hash, id)
}

func (r *UserRepo) ubah(ctx context.Context, query string, args ...any) error {
	result, err := r.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("ubah user: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return model.ErrNotFound
	}
	return nil
}
