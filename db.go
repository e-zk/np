package main

import (
	"database/sql"
	//	"errors"

	_ "github.com/mattn/go-sqlite3"
)

/* a log entry */
type Entry struct {
	ID        int64
	Artist    string
	Title     string
	Type      int64
	Link      string
	DateAdded string
}

type SQLiteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(db *sql.DB) *SQLiteRepo {
	return &SQLiteRepo{
		db: db,
	}
}

func (r *SQLiteRepo) Migrate() error {
	q := `
	CREATE TABLE IF NOT EXISTS nowplaying(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		artist TEXT NOT NULL,
		title TEXT NOT NULL,
		type INTEGER NOT NULL DEFAULT '0',
		link TEXT,
		date_added TEXT NOT NULL
	);
	`

	_, err := r.db.Exec(q)
	return err
}

func (r *SQLiteRepo) Add(e Entry) (*Entry, error) {
	res, err := r.db.Exec("INSERT INTO nowplaying(artist, title, type, link, date_added) values(?,?,?,?,?)", e.Artist, e.Title, e.Type, e.Link, e.DateAdded)
	if err != nil {
		//var sqlErr sqlite3.Error
		//if errors.As(err, &sqlErr) {
		//	if errors.Is(sqlErr.ExtendCode, sqlite3.ErrConstraintUnique) {
		//		return nil, ErrDuplicate
		//	}
		//}
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	e.ID = id

	return &e, nil
}

func (r *SQLiteRepo) All() ([]Entry, error) {
	entries, err := r.db.Query("SELECT * FROM nowplaying")
	if err != nil {
		return nil, err
	}
	defer entries.Close()

	var all []Entry
	for entries.Next() {
		var e Entry
		if err := entries.Scan(&e.ID, &e.Artist, &e.Title, &e.Type, &e.Link, &e.DateAdded); err != nil {
			return nil, err
		}
		all = append(all, e)
	}
	return all, nil
}
