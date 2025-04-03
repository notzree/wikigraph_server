package pathfinder

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

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

type RankingArgs struct {
	FromTitle  string
	FromOffset int32

	ToTitle  string
	ToOffset int32

	Length int
}

func (w *WikiPathFinder) RankEntry(args RankingArgs) (*string, error) {
	// Use INSERT with ON CONFLICT DO NOTHING to handle the case where entry already exists
	insertQuery := `
        INSERT INTO search_leaderboard (
            fromTitle,
            fromOffset,
            toTitle,
            toOffset,
            length,
            name
        )
        VALUES ($1, $2, $3, $4, $5, '')
        ON CONFLICT (fromOffset, toOffset) DO NOTHING
        RETURNING id
    `

	var newId string
	err := w.db.QueryRow(
		insertQuery,
		args.FromTitle,
		args.FromOffset,
		args.ToTitle,
		args.ToOffset,
		args.Length,
	).Scan(&newId)

	if err == sql.ErrNoRows {
		// This means the ON CONFLICT clause was triggered - the entry already exists
		return nil, nil
	} else if err != nil {
		// A different error occurred
		return nil, fmt.Errorf("error inserting entry: %w", err)
	}

	// If we get here, the insertion was successful and we have the new ID
	return &newId, nil
}

func (w *WikiPathFinder) ClaimRankedEntry(id string, name string) (bool, error) {
	updateQuery := `
        UPDATE search_leaderboard
        SET name = $1
        WHERE id = $2 AND name = ''
    `

	result, err := w.db.Exec(updateQuery, name, id)
	if err != nil {
		return false, fmt.Errorf("error updating entry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return false, nil
	}

	return true, nil
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

// LeaderboardEntry represents a single entry in the leaderboard
type LeaderboardEntry struct {
	ID         string    `json:"id"`
	FromTitle  string    `json:"fromTitle"`
	ToTitle    string    `json:"toTitle"`
	FromOffset int32     `json:"fromOffset"`
	ToOffset   int32     `json:"toOffset"`
	Date       time.Time `json:"date"`
	Length     int32     `json:"length"`
	Name       string    `json:"name"`
}

// LeaderboardResponse is the structure of the API response
type LeaderboardResponse struct {
	Entries []LeaderboardEntry `json:"entries"`
	Total   int                `json:"total"`
}

// GetLeaderboard returns the top 100 leaderboard entries sorted by length and date
func (w *WikiPathFinder) GetLeaderboard(ctx context.Context) (*LeaderboardResponse, error) {
	// Query to get top 100 entries sorted by length (desc) and then date
	query := `
        SELECT id, fromTitle, toTitle, fromOffset, toOffset, date, length, name
        FROM search_leaderboard
        ORDER BY length DESC, date ASC
        LIMIT 100
    `

	rows, err := w.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error querying leaderboard: %w", err)
	}
	defer rows.Close()

	// Parse the results
	var entries []LeaderboardEntry
	for rows.Next() {
		var entry LeaderboardEntry
		err := rows.Scan(
			&entry.ID,
			&entry.FromTitle,
			&entry.ToTitle,
			&entry.FromOffset,
			&entry.ToOffset,
			&entry.Date,
			&entry.Length,
			&entry.Name,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		entries = append(entries, entry)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Get total count of entries for additional context
	var total int
	countQuery := `SELECT COUNT(*) FROM search_leaderboard`
	err = w.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("error counting total entries: %w", err)
	}

	// Create and return the response
	response := &LeaderboardResponse{
		Entries: entries,
		Total:   total,
	}

	return response, nil
}
