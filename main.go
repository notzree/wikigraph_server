package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	c "github.com/notzree/wikigraph_server/client"
	w "github.com/notzree/wikigraph_server/database"
	g "github.com/notzree/wikigraph_server/graph"
	proto "github.com/notzree/wikigraph_server/proto"
	"github.com/redis/go-redis/v9"
)

const MAX_REQ_PER_HOUR = 10

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	db_url := os.Getenv("DATABASE_URL")
	conn, err := sql.Open("postgres", db_url)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	// graph := g.MustCreateNewWikigraph("simplewiki_binary_graph.bin")
	lookupHandler := w.NewWikigraphLookupHandler(conn)

	result, err := lookupHandler.LookupByOffset(16)
	if err != nil {
		panic(err)
	}
	log.Println(result)

}

func main2() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var (
		ctx     = context.Background()
		opt, _  = redis.ParseURL(os.Getenv("UPSTASH_URL"))
		client  = redis.NewClient(opt)
		rl_port = os.Getenv("RATE_LIMITER_PORT")
		pf_port = os.Getenv("PATH_FINDER_PORT")
		db_url  = os.Getenv("DATABASE_URL")
	)

	conn, err := sql.Open("postgres", db_url)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	grpcClient, err := c.NewGRPCClient(":8080")
	if err != nil {
		log.Fatal("Failed to connect to grpc server:", err)
	}

	go func() {
		time.Sleep(5 * time.Second)
		r, err := grpcClient.FindPath(ctx, &proto.PathRequest{From: "A", To: "B"})
		if err != nil {
			log.Fatal("Failed to find path:", err)
		}
		log.Println("Path from A to B:", r.Paths)

	}()
	graph := g.MustCreateNewWikigraph("simplewiki_binary_graph.bin")
	pf := &WikigraphPathFinder{graph: graph, db: conn}

	//start pathfinder grpc service (grpc)
	go BuildAndRunGRPCServer(pf, pf_port)

	//start rate_limiter service (json)
	rate_limited_service := NewRateLimitedService(client, ctx, 60, MAX_REQ_PER_HOUR)
	rate_limit_server := NewApiServer(rl_port)
	rate_limit_server.RegisterService("/find_path", rate_limited_service)
	if err := rate_limit_server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
