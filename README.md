# Wikigraph's server
## About
Have you ever wanted to measure the arbitrary distance between 2 things? Do you want to become the best at [Wikipedia Speedruns](https://wikispeedruns.com/)? Well look no further.
Wikigraph is a pathfinder service that computes the number of links it takes to get from 1 Wikipedia article to another using a compressed binary graph (inspired by [Tristan Hume](https://github.com/trishume/wikicrush))

This repo contains the code to run and deploy the Wikigraph Pathfinding microservice (which also contains a rate limiter microservice because I'm broke and aws is expensive). 
It exposes a rest api via the rate limiter, which communicates with the pathfinder service via gRPC + ProtoBuf. (still need to break down grpc and rest apis separately)

I wrote the rate limiter myself and it's an implementation of the [Token Bucket Algo](https://en.wikipedia.org/wiki/Token_bucket) using Redis.
The Pathfinder currently runs a BFS implemented in Go and traverses the binary graph created in this [repository](https://github.com/notzree/wikigraph_script). It's able to compute the shortest path between 2 wikipedia pages ~ 1 second with the database and everything running on my local machine. See the deployment section where I talk about real-world performance.

## Using the API
//todo: add some docs
## Deployment 
I ran into multiple issues with excessive memory usage. My first initial prototype running the 2 gRPC services + the database was taking around 3.5-4 Gbs of ram. In an attmept to lower my costs, I moved the database onto supabase (excellent free tier) and then removed the gRPC service entirely and just utilized a rest api. 

After lots of profiling + code optimizations, I lowered my memory usage from 3.5gbs to 1.4 gbs. Due to these cost-saving mechanisms, the actual latency of the API is around 2-4 seconds depending on the query.





