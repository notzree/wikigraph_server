<div align = "center">
<pre>
__        _____ _  _____ ____ ____      _    ____  _   _
\ \      / /_ _| |/ /_ _/ ___|  _ \    / \  |  _ \| | | |
 \ \ /\ / / | || ' / | | |  _| |_) |  / _ \ | |_) | |_| |
  \ V  V /  | || . \ | | |_| |  _ <  / ___ \|  __/|  _  |
   \_/\_/  |___|_|\_\___\____|_| \_\/_/   \_\_|   |_| |_|
  -------------------------------------------------------
  Golang API server to concurrently search wikipedia link graph
</pre>

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>
Try it out: https://wikigraph-client.vercel.app/

Ever wanted to cheat in your Wikipedia speedruns?
Try Wikigraph, an API to tell you the shortest distance between (almost) any 2 articles on wikipedia.
Outdated data? Create a fresh copy yourself using [wikigraph_script](https://github.com/notzree/wikigraph_script)

# Skip to:

- [performance](#Benchmarks)
- [implementation detail](#Implementation-details)


## Installation
Make sure you have docker installed and working. \
Clone this repo
```sh
git clone something
```
Download the Binary graph or create it (see below)
Google drive link: [Graph Link](https://drive.google.com/file/d/1GDBSYfmq6aJpdc_6L5Q5RVJDWMi0vTiK/view?usp=sharing) \
Be sure to move it to the root of the git repo
```sh
mv downloads/wikipedia_binary_graph.bin /path_to_git_repo
```
Download the database dumps:
Google drive link: [Dump Link](https://drive.google.com/file/d/10kCHg-DeNeQ36ASptNBYBzsh90-opFnS/view?usp=drive_link)
```sh
cd wikigraph_server_repo_path
mkdir database_infra
cd database_infra && mkdir initdb
mv downloads/wikigraph_dumps.sql /path_to_git_repo/database_infra/initdb
```
Build with docker, this will setup all the database stuff and run the server on port 80.
```sh
docker-compose build && docker-compose up
```
## Usage example
On other branches you may find working implementations of a gRPC service, for the sake of time I've only included interactions with the REST api.
The both must be a valid wikipedia article names in lowercase (more on this in the caveats section).
Invalid or misstyped article names will cause a 500 error.
The graph and database dumps found on the drive were created Feburary 2024. If you need a newer version, you can use [wikigraph_script](https://github.com/notzree/wikigraph_script)
### Find a path sequentially
```sh
curl  -X POST \
  'http://localhost:8080/search/sequential' \
  --header 'Accept: */*' \
  --header 'User-Agent: Thunder Client (https://www.thunderclient.com)' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "from": "university of waterloo",
  "to": "outer mongolia"
}'
```
### Find a path concurrently
```sh
curl  -X POST \
  'http://localhost:8080/search/concurrent' \
  --header 'Accept: */*' \
  --header 'User-Agent: Thunder Client (https://www.thunderclient.com)' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "from": "university of waterloo",
  "to": "outer mongolia"
}'
```
## Implementation details
For a deeper explanation of the graph format, checkout: [wikigraph_script](https://github.com/notzree/wikigraph_script) or [wikicrush](https://github.com/trishume/wikicrush) \
When the server starts, it loads the entire binary graph into memory as an array of int32 (around 2gbs). This represents over 61,000,000 articles!
The inputs are first converted into byte offsets using a postgres database, this is to save on runtime memory requirements.
### Sequential search
The sequential search runs a simple BFS algorithim to construct a predecessor map of the byteoffsets, then traverses the map to create the path. The algorithim takes advantage of the byteoffsets stored as values to quickly traverse the graph.
### Concurrent search
The concurrent search mode uses 2 goroutines to run bi-directional bfs starting from both the start node and the end node towards each other. This significantly reduces the time it takes to find another node. The basic principle is as follows:
```
If we know A->F
and Z->F
Then we know A->Z = A->F + reverse(Z->F)
```
As the 2 goroutines travserse and discover more nodes, they communicate these nodes through a buffered chan of size 1 to prevent them from blocking each other.
```go
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
```
 A third goroutine will keep track of visited nodes and when it deteces a middle node that both goroutines have detected, signals to exist and proceeds to construct the path.
While this approach is significantly faster, it does not guarantee the shortest path nor does it guarantee consistency between runs. The path may be different each time as the concurrent system is not necessarily deterministic. but if your trying to win a speedrun, the trade off may be acceptable.

## Performance
Real life performance is oftentimes in favour of the Sequential approach. This is because the `FindPathConcurrent` and `FindPathSequential` functions return an array of byteoffsets, which we need to query against the database to convert back to words. Since the sequential approach guarantees shortest path, it often beats the concurrent approach by 0.5s. \

However in scenarios where the 2 nodes are particularily far apart, such as with this api request:
```sh
curl  -X POST \
  'http://localhost:8080/search/concurrent' \
  --header 'Accept: */*' \
  --header 'User-Agent: Thunder Client (https://www.thunderclient.com)' \
  --header 'Content-Type: application/json' \
  --data-raw '{
  "from": "deciduous",
  "to": "anthropic"
}'
```
You will notice that the conccurrent approach will give you runtimes of ~3 seconds, whereas the sequential approach can take up to 15 seconds!
Benchmarks can provide deeper insights into the performance gains of concurrency

### Benchmarks

An issue I ran into when benchmarking was finding balanced testcases, as the benefits of the concurrent bi-directional BFS would fluctuate depending on the actual distance between the inputs provided. I believe the data below is good enough to at least indicate that there IS a performance gain to the concurrent approach, however more benchmarking is needed to accurately quantify the benefit. One possible approach is to find pairs of nodes with increasing number of links, and then graph the performance increase to quantify the percent gain per node.
<p float="left">
<div>
<h2>Concurrent Favoured Benchmark | 189% Faster</h2>

 ```sh
goos: darwin
goarch: arm64
pkg: github.com/notzree/wikigraph_server/benchmark
BenchmarkSequentialSearch-8            1        1260669208 ns/op
BenchmarkConcurrentSearch-8           42          35510332 ns/op
PASS
ok      github.com/notzree/wikigraph_server/benchmark   9.449s
 ```

</div>
<div>
<h2>Sequential Favoured Benchmark | 78% Faster</h2>

 ```sh
 goos: darwin
goarch: arm64
pkg: github.com/notzree/wikigraph_server/benchmark
BenchmarkSequentialSearch-8            2         630494688 ns/op
BenchmarkConcurrentSearch-8            4         275946031 ns/op
PASS
ok      github.com/notzree/wikigraph_server/benchmark   11.795s
 ```
</div>
</p>

## Caveats
I was not able to successfully parse all links and as such this graph is not 100% fully complete. I ran into an issue differentiating capitalized and lowercased pages. For example, the programming language ALGOL and the star Algol are differentiated by the casing. This works fine as long as links from other pages that references these pages obey the same capitalization convention, this wasn't the case. I kept running into casing issues resulting in duplicate key errors or not resulting in entries being found in the database. I do intend on polishing this in the future (maybe on another work term).
