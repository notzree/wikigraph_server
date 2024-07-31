package graph

import (
	"encoding/binary"
	"errors"
	"os"
	"sync"
)

// Wikigraph is a struct that represents a graph of wikipedia pages

type Wikigraph struct {
	graph     []int32 //graph is of little endian i32 integers
	pageCount int32
}

// MustCreateNewWikigraph attempts to build a new wikigraph from the file at fp
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

	start := index + 4

	return w.graph[start : start+num_links] //skip the first 4 bytes (node header)
}

// Findpath returns the shortest path from "from" to "to". "from" and "to" are byte offsets into the array (which represent nodes).
func (w *Wikigraph) FindPathSequential(start, target int32) ([]int32, error) {
	if start == target {
		return []int32{start}, nil
	}
	parents := make(map[int32]int32)
	queue := make([]int32, 0)
	queue = append(queue, start)
	parents[to_i(start)] = to_i(start)

	for len(queue) > 0 {
		current_node := queue[0] // node is a byte offset
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
			if _, ok := parents[link]; !ok { // Unvisited
				parents[link] = to_i(current_node)      // Mark as visited and mark predecessor
				queue = append(queue, link_byte_offset) // Add the byte offset to queue as GetLinks takes in byte offsets.
			}
		}
	}

	var path []int32

	// Construct path. Note that path will be constructed in reverse: from target to start.
	for v := to_i(target); v != -1 && parents[v] != -1; v = parents[v] { // Check parents[v] != -1 for safety.
		path = append(path, v*4) // Return byte offsets
		if v*4 == start {
			break
		}
	}
	reverse(path)
	return path, nil
}

func (w *Wikigraph) FindPathConcurrent(start, target int32) ([]int32, error) {
	if start == target {
		return []int32{start}, nil
	}
	startParents := make(map[int32]int32)
	startParents[to_i(start)] = start

	targetParents := make(map[int32]int32)
	targetParents[to_i(target)] = target
	var wg sync.WaitGroup
	closeSignal := make(chan struct{})
	startQueue := make(chan int32)  // byte offset
	targetQueue := make(chan int32) // byte offset

	wg.Add(2)
	//from start searching for target
	go w.bfs(BfsParams{
		searchTarget: target,
		parents:      startParents,
		closeSignal:  closeSignal,
		nodeQueue:    startQueue,
		wg:           &wg,
	})
	//from target searching for start
	go w.bfs(BfsParams{
		searchTarget: start,
		parents:      targetParents,
		closeSignal:  closeSignal,
		nodeQueue:    targetQueue,
		wg:           &wg,
	})

	go func() {
		for {
			select {
			case current_node_offset := <-startQueue: //Bfs searching from start found a node_offset
				//If this node's index is in the target's parents, we have found the path
				if _, ok := targetParents[to_i(current_node_offset)]; ok {
					closeSignal <- struct{}{}
					starting_path := w.ConstructPath(start, current_node_offset, startParents) //path from currentNode -> start
					reverse(starting_path)                                                     // path from start -> currentNode
					target_path := w.ConstructPath(target, current_node_offset, targetParents) //path from currentNode -> target
					return append(starting_path, target_path...)
				}
			case node_offset := <-targetQueue: //Bfs searching from target found a node_offset
				//If this node's index is in the start's parents, we have found the path
				if _, ok := startParents[to_i(node_offset)]; ok {
					closeSignal <- struct{}{}
					//todo: construct path
					return
				}

			}
		}
	}()
	startQueue <- start
	targetQueue <- target
	wg.Wait()
	return nil, errors.New("no path found")
}

type BfsParams struct {
	searchTarget int32 //byte offset
	parents      map[int32]int32
	closeSignal  chan struct{}
	nodeQueue    chan int32
	wg           *sync.WaitGroup
}

func (w *Wikigraph) bfs(params BfsParams) {
	defer params.wg.Done()
	for {
		select {
		case <-params.closeSignal:
			return
		case current_node_offset := <-params.nodeQueue:
			if current_node_offset == params.searchTarget {
				return
			}
			if current_node_offset == 0 {
				continue
			}
			for _, link_byte_offset := range w.GetLinks(current_node_offset) {
				if link_byte_offset == 0 {
					continue
				}
				link := to_i(link_byte_offset)
				if _, ok := params.parents[link]; !ok {
					params.parents[link] = to_i(current_node_offset)
					params.nodeQueue <- link_byte_offset
				}
			}

		}
	}
}

// Constructs the path from finish -> start using the parents map * NOT REVERSED *
// start and finish are byte offsets
func (w *Wikigraph) ConstructPath(start int32, finish int32, parents map[int32]int32) []int32 {
	path := make([]int32, len(parents)/2)
	for v := to_i(finish); v != -1 && parents[v] != -1; v = parents[v] { // Check parents[v] != -1 for safety.
		path = append(path, v*4) // Return byte offsets
		if v*4 == start {
			break
		}
	}
	return path
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
