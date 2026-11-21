package model

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("data tidak ditemukan")

type User struct {
	ID           int64
	Nama         string
	Username     string
	PasswordHash string
	Role         string
	Aktif        bool
	CreatedAt    time.Time
}

func (u User) BisaPOS() bool { return u.Role == "admin" || u.Role == "apoteker" }
