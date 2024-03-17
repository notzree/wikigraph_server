package main

import (
	"testing"

	g "github.com/notzree/wikigraph_server/graph"
)

type StringPair struct {
	First  int32
	Second int32
}

func BenchmarkProcessData(b *testing.B) {
	pairs := []StringPair{
		{148190564, 347026072},
		{26245292, 102589784},
		{5485692, 314943772},
	}
	graph := g.MustCreateNewWikigraph("wikipedia_binary_graph.bin")
	b.ResetTimer()
	for _, tuple := range pairs {
		graph.FindPath(tuple.First, tuple.Second)
	}
}
