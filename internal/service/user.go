package service

import (
	"context"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/twizzcode/apotek-pos/internal/model"
	"github.com/twizzcode/apotek-pos/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrLogin = errors.New("username atau password salah")

type ValidasiError struct{ Pesan string }

func (e *ValidasiError) Error() string { return e.Pesan }

type UserService struct{ Repo *repository.UserRepo }

func (s *UserService) Login(ctx context.Context, username, password string) (model.User, error) {
	u, err := s.Repo.ByUsername(ctx, strings.TrimSpace(username))
	if err != nil || !u.Aktif || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return model.User{}, ErrLogin
	}
	return u, nil
}

func (s *UserService) Daftar(ctx context.Context) ([]model.User, error) { return s.Repo.Semua(ctx) }

func (s *UserService) Buat(ctx context.Context, nama, username, password, role string) error {
	nama, username = strings.TrimSpace(nama), strings.TrimSpace(username)
	if nama == "" || username == "" || len(password) < 8 || !roleValid(role) {
		return &ValidasiError{"Nama dan username wajib diisi, password minimal 8 karakter, role harus valid"}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	err = s.Repo.Buat(ctx, model.User{Nama: nama, Username: username, PasswordHash: string(hash), Role: role})
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return &ValidasiError{"Username sudah digunakan"}
	}
	return err
}

func (s *UserService) SetAktif(ctx context.Context, id, actorID int64, aktif bool) error {
	if id == actorID && !aktif {
		return &ValidasiError{"Akun sendiri tidak bisa dinonaktifkan"}
	}
	return s.Repo.SetAktif(ctx, id, aktif)
}

func (s *UserService) ResetPassword(ctx context.Context, id int64, password string) error {
	if len(password) < 8 {
		return &ValidasiError{"Password minimal 8 karakter"}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.Repo.SetPassword(ctx, id, string(hash))
}

func roleValid(role string) bool { return role == "user" || role == "admin" || role == "apoteker" }
