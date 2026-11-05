package main

import (
	"bytes"
	"embed"
	"errors"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static
var staticFS embed.FS

// Field limits follow the database schema (VARCHAR(255)).
const (
	maxTitleLen        = 255
	maxAnnouncementLen = 255
	maxFullTextLen     = 60000
)

// Server wires HTTP handlers to a Store.
type Server struct {
	store  Store
	tmpl   *template.Template
	logger *slog.Logger
}

// NewServer parses all templates once at startup instead of on every request.
func NewServer(store Store, logger *slog.Logger) (*Server, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{store: store, tmpl: tmpl, logger: logger}, nil
}

// Routes uses the method- and wildcard-aware ServeMux from Go 1.22+.
func (s *Server) Routes() http.Handler {
	static, _ := fs.Sub(staticFS, "static")

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleIndex)
	mux.HandleFunc("GET /create", s.handleCreateForm)
	mux.HandleFunc("POST /save_article", s.handleSave)
	mux.HandleFunc("GET /post/{id}", s.handleShow)
	mux.HandleFunc("GET /healthz", s.handleHealth)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(static)))
	return mux
}

type formData struct {
	Error   string
	Article Article
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	articles, err := s.store.List(r.Context())
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	s.render(w, r, http.StatusOK, "index", articles)
}

func (s *Server) handleCreateForm(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, http.StatusOK, "create", formData{})
}

func (s *Server) handleSave(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	a := Article{
		Title:        strings.TrimSpace(r.PostFormValue("title")),
		Announcement: strings.TrimSpace(r.PostFormValue("announcement")),
		FullText:     strings.TrimSpace(r.PostFormValue("full_text")),
	}
	if msg := validate(a); msg != "" {
		s.render(w, r, http.StatusBadRequest, "create", formData{Error: msg, Article: a})
		return
	}

	id, err := s.store.Create(r.Context(), a)
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/post/"+strconv.FormatUint(id, 10), http.StatusSeeOther)
}

func (s *Server) handleShow(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		http.NotFound(w, r)
		return
	}

	a, err := s.store.Get(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		s.serverError(w, r, err)
		return
	}
	s.render(w, r, http.StatusOK, "show", a)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok"))
}

func validate(a Article) string {
	switch {
	case a.Title == "" || a.Announcement == "" || a.FullText == "":
		return "All fields are required."
	case utf8.RuneCountInString(a.Title) > maxTitleLen:
		return "Title is too long (max 255 characters)."
	case utf8.RuneCountInString(a.Announcement) > maxAnnouncementLen:
		return "Announcement is too long (max 255 characters)."
	case utf8.RuneCountInString(a.FullText) > maxFullTextLen:
		return "Article text is too long."
	}
	return ""
}

// render executes the template into a buffer first, so a template error
// never results in a half-written page with a 200 status.
func (s *Server) render(w http.ResponseWriter, r *http.Request, status int, name string, data any) {
	var buf bytes.Buffer
	if err := s.tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		s.serverError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

func (s *Server) serverError(w http.ResponseWriter, r *http.Request, err error) {
	s.logger.Error("request failed", "method", r.Method, "path", r.URL.Path, "err", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
