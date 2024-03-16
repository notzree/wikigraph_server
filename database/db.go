package database

import (
	"database/sql"
)

type LookupHandler interface {
	LookupByOffset(offset int32) (string, error)
	LookupByTitle(title string) (int32, error)
	CompleteString(prefix string) ([]string, error)
}

type WikigraphLookupHandler struct {
	db *sql.DB
}

func NewWikigraphLookupHandler(db *sql.DB) *WikigraphLookupHandler {
	return &WikigraphLookupHandler{db: db}
}

func (w *WikigraphLookupHandler) LookupByOffset(offset int32) (string, error) {
	var title string
	err := w.db.QueryRow("SELECT title FROM lookup WHERE byteoffset = $1", offset).Scan(&title)
	if err != nil {
		return "", err
	}
	return title, nil
}
func (w *WikigraphLookupHandler) LookupByTitle(title string) (int32, error) {
	var byteoffset int
	err := w.db.QueryRow("SELECT byteoffset FROM lookup WHERE title = $1", title).Scan(&byteoffset)
	if err != nil {
		return -1, err
	}
	return int32(byteoffset), nil
}

func (w *WikigraphLookupHandler) CompleteString(prefix string) ([]string, error) {
	rows, err := w.db.Query(`SELECT name FROM products WHERE name % $1 ORDER BY similarity(name, $1) DESC LIMIT 10;`, prefix)
	if err != nil {
		return nil, err
	}
	var titles []string
	for rows.Next() {
		var title string
		err = rows.Scan(&title)
		if err != nil {
			return nil, err
		}
		titles = append(titles, title)
	}
	return titles, nil
}
