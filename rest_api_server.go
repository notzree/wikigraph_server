package main

import (
	"log"
	"net/http"
)

type RestApiServer struct {
	address  string
	services map[string]Servicer // Map paths to services
}

func NewRestApiServer(address string) *RestApiServer {
	return &RestApiServer{
		address:  address,
		services: make(map[string]Servicer),
	}
}
func (server *RestApiServer) RegisterService(path string, service Servicer) {
	server.services[path] = service
}

func (server *RestApiServer) Start() error {
	for path, service := range server.services {
		// Capture service in local scope for the closure passed to HandleFunc
		localService := service
		http.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			localService.Serve(w, r)
		})
	}
	log.Println("Server starting on port", server.address)
	return http.ListenAndServe(server.address, nil)
}
