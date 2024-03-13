package main

import (
	"fmt"
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
	fmt.Println("Server starting on", server.address)
	return http.ListenAndServe(server.address, nil)
}
