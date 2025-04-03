package server

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/notzree/wikigraph_server/pathfinder"
)

type Server struct {
	port string
	pf   pathfinder.WikiPathFinder
}

func NewServer(pf pathfinder.WikiPathFinder, port string) *Server {
	return &Server{
		pf:   pf,
		port: port,
	}
}

type FindPathParams struct {
	FromOffset int32  `json:"from_offset"`
	FromTitle  string `json:"from_title"`

	ToOffset int32  `json:"to_offset"`
	ToTitle  string `json:"to_title"`
}

type FindPathResponse struct {
	Path    []string `json:"path"`
	EntryId *string  `json:"entry_id"`
}

func (s *Server) FindPathSequential(w http.ResponseWriter, r *http.Request) error {
	var req FindPathParams

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return InvalidJSON()
	}

	path, err := s.pf.FindPathSequential(req.FromOffset, req.ToOffset)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, err)
	}
	id, err := s.pf.RankEntry(pathfinder.RankingArgs{
		FromOffset: req.FromOffset,
		FromTitle:  req.FromTitle,

		ToOffset: req.ToOffset,
		ToTitle:  req.ToTitle,
		Length:   len(path),
	})
	if err != nil {
		log.Printf("error ranking entry: %v", err)
		return NewApiError(http.StatusInternalServerError, err)
	}
	res := FindPathResponse{
		Path:    path,
		EntryId: id,
	}
	return writeJSON(w, http.StatusOK, res)
}

type ClaimEntryParams struct {
	EntryId string `json:"entry_id"`
	Name    string `json:"name"`
}
type ClaimEntryResponse struct {
	Claimed bool `json:"claimed"`
}

func (s *Server) ClaimEntry(w http.ResponseWriter, r *http.Request) error {
	var req ClaimEntryParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return InvalidJSON()
	}
	claimed, err := s.pf.ClaimRankedEntry(req.EntryId, req.Name)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, err)
	}
	return writeJSON(w, http.StatusOK, ClaimEntryResponse{Claimed: claimed})

}

type CompleteParams struct {
	SearchTerm string `json:"search_term"`
}

func (s *Server) Complete(w http.ResponseWriter, r *http.Request) error {
	var req CompleteParams
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return InvalidJSON()
	}
	results, err := s.pf.Complete(r.Context(), req.SearchTerm, 10)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, err)
	}
	return writeJSON(w, http.StatusOK, results)
}

func (s *Server) Healthz(w http.ResponseWriter, r *http.Request) error {
	return writeJSON(w, http.StatusOK, "alive")
}

func (s *Server) GetLeaderboard(w http.ResponseWriter, r *http.Request) error {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		return NewApiError(http.StatusMethodNotAllowed, fmt.Errorf("method not allowed"))
	}

	// Call the business logic in WikiPathFinder
	response, err := s.pf.GetLeaderboard(r.Context())
	if err != nil {
		return NewApiError(http.StatusInternalServerError, err)
	}

	// Return the response
	return writeJSON(w, http.StatusOK, response)
}

// func (s *Server) ConcurrentFindPath(w http.ResponseWriter, r *http.Request) error {
// 	ctx := r.Context()
// 	var req FindPathParams

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		return InvalidJSON()
// 	}
// 	path, err := s.pf.FindPathConcurrent(ctx, req.From, req.To)
// 	if err != nil {
// 		return NewApiError(http.StatusInternalServerError, err)
// 	}
// 	return writeJSON(w, http.StatusOK, path)
// }

func (s *Server) Serve() error {
	log.Println("HTTP Server listening on", s.port)
	http.HandleFunc("/search", Make(s.FindPathSequential, true))
	http.HandleFunc("/complete", Make(s.Complete, true))
	http.HandleFunc("/healthz", Make(s.Healthz, false))
	http.HandleFunc("/claim", Make(s.ClaimEntry, true))
	http.HandleFunc("/leaderboard", Make(s.GetLeaderboard, false))
	return http.ListenAndServe(s.port, nil)
}

func (s *Server) ServeHTTPS(certFile, keyFile string) error {
	log.Println("HTTPS Server starting on", s.port)
	log.Println("Using certificate file:", certFile)
	log.Println("Using key file:", keyFile)

	// Check if certificate files exist
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Fatalf("Certificate file not found: %s", certFile)
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		log.Fatalf("Key file not found: %s", keyFile)
	}

	// Register handlers
	http.HandleFunc("/search", Make(s.FindPathSequential, true))
	http.HandleFunc("/complete", Make(s.Complete, true))
	http.HandleFunc("/healthz", Make(s.Healthz, false))
	http.HandleFunc("/claim", Make(s.ClaimEntry, true))
	http.HandleFunc("/leaderboard", Make(s.GetLeaderboard, false))

	// Create a custom server with timeouts and detailed TLS config
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	server := &http.Server{
		Addr:         s.port,
		TLSConfig:    tlsConfig,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	log.Println("Starting HTTPS server...")
	return server.ListenAndServeTLS(certFile, keyFile)
}
