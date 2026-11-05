//go:build integration

// Integration tests for MySQLStore. They need a running MySQL with the schema
// from init/db_schema.sql loaded and the connection settings in DB_* env vars:
//
//	DB_USERNAME=gotest DB_PASSWORD=... DB_NAME=gotest go test -tags integration ./...
package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func openTestStore(t *testing.T) *MySQLStore {
	t.Helper()
	cfg, err := LoadConfig("")
	if err != nil {
		t.Skipf("integration config not set: %v", err)
	}
	db, err := sql.Open("mysql", cfg.DSN())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("database not reachable: %v", err)
	}
	if _, err := db.Exec("DELETE FROM articles"); err != nil {
		t.Fatalf("clean table: %v", err)
	}
	return NewMySQLStore(db)
}

func TestMySQLStoreRoundTrip(t *testing.T) {
	store := openTestStore(t)
	ctx := context.Background()

	// Input that would have broken the old string-concatenated SQL.
	in := Article{
		Title:        "It's a test'); DROP TABLE articles; --",
		Announcement: `quotes " and \ backslash`,
		FullText:     "Юнікод ✓",
	}
	id, err := store.Create(ctx, in)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := store.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Title != in.Title || got.Announcement != in.Announcement || got.FullText != in.FullText {
		t.Errorf("round trip mismatch:\n got  %+v\n want %+v", got, in)
	}

	list, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].ID != id {
		t.Errorf("List = %+v, want one article with id %d", list, id)
	}

	if _, err := store.Get(ctx, id+1000); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get(missing) err = %v, want ErrNotFound", err)
	}
}
