package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

const MAX_REQ_PER_HOUR = 10

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var ctx = context.Background()
	opt, _ := redis.ParseURL(os.Getenv("UPSTASH_URL"))
	client := redis.NewClient(opt)
	exampleService := &ExampleService{}
	rate_limited_service := NewRateLimitedService(exampleService, client, ctx, 60, MAX_REQ_PER_HOUR)
	server := NewApiServer(":8080")
	server.RegisterService("/example", rate_limited_service)

	if err := server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
