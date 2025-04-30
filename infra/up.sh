#!/bin/bash

docker stop $(docker ps -aq)

docker remove $(docker ps -aq)

docker compose -f 'docker-compose.yml' up -d --build

docker start $(docker ps -aq)

docker ps -a
