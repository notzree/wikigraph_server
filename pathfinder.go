package main

import (
	"context"
	"database/sql"
	"log"

	g "github.com/notzree/wikigraph_server/graph"
)

type PathFinder interface {
	FindPath(ctx context.Context, from, to string) ([]string, error)
}

type WikigraphPathFinder struct {
	graph g.Wikigraph
	db    *sql.DB
}

func (w *WikigraphPathFinder) FindPath(ctx context.Context, from, to string) ([]string, error) {

	from_bytes, err := w.LookupByTitle(from)
	if err != nil {
		return nil, err
	}
	to_bytes, err := w.LookupByTitle(to)
	if err != nil {
		return nil, err
	}

	byte_array, err := w.graph.FindPath(int32(from_bytes), int32(to_bytes))
	path := make([]string, len(byte_array))
	if err != nil {
		return nil, err
	}
	for i, byte_offset := range byte_array {
		title, err := w.LookupByOffset(byte_offset)
		if err != nil {
			return nil, err
		}
		path[i] = title
	}
	return path, nil

}
func (w *WikigraphPathFinder) LookupByOffset(offset int32) (string, error) {
	var title string
	err := w.db.QueryRow("SELECT title FROM lookup WHERE byteoffset = $1", offset).Scan(&title)
	if err != nil {
		log.Println("lookupbyoffset error for :", offset, "err:", err)
		return "", err
	}
	return title, nil
}

func (w *WikigraphPathFinder) LookupByTitle(title string) (int32, error) {
	var byteoffset int
	err := w.db.QueryRow("(SELECT lookup.byteoffset FROM lookup WHERE title = $1 LIMIT 1)UNION ALL (SELECT lookup.byteoffset FROM redirect INNER JOIN lookup ON redirect.redirect_to = lookup.title WHERE redirect.redirect_from = $1 LIMIT 1)LIMIT 1;", title).Scan(&byteoffset)
	if err != nil {
		log.Println("LookupByTitle error for :", title, "err:", err)
		return -1, err
	}
	log.Println("LookupByTitle", title, "byteoffset", byteoffset)
	return int32(byteoffset), nil
}
