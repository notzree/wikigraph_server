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
func (w *Wikigraph) MustGetLinks(bo int32) []int32 {
	index := bo / 4
	num_links := w.graph[index+3] //3 unused bytes (0s)
	last_link_index := index + num_links + 3
	return w.graph[index+4 : last_link_index] //skip the first 4 bytes (node header)
}

// Findpath returns the shortest path from "from" to "to". "from" and "to" are byte offsets into the array (which represent nodes).
func (w *Wikigraph) FindPath(from, to int32) ([]int32, error) {
	parents := make([]int32, w.pageCount) //predecessor array. The index should not be a byteoffset, but the index. convert by dividing by 4
	queue := make([]int32, 0)
	queue = append(queue, from)
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		if node == to {
			break
		}
		links := w.MustGetLinks(node)
		for _, link_byte_offset := range links {
			link := link_byte_offset / 4
			if parents[link] == 0 { //unvisited
				parents[link] = node        //current
				queue = append(queue, link) //add to queue
			}
		}
	}
	path := make([]int32, 0)
	path = append(path, to)
	target := parents[to] //target is the byteoffset of the paret of the last node

	for target != from {
		path = append(path, target) //append the byteoffset !
		target = parents[target/4]  //index using the byteoffset converted into an index
	}
	//reverse path
	reverse(path)
	return path, nil
}

func (w *Wikigraph) Peek() []int32 {
	peek := make([]int32, 10)
	copy(peek, w.graph)
	return peek
}
func reverse(s []int32) { //reminder to self that slices are passed by reference
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
