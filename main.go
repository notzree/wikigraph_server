package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

const MAX_REQ_PER_HOUR = 10

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
	)

	//start pathfinder grpc service (grpc)
	go BuildAndRunGRPCServer(&WikigraphPathFinder{}, pf_port)

	//start rate_limiter service (json)
	rate_limited_service := NewRateLimitedService(client, ctx, 60, MAX_REQ_PER_HOUR)
	rate_limit_server := NewApiServer(rl_port)
	rate_limit_server.RegisterService("/find_path", rate_limited_service)
	if err := rate_limit_server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
