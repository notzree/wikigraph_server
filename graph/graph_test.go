package graph

import (
	"testing"
)

func ToByteOffset(index int32) int32 {
	return index * 4
}

func TestMustCreateWikigraph(t *testing.T) {
	_ = MustCreateNewWikigraph("../simplewiki_binary_graph.bin") //using scaled down graph for testing
}

func CreateTestGraph() []int32 {
	return []int32{
		0, 0, 1, 1, // File Header
		0, 0, 0, 3, ToByteOffset(11), ToByteOffset(21), ToByteOffset(26), // node 1 / index offset 4
		0, 0, 0, 2, ToByteOffset(4), ToByteOffset(26), // node 2 / index offset 11
		0, 0, 0, 0, // node 3 / index offset 17
		0, 0, 0, 1, ToByteOffset(17), // node 4 / index offset 21
		0, 0, 0, 4, ToByteOffset(4), ToByteOffset(11), ToByteOffset(17), ToByteOffset(21), // node 5 / index offset 26
	}
}

func TestGetLinks(t *testing.T) {
	w := Wikigraph{
		graph:     CreateTestGraph(),
		pageCount: 1,
	}

	expected := map[int32]int32{
		ToByteOffset(4):  3, // 1
		ToByteOffset(11): 2, // 2
		ToByteOffset(17): 0, // 3
		ToByteOffset(21): 1, // 4
		ToByteOffset(26): 4, // 5
	}
	for byte_offset, num_links := range expected {
		links := w.GetLinks(byte_offset)
		if len(links) != int(num_links) {
			t.Errorf("Expected %d links, got %d", num_links, len(links))
		}
	}
}

type TestFindPathSequentialTestCase struct {
	Start        int32
	Target       int32
	ExpectedPath *[]int32
	LastNode     *int32
}

func NewTestFindPathSequentialTestCase(start, target int32, expectedPath *[]int32) TestFindPathSequentialTestCase {
	return TestFindPathSequentialTestCase{
		Start:        start,
		Target:       target,
		ExpectedPath: expectedPath,
	}
}

func TestFindPathSequential(t *testing.T) {
	w := Wikigraph{
		graph:     CreateTestGraph(),
		pageCount: 1,
	}
	expected := []TestFindPathSequentialTestCase{
		NewTestFindPathSequentialTestCase(ToByteOffset(4), ToByteOffset(4), &[]int32{ToByteOffset(4)}),                                      //same start & end
		NewTestFindPathSequentialTestCase(ToByteOffset(4), ToByteOffset(17), &[]int32{ToByteOffset(4), ToByteOffset(21), ToByteOffset(17)}), //multi-step path
		NewTestFindPathSequentialTestCase(ToByteOffset(4), ToByteOffset(26), &[]int32{ToByteOffset(4), ToByteOffset(26)}),                   // direct path
		NewTestFindPathSequentialTestCase(ToByteOffset(17), ToByteOffset(11), &[]int32{}),                                                   // Unreachable path
	}
	for _, test := range expected {
		path, _ := w.FindPathSequential(test.Start, test.Target)
		if test.ExpectedPath != nil {
			if len(path) != len(*test.ExpectedPath) {
				t.Errorf("Expected path length %d, got %d", len(*test.ExpectedPath), len(path))
			}
			for i, node := range *test.ExpectedPath {
				if path[i] != node {
					t.Errorf("Expected path %v, got %v", *test.ExpectedPath, path)
					break
				}
			}
		}
	}
}
