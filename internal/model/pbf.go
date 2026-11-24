package model

import "time"

type PBF struct {
	ID        int64
	Nama      string
	Telepon   string
	Alamat    string
	Aktif     bool
	CreatedAt time.Time
}
