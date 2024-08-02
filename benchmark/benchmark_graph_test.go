package benchmark

import (
	"testing"

	"github.com/notzree/wikigraph_server/graph"
)

func BenchmarkSequentialSearch(b *testing.B) {
	w := graph.MustCreateNewWikigraph("../wikipedia_binary_graph.bin")
	university_of_waterloo := int32(27845660)
	outer_mongolia := int32(44956584)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.FindPathSequential(university_of_waterloo, outer_mongolia)
	}
}

func BenchmarkConcurrentSearch(b *testing.B) {
	w := graph.MustCreateNewWikigraph("../wikipedia_binary_graph.bin")
	university_of_waterloo := int32(27845660)
	outer_mongolia := int32(44956584)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.FindPathConcurrent(university_of_waterloo, outer_mongolia)
	}
}
