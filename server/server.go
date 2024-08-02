package server

import (
	"encoding/json"
	"log"
	"net/http"

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
	From string `json:"from"`
	To   string `json:"to"`
}

func (s *Server) FindPathSequential(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	var req FindPathParams

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return InvalidJSON()
	}

	path, err := s.pf.FindPathSequential(ctx, req.From, req.To)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, err)
	}
	return writeJSON(w, http.StatusOK, path)
}

func (s *Server) ConcurrentFindPath(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()
	var req FindPathParams

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return InvalidJSON()
	}
	path, err := s.pf.FindPathConcurrent(ctx, req.From, req.To)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, err)
	}
	return writeJSON(w, http.StatusOK, path)
}

func (s *Server) Serve() error {
	log.Println("Server listening on", s.port)
	http.HandleFunc("/search/sequential", Make(s.FindPathSequential))
	http.HandleFunc("/search/concurrent", Make(s.ConcurrentFindPath))
	return http.ListenAndServe(s.port, nil)
}
