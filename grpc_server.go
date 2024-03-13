package main

import (
	"context"
	"log"
	"net"

	"github.com/notzree/wikigraph_server/proto"
	"google.golang.org/grpc"
)

func BuildAndRunGRPCServer(svc PathFinder, listenAddr string) error {
	grpcPathFinder := NewGRPCPathFinderServer(svc)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)
	proto.RegisterPathFinderServer(server, grpcPathFinder)
	log.Println("Starting gRPC server at", listenAddr)
	return server.Serve(ln)

}

type GRPCPathFinderServer struct {
	svc PathFinder
	proto.UnimplementedPathFinderServer
}

func NewGRPCPathFinderServer(svc PathFinder) *GRPCPathFinderServer {
	return &GRPCPathFinderServer{
		svc: svc,
	}
}

func (pf *GRPCPathFinderServer) FindPath(ctx context.Context, req *proto.PathRequest) (*proto.PathResponse, error) {
	paths, err := pf.svc.FindPath(ctx, req.From, req.To) // this way the server does not need to know any business logic!
	if err != nil {
		return nil, err
	}
	resp := &proto.PathResponse{Paths: paths}

	return resp, err
}
