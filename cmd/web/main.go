package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/twizzcode/apotek-pos/internal/config"
	"github.com/twizzcode/apotek-pos/internal/handler"
	"github.com/twizzcode/apotek-pos/internal/middleware"
	"github.com/twizzcode/apotek-pos/internal/repository"
	"github.com/twizzcode/apotek-pos/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := bukaDB(cfg.DBDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userRepo := &repository.UserRepo{DB: db}
	userSvc := &service.UserService{Repo: userRepo}
	auth := &middleware.Auth{Repo: userRepo, Secret: []byte(cfg.SessionSecret), Secure: cfg.CookieSecure}
	view := handler.NewRenderer()
	authHandler := &handler.AuthHandler{Svc: userSvc, Auth: auth, View: view}
	dashboard := &handler.DashboardHandler{Repo: &repository.DashboardRepo{DB: db}, View: view}
	users := &handler.UserHandler{Svc: userSvc, View: view}
	pbfSvc := &service.PBFService{Repo: &repository.PBFRepository{DB: db}}
	pbfs := &handler.PBFHandler{Svc: pbfSvc, View: view}
	preview := &handler.PreviewHandler{View: view}

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("GET /login", authHandler.FormLogin)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /logout", authHandler.Logout)
	mux.Handle("GET /{$}", auth.RequireRole("admin", "apoteker")(http.HandlerFunc(dashboard.Home)))
	mux.Handle("GET /akses-ditolak", auth.RequireLogin(http.HandlerFunc(dashboard.AksesDitolak)))
	mux.Handle("GET /users", auth.RequireRole("admin")(http.HandlerFunc(users.Daftar)))
	mux.Handle("POST /users", auth.RequireRole("admin")(http.HandlerFunc(users.Buat)))
	mux.Handle("POST /users/{id}/status", auth.RequireRole("admin")(http.HandlerFunc(users.Status)))
	mux.Handle("POST /users/{id}/password", auth.RequireRole("admin")(http.HandlerFunc(users.ResetPassword)))

	pos := auth.RequireRole("admin", "apoteker")
	admin := auth.RequireRole("admin")
	mux.Handle("GET /obat", pos(preview.Page("obat/list.html", "Obat & Satuan")))
	mux.Handle("GET /obat/tambah", admin(preview.Page("obat/form.html", "Tambah Obat")))
	mux.Handle("GET /obat/{id}", pos(preview.Page("obat/detail.html", "Detail Obat")))
	mux.Handle("GET /obat/{id}/edit", admin(preview.Page("obat/form.html", "Edit Obat")))
	mux.Handle("GET /pbf", pos(http.HandlerFunc(pbfs.Daftar)))
	mux.Handle("GET /pbf/{id}/edit", admin(http.HandlerFunc(pbfs.Edit)))
	mux.Handle("POST /pbf", admin(http.HandlerFunc(pbfs.Buat)))
	mux.Handle("POST /pbf/{id}", admin(http.HandlerFunc(pbfs.Ubah)))
	mux.Handle("POST /pbf/{id}/status", admin(http.HandlerFunc(pbfs.Status)))
	mux.Handle("GET /penerimaan", pos(preview.Page("penerimaan/list.html", "Penerimaan")))
	mux.Handle("GET /penerimaan/tambah", pos(preview.Page("penerimaan/form.html", "Penerimaan Baru")))
	mux.Handle("GET /penerimaan/{id}", pos(preview.Page("penerimaan/detail.html", "Detail Penerimaan")))
	mux.Handle("GET /kasir", pos(preview.Page("kasir/index.html", "Kasir")))
	mux.Handle("GET /penjualan", pos(preview.Page("penjualan/list.html", "Riwayat Penjualan")))
	mux.Handle("GET /penjualan/{id}", pos(preview.Page("penjualan/detail.html", "Detail Penjualan")))
	mux.Handle("GET /penjualan/{id}/struk", pos(preview.Page("penjualan/struk.html", "Struk Penjualan")))
	mux.Handle("GET /retur", pos(preview.Page("retur/list.html", "Retur Penjualan")))
	mux.Handle("GET /retur/tambah", pos(preview.Page("retur/form.html", "Retur Baru")))
	mux.Handle("GET /retur/{id}", pos(preview.Page("retur/detail.html", "Detail Retur")))
	mux.Handle("GET /opname", pos(preview.Page("opname/list.html", "Stock Opname")))
	mux.Handle("GET /opname/tambah", pos(preview.Page("opname/form.html", "Opname Baru")))
	mux.Handle("GET /opname/{id}", pos(preview.Page("opname/detail.html", "Detail Opname")))
	mux.Handle("GET /laporan/stok", pos(preview.Page("laporan/stok.html", "Laporan Stok")))
	mux.Handle("GET /laporan/kedaluwarsa", pos(preview.Page("laporan/kedaluwarsa.html", "Laporan Kedaluwarsa")))
	mux.Handle("GET /laporan/laba", admin(preview.Page("laporan/laba.html", "Laporan Laba Kotor")))

	root := middleware.Logging(middleware.SecurityHeaders(auth.LoadUser(mux)))
	srv := &http.Server{
		Addr: cfg.Addr, Handler: root, ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 10 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second,
	}
	log.Printf("POS Apotek jalan di http://localhost%s", cfg.Addr)
	log.Fatal(srv.ListenAndServe())
}

func bukaDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	var pingErr error
	for attempt := 1; attempt <= 10; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		pingErr = db.PingContext(ctx)
		cancel()
		if pingErr == nil {
			return db, nil
		}
		log.Printf("menunggu database (%d/10): %v", attempt, pingErr)
		time.Sleep(2 * time.Second)
	}
	db.Close()
	return nil, pingErr
}
