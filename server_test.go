package main

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
)

// memStore is an in-memory Store used to test handlers without MySQL.
type memStore struct {
	mu       sync.Mutex
	articles []Article
	failWith error
}

func (m *memStore) List(context.Context) ([]Article, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failWith != nil {
		return nil, m.failWith
	}
	return append([]Article(nil), m.articles...), nil
}

func (m *memStore) Get(_ context.Context, id uint64) (Article, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.articles {
		if a.ID == id {
			return a, nil
		}
	}
	return Article{}, ErrNotFound
}

func (m *memStore) Create(_ context.Context, a Article) (uint64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a.ID = uint64(len(m.articles) + 1)
	m.articles = append(m.articles, a)
	return a.ID, nil
}

func newTestServer(t *testing.T, store Store) http.Handler {
	t.Helper()
	srv, err := NewServer(store, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}
	return srv.Routes()
}

func do(t *testing.T, h http.Handler, req *http.Request) (*httptest.ResponseRecorder, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec, rec.Body.String()
}

func postForm(path string, form url.Values) *http.Request {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func TestIndexEmpty(t *testing.T) {
	h := newTestServer(t, &memStore{})
	rec, body := do(t, h, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(body, "There are no articles yet.") {
		t.Errorf("empty-state message missing from page")
	}
}

func TestIndexListsArticles(t *testing.T) {
	store := &memStore{articles: []Article{{ID: 1, Title: "First", Announcement: "Hello", FullText: "Body"}}}
	h := newTestServer(t, store)
	_, body := do(t, h, httptest.NewRequest(http.MethodGet, "/", nil))
	if !strings.Contains(body, "First") || !strings.Contains(body, `href="/post/1"`) {
		t.Errorf("article not rendered in index page")
	}
}

func TestIndexStoreError(t *testing.T) {
	h := newTestServer(t, &memStore{failWith: errors.New("db down")})
	rec, body := do(t, h, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if strings.Contains(body, "db down") {
		t.Errorf("internal error details leaked to the client")
	}
}

func TestShow(t *testing.T) {
	store := &memStore{articles: []Article{{ID: 1, Title: "Title", Announcement: "Ann", FullText: "Text"}}}
	h := newTestServer(t, store)

	tests := []struct {
		path string
		code int
	}{
		{"/post/1", http.StatusOK},
		{"/post/2", http.StatusNotFound},
		{"/post/0", http.StatusNotFound},
		{"/post/abc", http.StatusNotFound},
		{"/post/1%27%20OR%20%271%27=%271", http.StatusNotFound},
	}
	for _, tc := range tests {
		rec, _ := do(t, h, httptest.NewRequest(http.MethodGet, tc.path, nil))
		if rec.Code != tc.code {
			t.Errorf("GET %s: status = %d, want %d", tc.path, rec.Code, tc.code)
		}
	}
}

func TestSaveCreatesAndRedirects(t *testing.T) {
	store := &memStore{}
	h := newTestServer(t, store)

	form := url.Values{
		"title":        {"  New post  "},
		"announcement": {"Short"},
		"full_text":    {"Long text"},
	}
	rec, _ := do(t, h, postForm("/save_article", form))
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want 303", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/post/1" {
		t.Errorf("Location = %q, want /post/1", loc)
	}
	if len(store.articles) != 1 || store.articles[0].Title != "New post" {
		t.Errorf("article not stored or not trimmed: %+v", store.articles)
	}
}

func TestSaveValidation(t *testing.T) {
	tests := map[string]url.Values{
		"missing title":     {"announcement": {"a"}, "full_text": {"b"}},
		"blank fields":      {"title": {"   "}, "announcement": {"a"}, "full_text": {"b"}},
		"title too long":    {"title": {strings.Repeat("x", 256)}, "announcement": {"a"}, "full_text": {"b"}},
		"announcement long": {"title": {"t"}, "announcement": {strings.Repeat("я", 256)}, "full_text": {"b"}},
	}
	for name, form := range tests {
		t.Run(name, func(t *testing.T) {
			store := &memStore{}
			rec, body := do(t, newTestServer(t, store), postForm("/save_article", form))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", rec.Code)
			}
			if !strings.Contains(body, "alert-warning") {
				t.Errorf("validation message not shown")
			}
			if len(store.articles) != 0 {
				t.Errorf("invalid article was stored")
			}
		})
	}
}

func TestOutputIsHTMLEscaped(t *testing.T) {
	store := &memStore{articles: []Article{{ID: 1, Title: "<script>alert(1)</script>", Announcement: "a", FullText: "b"}}}
	_, body := do(t, newTestServer(t, store), httptest.NewRequest(http.MethodGet, "/post/1", nil))
	if strings.Contains(body, "<script>alert(1)</script>") {
		t.Errorf("user content rendered without escaping")
	}
}

func TestMethodNotAllowed(t *testing.T) {
	rec, _ := do(t, newTestServer(t, &memStore{}), httptest.NewRequest(http.MethodGet, "/save_article", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /save_article: status = %d, want 405", rec.Code)
	}
}

func TestStaticAndHealth(t *testing.T) {
	h := newTestServer(t, &memStore{})
	for _, path := range []string{"/static/css/cover.css", "/healthz"} {
		rec, _ := do(t, h, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s: status = %d, want 200", path, rec.Code)
		}
	}
}
