package main

import (
	"context"
	"log"
	"net"

	"github.com/notzree/wikigraph_server/proto"
	"google.golang.org/grpc"
)

type GRPCAutoCompleteServer struct {
	svc AutoCompleter
	proto.UnimplementedAutoCompleteServer
}

func NewGRPCAutoCompleteServer(svc AutoCompleter) *GRPCAutoCompleteServer {
	return &GRPCAutoCompleteServer{
		svc: svc,
	}
}

func BuildAndRunAutoCompleteServer(svc AutoCompleter, listenAddr string) error {
	grpcAutoCompleter := NewGRPCAutoCompleteServer(svc)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return err
	}
	opts := []grpc.ServerOption{}
	server := grpc.NewServer(opts...)
	proto.RegisterAutoCompleteServer(server, grpcAutoCompleter)
	log.Println("Starting gRPC AutoCompleter server at", listenAddr)
	return server.Serve(ln)
}

func (pf *GRPCAutoCompleteServer) Complete(ctx context.Context, req *proto.CompleteRequest) (*proto.CompleteResponse, error) {
	completions, err := pf.svc.Complete(ctx, req.Prefix) // Server handles no logic
	if err != nil {
		return nil, err
	}
	resp := &proto.CompleteResponse{Completions: completions}

	return resp, err
}
