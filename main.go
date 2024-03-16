package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	c "github.com/notzree/wikigraph_server/client"
	w "github.com/notzree/wikigraph_server/database"
	g "github.com/notzree/wikigraph_server/graph"
	"github.com/redis/go-redis/v9"
)

const MAX_REQ_PER_HOUR = 60

func main() {
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

	grpcClient, err := c.NewGRPCPathFinderClient(pf_port)
	if err != nil {
		log.Fatal("Failed to connect to grpc server:", err)
	}

	var (
		graph         = g.MustCreateNewWikigraph("wikipedia_binary_graph.bin")
		lookupHandler = w.NewWikigraphLookupHandler(conn)
		pf            = &WikigraphPathFinder{graph: *graph, lookupHandler: lookupHandler}
		pfs           = NewPathFinderService(grpcClient, ctx)
	)
	go BuildAndRunGRPCServer(pf, pf_port)

	//start rate_limiter service (json)
	var (
		rate_limited_pf   = NewRateLimitedService(client, ctx, pfs, 60, MAX_REQ_PER_HOUR)
		rate_limit_server = NewRestApiServer(rl_port)
	)

	rate_limit_server.RegisterService("/find_path", rate_limited_pf)
	if err := rate_limit_server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
