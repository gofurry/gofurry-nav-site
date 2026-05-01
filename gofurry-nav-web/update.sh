#!/usr/bin/env bash
set -euo pipefail
docker compose -f docker-compose.prod.yml up -d --build
