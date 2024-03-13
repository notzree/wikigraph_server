package database

import "database/sql"

type LookupHandler interface {
	LookupByOffset(offset int32) (string, error)
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
