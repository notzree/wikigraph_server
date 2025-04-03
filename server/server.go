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
	From int32 `json:"from"`
	To   int32 `json:"to"`
}

func (s *Server) FindPathSequential(w http.ResponseWriter, r *http.Request) error {
	var req FindPathParams

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return InvalidJSON()
	}

	path, err := s.pf.FindPathSequential(req.From, req.To)
	if err != nil {
		return NewApiError(http.StatusInternalServerError, err)
	}
	return writeJSON(w, http.StatusOK, path)
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
	http.HandleFunc("/search", Make(s.FindPathSequential))
	http.HandleFunc("/complete", Make(s.Complete))
	http.HandleFunc("/healthz", Make(s.Healthz))
	return http.ListenAndServe(s.port, nil)
}

func (s *Server) ServeHTTPS(certFile, keyFile string) error {
	log.Println("HTTPS Server listening on", s.port)

	// Register handlers
	http.HandleFunc("/search", Make(s.FindPathSequential))
	http.HandleFunc("/complete", Make(s.Complete))
	http.HandleFunc("/healthz", Make(s.Healthz))

	// Start HTTPS server with TLS certificates
	return http.ListenAndServeTLS(s.port, certFile, keyFile, nil)
}
