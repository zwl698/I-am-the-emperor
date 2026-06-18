#!/usr/bin/env bash
#
# start.sh — 同时启动 new_sanguoby 的后端 (Go) 与前端 (Vite)。
#
# 用法:
#   ./start.sh            # 启动前后端
#   ./start.sh --backend  # 仅启动后端
#   ./start.sh --frontend # 仅启动前端
#
# 环境变量 (可选):
#   HOST                 后端监听地址 (默认 127.0.0.1)
#   PORT                 后端监听端口 (默认 8642)
#   LEGACY_ARCHIVE_PATH  legacy 档案路径 (默认 ../sanguobaye_c-master/src/dat.lib.orig)
#
# 前端启动后会自动在浏览器打开 (见 web/vite.config.ts 的 server.open)。
#
# 按 Ctrl-C 可同时停止前后端。

if [ -z "${BASH_VERSION:-}" ]; then
  exec /usr/bin/env bash "$0" "$@"
fi

set -euo pipefail

# 切换到脚本所在目录 (new_sanguoby)
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT_DIR"

BACKEND_HOST="${HOST:-127.0.0.1}"
BACKEND_PORT="${PORT:-8642}"
# 默认指向同级 sanguobaye_c-master 的原始档案 (从 new_sanguoby 运行时为上一级目录)
DEFAULT_ARCHIVE="$ROOT_DIR/../sanguobaye_c-master/src/dat.lib.orig"
LEGACY_ARCHIVE_PATH="${LEGACY_ARCHIVE_PATH:-$DEFAULT_ARCHIVE}"

# 使用项目内的 Go build cache, 与既有约定保持一致
export GOCACHE="${GOCACHE:-$ROOT_DIR/.gocache}"
export GOFLAGS="${GOFLAGS:--mod=mod}"

RUN_BACKEND=1
RUN_FRONTEND=1
case "${1:-}" in
  --backend)  RUN_FRONTEND=0 ;;
  --frontend) RUN_BACKEND=0 ;;
  "")         ;;
  *) echo "未知参数: $1 (可用: --backend, --frontend)"; exit 1 ;;
esac

PIDS=()
CLEANED_UP=0

cleanup() {
  if [[ "$CLEANED_UP" -eq 1 ]]; then
    return
  fi
  CLEANED_UP=1
  echo ""
  echo "[start] 正在停止服务..."
  for pid in "${PIDS[@]:-}"; do
    if [[ -n "$pid" ]] && kill -0 "$pid" 2>/dev/null; then
      # 结束整个进程组, 确保子进程 (vite/go run) 一并退出
      kill "$pid" 2>/dev/null || true
    fi
  done
  wait 2>/dev/null || true
  echo "[start] 已全部停止。"
}
trap cleanup INT TERM EXIT

color() { printf "\033[%sm%s\033[0m" "$1" "$2"; }

ensure_node_runtime() {
  if command -v node >/dev/null 2>&1; then
    return
  fi

  local candidates=(
    "$HOME/.cache/codex-runtimes/codex-primary-runtime/dependencies/node/bin"
    "$HOME/.local/node/bin"
    "/opt/homebrew/bin"
    "/usr/local/bin"
  )
  local dir
  for dir in "${candidates[@]}"; do
    if [[ -x "$dir/node" ]]; then
      export PATH="$dir:$PATH"
      echo "$(color '35' '[frontend]') 已自动使用 Node: $dir/node"
      return
    fi
  done

  echo "$(color '31' '[frontend]') 未找到 node。请安装 Node.js，或将 node 所在目录加入 PATH 后重试。"
  exit 1
}

ensure_frontend_tools() {
  ensure_node_runtime
  if ! command -v npm >/dev/null 2>&1 && [[ -x "$HOME/.local/bin/npm" ]]; then
    export PATH="$HOME/.local/bin:$PATH"
  fi
  if ! command -v npm >/dev/null 2>&1; then
    echo "$(color '31' '[frontend]') 未找到 npm。请安装 npm，或将 npm 所在目录加入 PATH 后重试。"
    exit 1
  fi
}

ensure_go_runtime() {
  if command -v go >/dev/null 2>&1 && go list encoding/binary >/dev/null 2>&1; then
    return
  fi

  local candidates=(
    "$HOME/go/go1.26.1/bin"
    "$HOME/sdk/go1.26.1/bin"
    "/opt/homebrew/opt/go@1.26/bin"
    "/opt/homebrew/opt/go/bin"
    "/usr/local/go/bin"
  )
  local dir
  for dir in "${candidates[@]}"; do
    if [[ -x "$dir/go" ]] && "$dir/go" list encoding/binary >/dev/null 2>&1; then
      export PATH="$dir:$PATH"
      echo "$(color '36' '[backend]') 已自动使用 Go: $dir/go"
      return
    fi
  done

  echo "$(color '31' '[backend]') 未找到可用 Go。请安装 Go，或将可用 go 所在目录加入 PATH 后重试。"
  exit 1
}

# 检测端口是否被占用, 被占用则提示并退出, 避免 go run 静默失败。
ensure_port_free() {
  local port="$1" label="$2"
  if lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
    echo "$(color '31' '[start]') 端口 $port ($label) 已被占用, 请先停止占用进程或设置其他端口:"
    lsof -nP -iTCP:"$port" -sTCP:LISTEN 2>/dev/null | sed 's/^/  /'
    exit 1
  fi
}

start_backend() {
  ensure_go_runtime
  echo "$(color '36' '[backend]') 启动 Go 服务于 http://$BACKEND_HOST:$BACKEND_PORT"
  echo "$(color '36' '[backend]') 档案路径: $LEGACY_ARCHIVE_PATH"
  if [[ ! -f "$LEGACY_ARCHIVE_PATH" ]]; then
    echo "$(color '33' '[backend]') 警告: 未找到 legacy 档案, 后端将回退到内置 seed 数据。"
  fi
  HOST="$BACKEND_HOST" PORT="$BACKEND_PORT" LEGACY_ARCHIVE_PATH="$LEGACY_ARCHIVE_PATH" \
    go run ./backend/cmd/server 2>&1 | sed "s/^/$(color '36' '[backend] ')/" &
  PIDS+=("$!")
}

start_frontend() {
  ensure_frontend_tools
  if [[ ! -d "web/node_modules" ]]; then
    echo "$(color '35' '[frontend]') 未检测到 node_modules, 正在安装依赖..."
    (cd web && npm install)
  fi
  echo "$(color '35' '[frontend]') 启动 Vite 开发服务器于 http://127.0.0.1:5173"
  (cd web && npm run dev) 2>&1 | sed "s/^/$(color '35' '[frontend] ')/" &
  PIDS+=("$!")
}

[[ "$RUN_BACKEND" -eq 1 ]] && ensure_port_free "$BACKEND_PORT" "backend"
[[ "$RUN_FRONTEND" -eq 1 ]] && ensure_port_free 5173 "frontend"

[[ "$RUN_BACKEND" -eq 1 ]] && start_backend
[[ "$RUN_FRONTEND" -eq 1 ]] && start_frontend

echo "[start] 服务已启动, 按 Ctrl-C 停止。"
wait
