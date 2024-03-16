package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	c "github.com/notzree/wikigraph_server/client"
	g "github.com/notzree/wikigraph_server/graph"
	"github.com/redis/go-redis/v9"
)

const MAX_REQ_PER_HOUR = 60

func main() {
	// setup environment
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
		ac_port = os.Getenv("AUTO_COMPLETER_PORT")
		db_url  = os.Getenv("DATABASE_URL")
	)

	// open DB connection
	conn, err := sql.Open("postgres", db_url)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	//Create pathfinder dependencies
	graph := g.MustCreateNewWikigraph("wikipedia_binary_graph.bin")
	pf := &WikigraphPathFinder{graph: *graph, db: conn}
	go BuildAndRunPathFinderServer(pf, pf_port)

	pf_client, err := c.NewGRPCPathFinderClient(pf_port)
	if err != nil {
		log.Fatal("Failed to connect to grpc server:", err)
	}
	pfs := NewPathFinderService(pf_client, ctx)

	//Create AutoCompleter dependencies
	ac := &WikigraphAutoCompleter{db: conn}
	go BuildAndRunAutoCompleteServer(ac, ac_port)

	ac_client, err := c.NewGRPCAutoCompleterClient(ac_port)
	if err != nil {
		log.Fatal("Failed to connect to grpc server:", err)
	}
	acs := NewAutoCompleterService(ac_client, ctx)

	//Rate limit requests
	var (
		rl_pfs          = NewRateLimitedService(client, ctx, pfs, 60, MAX_REQ_PER_HOUR)
		rest_api_server = NewRestApiServer(rl_port)
	)

	rest_api_server.RegisterService("/find_path", rl_pfs)
	rest_api_server.RegisterService("/complete", acs)
	err = rest_api_server.Start()
	if err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
