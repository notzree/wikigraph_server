package main

import (
	"context"
	"database/sql"

	g "github.com/notzree/wikigraph_server/graph"
)

type PathFinder interface {
	FindPath(ctx context.Context, from, to string) ([]string, error)
}

type WikigraphPathFinder struct {
	graph *g.Wikigraph
	db    *sql.DB
}

func (w *WikigraphPathFinder) FindPath(ctx context.Context, from, to string) ([]string, error) {
	print("Finding path from ", from, " to ", to)
	return []string{"path1", "path2"}, nil

}
