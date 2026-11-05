package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Article is a single blog post.
type Article struct {
	ID           uint64
	Title        string
	Announcement string
	FullText     string
}

// ErrNotFound is returned when an article does not exist.
var ErrNotFound = errors.New("article not found")

// Store is the persistence interface used by the HTTP handlers.
// Keeping it small lets the handlers be tested without a database.
type Store interface {
	List(ctx context.Context) ([]Article, error)
	Get(ctx context.Context, id uint64) (Article, error)
	Create(ctx context.Context, a Article) (uint64, error)
}

// MySQLStore implements Store on top of database/sql.
// All queries use placeholders, so user input is never concatenated into SQL.
type MySQLStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) *MySQLStore {
	return &MySQLStore{db: db}
}

func (s *MySQLStore) List(ctx context.Context) ([]Article, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT id, title, announcement, full_text FROM articles ORDER BY id DESC")
	if err != nil {
		return nil, fmt.Errorf("list articles: %w", err)
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		var fullText sql.NullString
		if err := rows.Scan(&a.ID, &a.Title, &a.Announcement, &fullText); err != nil {
			return nil, fmt.Errorf("scan article: %w", err)
		}
		a.FullText = fullText.String
		articles = append(articles, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate articles: %w", err)
	}
	return articles, nil
}

func (s *MySQLStore) Get(ctx context.Context, id uint64) (Article, error) {
	var a Article
	var fullText sql.NullString
	err := s.db.QueryRowContext(ctx,
		"SELECT id, title, announcement, full_text FROM articles WHERE id = ?", id).
		Scan(&a.ID, &a.Title, &a.Announcement, &fullText)
	if errors.Is(err, sql.ErrNoRows) {
		return Article{}, ErrNotFound
	}
	if err != nil {
		return Article{}, fmt.Errorf("get article %d: %w", id, err)
	}
	a.FullText = fullText.String
	return a, nil
}

func (s *MySQLStore) Create(ctx context.Context, a Article) (uint64, error) {
	res, err := s.db.ExecContext(ctx,
		"INSERT INTO articles (title, announcement, full_text) VALUES (?, ?, ?)",
		a.Title, a.Announcement, a.FullText)
	if err != nil {
		return 0, fmt.Errorf("insert article: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	return uint64(id), nil
}
