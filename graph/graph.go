package graph

import (
	"encoding/binary"
	"os"
)

type Grapher interface {
	FindPath(from, to string) ([]string, error)
}

type Wikigraph struct {
	graph     []int32 //graph is of little endian i32 integers
	pageCount int32
}

// MustCreateNewWikigraph attemps to build a new wikigraph from the file at fp
func MustCreateNewWikigraph(fp string) *Wikigraph {
	file, err := os.Open(fp)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	filesize := fileInfo.Size()
	w := &Wikigraph{}
	w.graph = make([]int32, filesize/4)
	w.pageCount = int32(filesize / 4)
	err = binary.Read(file, binary.LittleEndian, w.graph)
	if err != nil {
		panic(err)
	}
	return w
}

// MustgetLinks takes in a byteoffset and returns an array representing the links (also bytesoffsets)
func (w *Wikigraph) GetLinks(bo int32) []int32 {
	index := to_i(bo)
	num_links := w.graph[index+3] //3 unused bytes (0s)
	if num_links == 0 {
		return []int32{}
	}
	// println("num_links", num_links)

	start := index + 4

	return w.graph[start : start+num_links] //skip the first 4 bytes (node header)
}

// Findpath returns the shortest path from "from" to "to". "from" and "to" are byte offsets into the array (which represent nodes).
func (w *Wikigraph) FindPath(start, target int32) ([]int32, error) {
	parents := make([]int32, w.pageCount) //predecessor array. The index should not be a byteoffset, but the index
	for i := range parents {
		parents[i] = -1
	}
	visited := make([]bool, w.pageCount)
	queue := make([]int32, 0)
	queue = append(queue, start)
	visited[to_i(start)] = true
	for len(queue) > 0 {
		current_node := queue[0] //node is a byteoffset
		queue = queue[1:]
		if current_node == 0 {
			continue
		}
		if current_node == target {
			break
		}
		for _, link_byte_offset := range w.GetLinks(current_node) {
			if link_byte_offset == 0 {
				continue
			}
			link := to_i(link_byte_offset)
			if !visited[link] { //unvisited
				visited[link] = true
				parents[link] = to_i(current_node)      //mark as visited and mark predecessor
				queue = append(queue, link_byte_offset) //add the bo to queue as GetLinks takes in byteoffsets
			}
		}
	}

	var path []int32

	for v := to_i(target); v != -1; v = parents[v] {
		path = append(path, v*4) //return byteoffsets
		if v == start {
			break
		}
	}
	reverse(path)

	return path, nil
}

func (w *Wikigraph) Peek(i int32) []int32 {
	peek := make([]int32, 4)
	copy(peek, w.graph[i:])
	return peek
}

func reverse(s []int32) { //reminder to self that slices are passed by reference
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

}
func to_i(byte_offset int32) int32 {
	return byte_offset / 4
}
