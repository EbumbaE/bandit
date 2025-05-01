#!/bin/bash

docker stop $(docker ps -aq)

docker remove $(docker ps -aq)

docker-compose up -d rule_admin -d --build
docker-compose up -d bandit_indexer -d --build
docker-compose up -d rule_diller -d --build
docker-compose up -d rule_analytic -d --build
docker-compose up -d rule_test -d --build
docker-compose up -d grafana -d --build

docker ps -a --format "table {{.ID}}\t{{.Names}}\t{{.Status}}"
