package main

import (
	"context"
	"log"

	db "github.com/notzree/wikigraph_server/database"
	g "github.com/notzree/wikigraph_server/graph"
)

type PathFinder interface {
	FindPath(ctx context.Context, from, to string) ([]string, error)
}

type WikigraphPathFinder struct {
	graph         g.Wikigraph
	lookupHandler db.LookupHandler
}

func (w *WikigraphPathFinder) FindPath(ctx context.Context, from, to string) ([]string, error) {

	from_bytes, err := w.lookupHandler.LookupByTitle(from)
	if err != nil {
		return nil, err
	}
	log.Println("from_bytes", from_bytes)
	to_bytes, err := w.lookupHandler.LookupByTitle(to)
	if err != nil {
		return nil, err
	}
	log.Println("to_bytes", to_bytes)

	byte_array, err := w.graph.FindPath(int32(from_bytes), int32(to_bytes))
	path := make([]string, len(byte_array))
	if err != nil {
		return nil, err
	}
	for i, byte_offset := range byte_array {
		title, err := w.lookupHandler.LookupByOffset(byte_offset)
		if err != nil {
			return nil, err
		}
		path[i] = title
	}
	return path, nil

}
