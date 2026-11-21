package handler

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/twizzcode/apotek-pos/internal/middleware"
	"github.com/twizzcode/apotek-pos/internal/model"
)

type PageData struct {
	Title string
	User  model.User
	Error string
	Pesan string
	Data  any
}

type Renderer struct{ pages map[string]*template.Template }

func NewRenderer() *Renderer {
	funcs := template.FuncMap{
		"rupiah":  rupiah,
		"tanggal": func(t time.Time) string { return t.Format("02 Jan 2006") },
	}
	files := []string{"login.html", "dashboard.html", "akses-ditolak.html", "users/list.html"}
	r := &Renderer{pages: make(map[string]*template.Template)}
	for _, file := range files {
		r.pages[file] = template.Must(template.New("layout").Funcs(funcs).ParseFiles(
			"templates/layout.html", "templates/header.html", "templates/sidebar.html", "templates/"+file,
		))
	}
	return r
}

func (rd *Renderer) Tampil(w http.ResponseWriter, r *http.Request, file string, data PageData) {
	if u, ok := middleware.User(r); ok {
		data.User = u
	}
	var buf bytes.Buffer
	if err := rd.pages[file].ExecuteTemplate(&buf, "layout", data); err != nil {
		log.Printf("render %s: %v", file, err)
		http.Error(w, "Terjadi kesalahan pada server", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}

func rupiah(n int64) string {
	s := strconv.FormatInt(n, 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return "Rp" + s
}
