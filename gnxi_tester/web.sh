#!/bin/bash

# Shallow clone gnxi 
git clone \
    --depth 1 \
    https://github.com/google/gnxi.git
cd gnxi/gnxi_tester

docker-compose up -d && 
    echo "Web UI running on http://localhost:4200. Type 'docker ps' to see more info on running containers."
