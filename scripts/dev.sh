#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

export GOCACHE="${GOCACHE:-$ROOT_DIR/.gocache}"
export PORT="${PORT:-8080}"

mkdir -p "$GOCACHE"

echo "启动《我是皇帝》开发服务..."
echo "后端 API: http://localhost:${PORT}/api"
echo "前端页面: http://localhost:${PORT}"
echo "按 Ctrl+C 停止服务。"
echo

exec go run ./cmd/emperor
