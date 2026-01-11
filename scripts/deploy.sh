#!/usr/bin/env bash
set -euo pipefail

ACTION=${1:-start}
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="${ROOT_DIR}/bin"
BIN_FILE="${BIN_DIR}/resume-to-job"
RUNTIME_DIR="${ROOT_DIR}/.runtime"
PID_FILE="${RUNTIME_DIR}/server.pid"
LOG_FILE="${RUNTIME_DIR}/server.out"

export GIN_MODE=${GIN_MODE:-release}
export PORT=${PORT:-8080}

build() {
  mkdir -p "${BIN_DIR}" "${RUNTIME_DIR}"
  cd "${ROOT_DIR}"
  go mod tidy
  go build -o "${BIN_FILE}" "${ROOT_DIR}/main.go"
}

start() {
  if [ -f "${PID_FILE}" ] && kill -0 "$(cat "${PID_FILE}")" 2>/dev/null; then
    echo "running: $(cat "${PID_FILE}")"
    exit 0
  fi
  build
  cd "${ROOT_DIR}"
  nohup "${BIN_FILE}" >"${LOG_FILE}" 2>&1 &
  echo $! >"${PID_FILE}"
  echo "started: $(cat "${PID_FILE}")"
}

stop() {
  if [ -f "${PID_FILE}" ] && kill -0 "$(cat "${PID_FILE}")" 2>/dev/null; then
    kill "$(cat "${PID_FILE}")" || true
    rm -f "${PID_FILE}"
    echo "stopped"
  else
    echo "not running"
  fi
}

status() {
  if [ -f "${PID_FILE}" ] && kill -0 "$(cat "${PID_FILE}")" 2>/dev/null; then
    echo "running: $(cat "${PID_FILE}")"
  else
    echo "stopped"
  fi
}

restart() {
  stop || true
  start
}

case "${ACTION}" in
  start) start ;;
  stop) stop ;;
  restart) restart ;;
  status) status ;;
  build) build ;;
  *) echo "usage: $0 [start|stop|restart|status|build]" ; exit 1 ;;
esac

