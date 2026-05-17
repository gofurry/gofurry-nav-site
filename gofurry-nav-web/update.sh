#!/usr/bin/env bash
set -euo pipefail

compose_file="docker-compose.prod.yml"

docker compose -f "$compose_file" config -q
docker compose -f "$compose_file" up -d --build --remove-orphans
docker compose -f "$compose_file" ps
