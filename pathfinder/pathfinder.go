package pathfinder

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	g "github.com/notzree/wikigraph_server/graph"
)

type PathFinder interface {
	FindPathSequential(ctx context.Context, from, to string) ([]string, error)
}

type WikiPathFinder struct {
	graph g.Wikigraph
	db    *sql.DB
}

func NewWikiPathFinder(graph g.Wikigraph, db *sql.DB) *WikiPathFinder {
	return &WikiPathFinder{
		graph: graph,
		db:    db,
	}
}

func (w *WikiPathFinder) FindPathSequential(ctx context.Context, from, to string) ([]string, error) {

	from_bytes, err := w.LookupByTitle(from)
	if err != nil {
		return nil, err
	}
	to_bytes, err := w.LookupByTitle(to)
	if err != nil {
		return nil, err
	}
	byte_array, err := w.graph.FindPathSequential(int32(from_bytes), int32(to_bytes))
	if err != nil {
		return nil, err
	}
	path, err := w.LookupByOffset(byte_array)
	if err != nil {
		return nil, err
	}
	return path, nil

}

func (w *WikiPathFinder) LookupByOffset(offsets []int32) ([]string, error) {
	byteToTitle := make(map[int32]string)
	placeholders := make([]string, len(offsets))
	args := make([]interface{}, len(offsets))
	paths := make([]string, len(offsets))

	for i, offset := range offsets {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = offset
	}
	query := fmt.Sprintf("SELECT byteoffset, title FROM lookup WHERE byteoffset IN (%s)", strings.Join(placeholders, ", "))
	rows, err := w.db.Query(query, args...)
	if err != nil {
		log.Println("lookupbyoffset error for :", offsets, "err:", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var byteOffset int32
		var title string
		if err := rows.Scan(&byteOffset, &title); err != nil {
			log.Printf("Scan failed: %v", err)
			continue
		}
		byteToTitle[byteOffset] = title
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Row iteration error: %v", err)
	}

	for i, byte := range offsets {
		paths[i] = byteToTitle[byte]
	}
	return paths, nil
}

func (w *WikiPathFinder) LookupByTitle(title string) (int32, error) {
	var byteoffset int
	err := w.db.QueryRow("(SELECT lookup.byteoffset FROM lookup WHERE title = $1 LIMIT 1)UNION ALL (SELECT lookup.byteoffset FROM redirect INNER JOIN lookup ON redirect.redirect_to = lookup.title WHERE redirect.redirect_from = $1 LIMIT 1)LIMIT 1;", title).Scan(&byteoffset)
	if err != nil {
		log.Println("LookupByTitle error for :", title, "err:", err)
		return -1, err
	}
	return int32(byteoffset), nil
}

func (w *WikiPathFinder) FindPathConcurrent(ctx context.Context, from, to string) ([]string, error) {

	from_bytes, err := w.LookupByTitle(from)
	if err != nil {
		return nil, err
	}
	to_bytes, err := w.LookupByTitle(to)
	if err != nil {
		return nil, err
	}
	byte_array, err := w.graph.FindPathConcurrent(int32(from_bytes), int32(to_bytes))

	if err != nil {
		return nil, err
	}
	path, err := w.LookupByOffset(byte_array)
	if err != nil {
		return nil, err
	}
	return path, nil
}
