#!/usr/bin/env bash
set -euo pipefail

compose_file="docker-compose.prod.yml"

docker_proxy_env_file="/etc/gofurry/nav-web-docker-proxy.env"
if [[ -f "$docker_proxy_env_file" && -r "$docker_proxy_env_file" ]]; then
  # Export plain KEY=value assignments for Docker CLI and Buildx as well.
  set -a
  # shellcheck source=/dev/null
  source "$docker_proxy_env_file"
  set +a
fi

docker compose -f "$compose_file" config -q
docker compose -f "$compose_file" up -d --build --remove-orphans
docker compose -f "$compose_file" ps
