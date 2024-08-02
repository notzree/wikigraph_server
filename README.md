<div align = "center">
<pre>
__        _____ _  _____ ____ ____      _    ____  _   _
\ \      / /_ _| |/ /_ _/ ___|  _ \    / \  |  _ \| | | |
 \ \ /\ / / | || ' / | | |  _| |_) |  / _ \ | |_) | |_| |
  \ V  V /  | || . \ | | |_| |  _ <  / ___ \|  __/|  _  |
   \_/\_/  |___|_|\_\___\____|_| \_\/_/   \_\_|   |_| |_|
  -------------------------------------------------------
  Golang API server to become the #1 wikipedia speedrunner
</pre>
    
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

</div>

Ever wanted to cheat in your Wikipedia speedruns?
Try Wikigraph, an API to tell you the shortest distance between (almost) any 2 articles on wikipedia.
Outdated data? Create a fresh copy yourself using [wikigraph_script](https://github.com/notzree/wikigraph_script)

# quick links

- [Installation](#installation)
- [Usage example](#usage-example)

  
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
mv downloads/wikipedia_binary_graph.bin /path_to_git_repo/wikigraph
```
Build with docker, this will setup all the database stuff and run the server on port 80.
```sh
docker-compose build && docker-compose up
```
## Usage example


## Implementation details

## Caveats 

## Create your own wikigraph







