#!/bin/bash

echo " "
echo "Остановка контейнеров:"
docker stop $(docker ps -aq)

echo " "
echo "Удаление контейнеров:"
docker remove $(docker ps -aq)

echo " "
echo "Поднимаем rule-admin:"
docker-compose up -d rule_admin -d --build

echo " "
echo "Поднимаем rule-indexer:"
docker-compose up -d bandit_indexer -d --build

echo " "
echo "Поднимаем rule-diller:"
docker-compose up -d rule_diller -d --build

echo " "
echo "Поднимаем rule-analytic:"
docker-compose up -d rule_analytic -d --build

echo " "
echo "Поднимаем rule-test:"
docker-compose up -d rule_test -d --build

echo " "
echo "Поднимаем grafana:"
docker-compose up -d grafana -d --build

echo " "
echo "Статусы контейнеров:"
docker ps -a --format "table {{.ID}}\t{{.Names}}\t{{.Status}}"
