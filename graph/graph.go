package graph

import (
	"encoding/binary"
	"log"
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
	queue := make([]int32, 0, len(w.graph)/2)
	queue = append(queue, start)
	parents[to_i(start)] = to_i(start)
	pathFound := false
	for len(queue) > 0 {
		current_node := queue[0] // node is a byte offset
		queue = queue[1:]
		if current_node == 0 {
			continue
		}
		if current_node == target {
			pathFound = true
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
	if !pathFound {
		return []int32{}, nil
	}
	path := w.ConstructPath(start, target, parents)
	reverse(path)
	return path, nil
}

func (w *Wikigraph) FindPathConcurrent(start, target int32) ([]int32, error) {
	if start == target {
		return []int32{start}, nil
	}
	// parents are used by respective goroutines to keep track of their internal predecessors
	startParents := make(map[int32]int32)
	startParents[to_i(start)] = to_i(start)
	targetParents := make(map[int32]int32)
	targetParents[to_i(target)] = to_i(start)
	// discovered is a buffered chan of size 1 uesd to communicate discovered nodes to the orchestrating goroutine.
	discovered := make(chan int32, 1)

	startQueue := make([]int32, 0, len(w.graph)/2)
	startQueue = append(startQueue, start)
	targetQueue := make([]int32, 0, len(w.graph)/2)
	targetQueue = append(targetQueue, target)
	var wg sync.WaitGroup
	closeSignal := make(chan struct{})
	resultChan := make(chan int32)

	wg.Add(2)
	//from start searching for target
	go w.bfs(BfsParams{
		searchTarget: target,
		parents:      startParents,
		closeSignal:  closeSignal,
		queue:        startQueue,
		discovered:   discovered,
		wg:           &wg,
	})
	//from target searching for start
	go w.bfs(BfsParams{
		searchTarget: start,
		parents:      targetParents,
		closeSignal:  closeSignal,
		queue:        targetQueue,
		discovered:   discovered,
		wg:           &wg,
	})

	go func() {
		discoveredNodes := make(map[int32]bool)
		for node := range discovered {
			if _, ok := discoveredNodes[node]; ok {
				close(closeSignal)
				resultChan <- node
				return
			}
			discoveredNodes[node] = true
		}
	}()

	wg.Wait()
	//middle is the middle value that both start and target know how to reach
	middle := <-resultChan
	//all goroutines are closed so can access maps safely
	startPath := w.ConstructPath(start, middle, startParents) //middle -> start
	reverse(startPath)                                        // start -> middle
	//remove the middle node
	startPath = startPath[:len(startPath)-1]
	targetPath := w.ConstructPath(target, middle, targetParents) // middle -> target

	//join the two paths
	return append(startPath, targetPath...), nil
}

type BfsParams struct {
	searchTarget int32 //byte offset
	parents      map[int32]int32
	closeSignal  chan struct{}
	queue        []int32
	discovered   chan int32
	wg           *sync.WaitGroup
}

func (w *Wikigraph) bfs(params BfsParams) {
	defer params.wg.Done()
	for len(params.queue) > 0 {
		current_node_offset := params.queue[0]
		params.queue = params.queue[1:]
		if current_node_offset == 0 {
			continue
		}
		if current_node_offset == params.searchTarget {
			params.discovered <- current_node_offset
		}
		for _, link_byte_offset := range w.GetLinks(current_node_offset) {
			if link_byte_offset == 0 {
				continue
			}
			link := to_i(link_byte_offset)
			if _, ok := params.parents[link]; !ok { //communicate undiscovered nodes to the orchestrating goroutine
				params.parents[link] = to_i(current_node_offset)
				params.queue = append(params.queue, link_byte_offset)
				select {
				case params.discovered <- link_byte_offset:
				case <-params.closeSignal:
					return
				}
			}
		}

	}
	log.Default().Println("Queue is empty")
}

// Constructs the path from finish -> start using the parents map * NOT REVERSED *
// start and finish are byte offsets
func (w *Wikigraph) ConstructPath(start int32, finish int32, parents map[int32]int32) []int32 {
	path := make([]int32, 0)
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

func to_i(byte_offset int32) int32 {
	if byte_offset == 0 {
		return 0
	}
	return byte_offset / 4
}

func reverse(s []int32) { //reminder to self that slices are passed by reference
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

// Summary of new approach:

// Each bfs goroutine will send newly discovered nodes to the 3rd goroutine.
//This goroutine will keep a record of all nodes discovered by both bfs goroutines.
//If a node is discovered by both bfs goroutines, then it will signal to both goroutines to stop, it will then exit.
//the main goroutine will be listening for the map of parents from both bfs goroutines on a channel

//each bfs goroutine will be creating it's own map of parents. When the signal to stop is received,
//it will send this parent map back on the channel. The main goroutine will use a select statement to listen for both maps
// It will then construct the paths and join them.
