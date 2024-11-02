package database

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/iamhectorsosa/anchor/internal/store"

	_ "github.com/tursodatabase/go-libsql"
)

type Store struct {
	db *sql.DB
}

func New() (store *Store, cleanup func() error, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, err
	}

	dbPath := filepath.Join(homeDir, ".config", "anchor", "local.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), os.ModePerm); err != nil {
		return nil, nil, err
	}

	db, err := sql.Open("libsql", "file:"+dbPath)
	if err != nil {
		return nil, nil, err
	}

	cleanup = db.Close
	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS anchors (id INTEGER PRIMARY KEY, key TEXT UNIQUE, value TEXT)`); err != nil {
		return nil, cleanup, err
	}

	return &Store{db}, cleanup, nil
}

func NewInMemory() (store *Store, cleanup func() error, err error) {
	db, err := sql.Open("libsql", ":memory:")
	if err != nil {
		return nil, nil, err
	}
	cleanup = db.Close

	if _, err = db.Exec(`CREATE TABLE IF NOT EXISTS anchors (id INTEGER PRIMARY KEY, key TEXT UNIQUE, value TEXT)`); err != nil {
		return nil, cleanup, err
	}
	return &Store{db}, cleanup, nil
}

func (s *Store) Create(key, value string) error {
	_, err := s.db.Exec(`INSERT INTO anchors (key, value) VALUES (?, ?)`, key, value)
	return err
}

func (s *Store) Read(key string) (store.Anchor, error) {
	var anchor store.Anchor
	err := s.db.QueryRow("SELECT * FROM anchors WHERE key = ?", key).Scan(
		&anchor.Id,
		&anchor.Key,
		&anchor.Value,
	)
	return anchor, err
}

func (s *Store) ReadAll() ([]store.Anchor, error) {
	rows, err := s.db.Query("SELECT * FROM anchors")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var anchors []store.Anchor
	for rows.Next() {
		var anchor store.Anchor
		if err := rows.Scan(&anchor.Id, &anchor.Key, &anchor.Value); err != nil {
			return nil, err
		}
		anchors = append(anchors, anchor)
	}
	if err := rows.Err(); err != nil {
		return anchors, err
	}
	return anchors, nil
}

func (s *Store) Update(anchor store.Anchor) error {
	_, err := s.db.Exec(`UPDATE anchors SET key = ?, value = ? WHERE key = ?`, anchor.Key, anchor.Value, anchor.Key)
	return err
}

func (s *Store) Delete(Key string) error {
	_, err := s.db.Exec("DELETE FROM anchors WHERE key = ?", Key)
	return err
}

func (s *Store) Reset() error {
	_, err := s.db.Exec("DELETE FROM anchors")
	return err
}

func (s *Store) Import(anchors []store.Anchor) error {
	if len(anchors) == 0 {
		return nil
	}

	insertQuery := "INSERT OR IGNORE INTO anchors (key, value) VALUES "
	var args []interface{}
	for i, anchor := range anchors {
		if i > 0 {
			insertQuery += ", "
		}
		insertQuery += "(?, ?)"
		args = append(args, anchor.Key, anchor.Value)
	}

	_, err := s.db.Exec(insertQuery, args...)
	return err
}
