package main

import "context"

type PathFinder interface {
	FindPath(ctx context.Context, from, to string) ([]string, error)
}

type WikigraphPathFinder struct{ graph_path string }

func (w *WikigraphPathFinder) FindPath(ctx context.Context, from, to string) ([]string, error) {

}
