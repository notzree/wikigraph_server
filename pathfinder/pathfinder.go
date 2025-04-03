package pathfinder

import (
	"context"
	"database/sql"
	"log"
	"strings"

	g "github.com/notzree/wikigraph_server/graph"
)

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

type CompletionResult struct {
	Title  string
	Offset int32
}

func (w *WikiPathFinder) FindPathSequential(from, to int32) ([]string, error) {

	// from_bytes, err := w.LookupByTitle(from)
	// if err != nil {
	// 	return nil, err
	// }
	// to_bytes, err := w.LookupByTitle(to)
	// if err != nil {
	// 	return nil, err
	// }
	byte_array, err := w.graph.FindPathSequential(from, to)
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

func (w *WikiPathFinder) LookupByOffset(offset int32) (string, error) {
	var title string
	err := w.db.QueryRow("SELECT title FROM lookup WHERE byteoffset = $1", offset).Scan(&title)
	if err != nil {
		log.Println("lookupbyoffset error for :", offset, "err:", err)
		return "", err
	}
	return title, nil
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

// Complete takes a search term and returns matching title completions with their byte offsets
func (w *WikiPathFinder) Complete(ctx context.Context, searchTerm string, limit int) ([]CompletionResult, error) {
	if limit <= 0 {
		limit = 10
	}
	searchTerm = strings.ToLower(searchTerm)
	rows, err := w.db.QueryContext(ctx, `
        SELECT title, byteoffset
        FROM lookup
        WHERE title ILIKE $1
        ORDER BY LENGTH(title), title
        LIMIT $2
    `, "%"+searchTerm+"%", limit)

	if err != nil {
		log.Printf("Complete error for search term: %s, err: %v", searchTerm, err)
		return nil, err
	}
	defer rows.Close()

	var results []CompletionResult
	for rows.Next() {
		var title string
		var offset int

		if err := rows.Scan(&title, &offset); err != nil {
			log.Printf("Error scanning completion row: %v", err)
			return results, err
		}

		results = append(results, CompletionResult{
			Title:  title,
			Offset: int32(offset),
		})
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating through completion results: %v", err)
		return results, err
	}

	return results, nil
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
