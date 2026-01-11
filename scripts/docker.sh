#!/usr/bin/env bash
set -euo pipefail

ACTION=${1:-up}

case "$ACTION" in
  build)
    docker build -t simple-resume .
    ;;
  up)
    docker compose up -d --build
    ;;
  down)
    docker compose down
    ;;
  logs)
    docker compose logs -f
    ;;
  *)
    echo "usage: $0 [build|up|down|logs]" ; exit 1 ;;
esac

