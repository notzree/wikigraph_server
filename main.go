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

const MAX_REQ_PER_HOUR = 10

// func main() {
// 	graph := g.MustCreateNewWikigraph("simplewiki_binary_graph.bin")
// 	start_bytes := int32(9332)  //atom
// 	end_bytes := int32(2857212) //fingre
// 	// log.Println(graph.GetLinks(start_bytes))
// 	// log.Println(graph.GetLinks(end_bytes))
// 	res, err := graph.FindPath(start_bytes, end_bytes)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Println("Path from atom to google:", res)

// }

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

	grpcClient, err := c.NewGRPCClient(":8080")
	if err != nil {
		log.Fatal("Failed to connect to grpc server:", err)
	}

	// go func() {
	// 	time.Sleep(5 * time.Second)
	// 	r, err := grpcClient.FindPath(ctx, &proto.PathRequest{From: "A", To: "B"})
	// 	if err != nil {
	// 		log.Fatal("Failed to find path:", err)
	// 	}
	// 	log.Println("Path from A to B:", r.Paths)

	// }()
	//start pathfinder grpc service (grpc)
	var (
		graph         = g.MustCreateNewWikigraph("simplewiki_binary_graph.bin")
		lookupHandler = w.NewWikigraphLookupHandler(conn)
		pf            = &WikigraphPathFinder{graph: *graph, lookupHandler: lookupHandler}
	)
	go BuildAndRunGRPCServer(pf, pf_port)

	//start rate_limiter service (json)
	var (
		rate_limited_pf = NewRateLimitedService(client, ctx, grpcClient, 60, MAX_REQ_PER_HOUR)

		rate_limit_server = NewApiServer(rl_port)
	)
	rate_limit_server.RegisterService("/find_path", rate_limited_pf)
	if err := rate_limit_server.Start(); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
