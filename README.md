# Wikigraph's server
## About
Have you ever wanted to measure the arbitrary distance between 2 things? Do you want to become the best at [Wikipedia Speedruns](https://wikispeedruns.com/)? Well look no further.
Wikigraph is a pathfinder service that computes the number of links it takes to get from 1 Wikipedia article to another using a compressed binary graph (inspired by [Tristan Hume](https://github.com/trishume/wikicrush))

This repo contains the code to run and deploy the Wikigraph Pathfinding microservice (which also contains a rate limiter microservice because I'm broke and aws is expensive). 
It exposes a rest api via the rate limiter, which communicates with the pathfinder service via gRPC + ProtoBuf. I wrote the rate limiter myself and it's an implementation of the [Token Bucket Algo](https://en.wikipedia.org/wiki/Token_bucket) using Redis.
The Pathfinder currently runs a BFS implemented in Go and traverses the binary graph created in this [repository](https://github.com/notzree/wikigraph_script). It's able to compute the shortest path between 2 wikipedia pages sub millisecond

Photo of it working
<img width="1184" alt="image" src="https://github.com/notzree/wikigraph_server/assets/118649285/f882e24e-6b74-4e8a-9729-4ba59d17ad70">
