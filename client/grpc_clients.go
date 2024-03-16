package client

//Auto-complete client
import (
	"github.com/notzree/wikigraph_server/proto"
	"google.golang.org/grpc"
)

func NewGRPCAutoCompleterClient(remoteAddr string) (proto.AutoCompleteClient, error) {
	conn, err := grpc.Dial(remoteAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	c := proto.NewAutoCompleteClient(conn)

	return c, nil
}

func NewGRPCPathFinderClient(remoteAddr string) (proto.PathFinderClient, error) {
	conn, err := grpc.Dial(remoteAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	c := proto.NewPathFinderClient(conn)

	return c, nil
}
