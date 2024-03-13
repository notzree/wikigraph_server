package main

import (
	"context"
	"database/sql"
)

type PathFinder interface {
	FindPath(ctx context.Context, from, to string) ([]string, error)
}

type WikigraphPathFinder struct {
	graph_path string
	db         *sql.DB
}

func (w *WikigraphPathFinder) FindPath(ctx context.Context, from, to string) ([]string, error) {
	print("Finding path from ", from, " to ", to)
	return []string{"path1", "path2"}, nil

}
