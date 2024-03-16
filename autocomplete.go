package main

import (
	"context"
	"database/sql"
)

type AutoCompleter interface {
	Complete(ctx context.Context, prefix string) ([]string, error)
}

type WikigraphAutoCompleter struct {
	db *sql.DB
}

func (w *WikigraphAutoCompleter) Complete(ctx context.Context, prefix string) ([]string, error) {
	rows, err := w.db.Query(`SELECT title FROM lookup WHERE title % $1 ORDER BY ORDER BY LENGTH(title) DESC LIMIT 10;`, prefix)
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
