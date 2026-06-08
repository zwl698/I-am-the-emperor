#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

export GOCACHE="${GOCACHE:-$ROOT_DIR/.gocache}"
export PORT="${PORT:-8080}"

URL="http://localhost:${PORT}"

mkdir -p "$GOCACHE"

open_frontend() {
  if command -v open >/dev/null 2>&1; then
    open "$URL" >/dev/null 2>&1 || true
    return
  fi
  if command -v xdg-open >/dev/null 2>&1; then
    xdg-open "$URL" >/dev/null 2>&1 || true
    return
  fi
  if command -v start >/dev/null 2>&1; then
    start "$URL" >/dev/null 2>&1 || true
  fi
}

wait_for_server() {
  for _ in {1..60}; do
    if curl -fsS "$URL" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.2
  done
  return 1
}

echo "启动《我是皇帝》..."
echo "前端页面: $URL"
echo "后端 API: ${URL}/api"
echo "按 Ctrl+C 停止服务。"
echo

go run ./backend/cmd/emperor &
SERVER_PID=$!

cleanup() {
  kill "$SERVER_PID" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

if wait_for_server; then
  open_frontend
else
  echo "服务启动较慢或端口不可用，请手动打开：$URL"
fi

wait "$SERVER_PID"
