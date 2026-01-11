#!/usr/bin/env bash
set -euo pipefail

ACTION=${1:-up}
SERVICE=${2:-simple-resume}
TAIL=${TAIL:-100}

case "$ACTION" in
  build)
    docker build -t simple-resume .
    ;;
  up)
    docker compose up -d --build
    ;;
  upf|up-follow)
    docker compose up -d --build
    docker compose logs -f --tail=$TAIL $SERVICE
    ;;
  down)
    docker compose down
    ;;
  logs)
    docker compose logs -f --tail=$TAIL $SERVICE
    ;;
  *)
    echo "usage: $0 [build|up|upf|down|logs] [service]" ; exit 1 ;;
esac
