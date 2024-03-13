package main

import (
	"log"
	"net/http"
)

type ApiServer struct {
	address  string
	services map[string]Servicer // Map paths to services
}

func NewApiServer(address string) *ApiServer {
	return &ApiServer{
		address:  address,
		services: make(map[string]Servicer),
	}
}
func (server *ApiServer) RegisterService(path string, service Servicer) {
	server.services[path] = service
}

func (server *ApiServer) Start() error {
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
