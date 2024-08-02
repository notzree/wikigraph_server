package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/notzree/wikigraph_server/graph"
	"github.com/notzree/wikigraph_server/pathfinder"
	"github.com/notzree/wikigraph_server/server"
)

const MAX_REQ_PER_HOUR = 60

func main() {
	// setup environment
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	var (
		pf_port = ":8080"
		db_url  = os.Getenv("DATABASE_URL")
	)

	// open DB connection
	conn, err := sql.Open("postgres", db_url)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	sequential_wikigraph := graph.MustCreateNewWikigraph("wikipedia_binary_graph.bin")

	sequential_pathfinder := pathfinder.NewWikiPathFinder(
		*sequential_wikigraph, conn,
	)
	server := server.NewServer(*sequential_pathfinder, pf_port)
	log.Fatal(server.Serve())

}
