#!/usr/bin/env bash
# =============================================================================
# sub2api 无感（零停机）蓝绿部署脚本
# =============================================================================
#
# 思路：新版本先以「绿」容器起在备用端口，通过 /health 健康检查后再把
#       nginx upstream 切过去，确认无误再停掉「蓝」（旧）容器。
#       中转站请求在整个过程中始终有可用后端，切换只丢失极少量 in-flight 连接。
#
# 用法（在服务器上执行）：
#   ./bluegreen-deploy.sh build            # 构建新镜像（不切换）
#   ./bluegreen-deploy.sh deploy <tag>     # 起新容器 + 健康检查 + 切换 + 收尾
#   ./bluegreen-deploy.sh rollback         # 切回上一个颜色
#   ./bluegreen-deploy.sh status           # 查看当前颜色与容器状态
#
# 环境变量（可用 .env 覆盖）：
#   APP_NAME       容器名前缀           默认 sub2api
#   IMAGE          镜像名               默认 sub2api
#   NETWORK        docker 网络          默认 sub2api-net
#   BLUE_PORT      蓝端口               默认 18080
#   GREEN_PORT     绿端口               默认 18081
#   NGINX_CONF     nginx upstream 配置   默认 /etc/nginx/conf.d/sub2api-upstream.conf
#   STATE_FILE     记录当前颜色          默认 /opt/sub2api/.bluegreen-state
#   LEGACY_APP_CONTAINERS  首次迁移时需清理的旧应用容器名，逗号分隔
#   HAS_NGINX      1=用 nginx 切换, 0=用 docker 端口直切  默认 1
#
set -euo pipefail

SCRIPT_DIR="$(cd "${BASH_SOURCE[0]%/*}" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# shellcheck disable=SC1091
[[ -f "${SCRIPT_DIR}/.env" ]] && set -a && . "${SCRIPT_DIR}/.env" && set +a

APP_NAME="${APP_NAME:-sub2api}"
IMAGE="${IMAGE:-sub2api}"
NETWORK="${NETWORK:-sub2api-net}"
BLUE_PORT="${BLUE_PORT:-18080}"
GREEN_PORT="${GREEN_PORT:-18081}"
NGINX_CONF="${NGINX_CONF:-/etc/nginx/conf.d/sub2api-upstream.conf}"
STATE_FILE="${STATE_FILE:-/opt/sub2api/.bluegreen-state}"
HAS_NGINX="${HAS_NGINX:-1}"
DATA_VOLUME="${DATA_VOLUME:-sub2api_data}"
ENV_FILE="${ENV_FILE:-${SCRIPT_DIR}/.env}"
HEALTH_TIMEOUT="${HEALTH_TIMEOUT:-120}"
LEGACY_APP_CONTAINERS="${LEGACY_APP_CONTAINERS:-}"

log()  { printf '\033[1;34m[%s]\033[0m %s\n' "$(date +%H:%M:%S)" "$*"; }
warn() { printf '\033[1;33mWARN\033[0m %s\n' "$*" >&2; }
die()  { printf '\033[1;31mERROR\033[0m %s\n' "$*" >&2; exit 1; }

# --- 当前颜色 ----------------------------------------------------------------
current_color() {
  if [[ -f "$STATE_FILE" ]]; then cat "$STATE_FILE"; else echo "none"; fi
}
other_color() {
  [[ "$(current_color)" == "blue" ]] && echo "green" || echo "blue"
}
port_of() {
  [[ "$1" == "blue" ]] && echo "$BLUE_PORT" || echo "$GREEN_PORT"
}
container_of() { echo "${APP_NAME}-$1"; }
tag_file_of() { echo "${STATE_FILE}.$1.tag"; }

save_color() { mkdir -p "$(dirname "$STATE_FILE")"; printf '%s' "$1" > "$STATE_FILE"; }
save_tag() {
  local color="$1" tag="$2"
  mkdir -p "$(dirname "$(tag_file_of "$color")")"
  printf '%s' "$tag" > "$(tag_file_of "$color")"
}
tag_of() { [[ -f "$(tag_file_of "$1")" ]] && cat "$(tag_file_of "$1")"; }

# --- 构建 --------------------------------------------------------------------
do_build() {
  local tag="${1:-0.2.18}"
  log "构建镜像 ${IMAGE}:${tag}"
  docker build \
    -t "${IMAGE}:${tag}" \
    -t "${IMAGE}:latest" \
    --build-arg GOPROXY="${GOPROXY:-https://goproxy.cn,direct}" \
    --build-arg GOSUMDB="${GOSUMDB:-sum.golang.google.cn}" \
    -f "${REPO_ROOT}/Dockerfile" \
    "${REPO_ROOT}"
  log "镜像构建完成：${IMAGE}:${tag}"
}

# --- 启动一个颜色 ------------------------------------------------------------
start_color() {
  local color="$1" tag="$2"
  local name port
  name="$(container_of "$color")"
  port="$(port_of "$color")"

  log "启动 ${color} 容器 ${name} → 宿主端口 ${port}"
  docker rm -f "$name" >/dev/null 2>&1 || true

  local -a env_args=(--env-file "$ENV_FILE")
  [[ -f "$ENV_FILE" ]] || { warn "未找到 $ENV_FILE，改用 -e 透传"; env_args=(-e AUTO_SETUP=true); }

  docker run -d \
    --name "$name" \
    --label com.jokerdeck.sub2api.managed=true \
    --label "com.jokerdeck.sub2api.app=${APP_NAME}" \
    --label com.jokerdeck.sub2api.role=app \
    --restart unless-stopped \
    --network "$NETWORK" \
    -p "127.0.0.1:${port}:8080" \
    -v "${DATA_VOLUME}:/app/data" \
    --security-opt no-new-privileges:true \
    --ulimit nofile=100000:100000 \
    "${env_args[@]}" \
    -e SERVER_PORT=8080 \
    -e SERVER_HOST=0.0.0.0 \
    "${IMAGE}:${tag}"

  log "等待 ${name} 通过健康检查（最长 ${HEALTH_TIMEOUT}s）..."
  local waited=0
  until docker exec "$name" wget -q -T 3 -O /dev/null "http://localhost:8080/health" 2>/dev/null; do
    if (( waited >= HEALTH_TIMEOUT )); then
      warn "健康检查超时，输出最后 40 行日志："
      docker logs --tail 40 "$name" >&2 || true
      die "${name} 未通过健康检查，部署中止（旧版本仍在服务）"
    fi
    if [[ "$(docker inspect -f '{{.State.Running}}' "$name" 2>/dev/null)" != "true" ]]; then
      docker logs --tail 60 "$name" >&2 || true
      die "${name} 已退出，部署中止（旧版本仍在服务）"
    fi
    sleep 3; waited=$((waited+3))
    (( waited % 15 == 0 )) && log "  ...已等待 ${waited}s"
  done
  save_tag "$color" "$tag"
  log "${name} 健康检查通过 ✅"
}

# --- 切换流量 ----------------------------------------------------------------
switch_traffic() {
  local color="$1"
  local port
  port="$(port_of "$color")"

  if [[ "$HAS_NGINX" == "1" ]]; then
    log "切换 nginx upstream → 127.0.0.1:${port}"
    [[ -w "$(dirname "$NGINX_CONF")" ]] || die "无法写入 $NGINX_CONF（需要 root）"
    [[ -f "$NGINX_CONF" ]] && cp "$NGINX_CONF" "${NGINX_CONF}.bak.$(date +%s)"
    cat > "$NGINX_CONF" <<EOF
# 由 bluegreen-deploy.sh 生成于 $(date -Iseconds)
# 当前活动颜色：${color}
upstream sub2api_backend {
    server 127.0.0.1:${port} max_fails=3 fail_timeout=10s;
    keepalive 64;
}
EOF
    nginx -t || die "nginx 配置校验失败，已保留旧配置"
    nginx -s reload
    log "nginx 已重载，流量切到 ${color}"
  else
    log "HAS_NGINX=0：请自行确认 8080 已指向 127.0.0.1:${port}"
  fi
  save_color "$color"
}

# --- 清理旧应用容器 ----------------------------------------------------------
remove_app_container() {
  local name="$1"
  if docker ps -a --format '{{.Names}}' | grep -Fqx -- "$name"; then
    log "停止旧应用容器 ${name}"
    # 给 in-flight 请求一点排空时间，然后删除容器本身。
    docker stop -t 30 "$name" >/dev/null || true
    docker rm "$name" >/dev/null
    log "旧应用容器已删除：${name}"
  fi
}

cleanup_old_app_containers() {
  local keep="$1" old name legacy
  old="$([[ "$keep" == "blue" ]] && echo green || echo blue)"
  remove_app_container "$(container_of "$old")"

  # 只清理本脚本创建并打过标签的应用容器，不触碰数据库、Redis、Nginx。
  while IFS= read -r name; do
    [[ -z "$name" || "$name" == "$(container_of "$keep")" ]] && continue
    remove_app_container "$name"
  done < <(docker ps -a \
    --filter 'label=com.jokerdeck.sub2api.managed=true' \
    --filter "label=com.jokerdeck.sub2api.app=${APP_NAME}" \
    --filter 'label=com.jokerdeck.sub2api.role=app' \
    --format '{{.Names}}')

  # 首次迁移可显式传入旧应用容器名。必须是 APP_NAME 本身或使用 APP_NAME- 前缀。
  local -a legacy_names=()
  IFS=',' read -r -a legacy_names <<< "$LEGACY_APP_CONTAINERS"
  for legacy in "${legacy_names[@]}"; do
    legacy="${legacy#"${legacy%%[![:space:]]*}"}"
    legacy="${legacy%"${legacy##*[![:space:]]}"}"
    [[ -z "$legacy" ]] && continue
    [[ "$legacy" == "$APP_NAME" || "$legacy" == "${APP_NAME}-"* ]] || die "拒绝清理非应用容器：${legacy}"
    [[ "$legacy" == "$(container_of "$keep")" ]] && continue
    remove_app_container "$legacy"
  done
}

validate_legacy_app_containers() {
  local legacy
  local -a legacy_names=()
  IFS=',' read -r -a legacy_names <<< "$LEGACY_APP_CONTAINERS"
  for legacy in "${legacy_names[@]}"; do
    legacy="${legacy#"${legacy%%[![:space:]]*}"}"
    legacy="${legacy%"${legacy##*[![:space:]]}"}"
    [[ -z "$legacy" ]] && continue
    [[ "$legacy" == "$APP_NAME" || "$legacy" == "${APP_NAME}-"* ]] || die "拒绝清理非应用容器：${legacy}"
  done
}

# --- 命令 --------------------------------------------------------------------
cmd_deploy() {
  local tag="${1:-0.2.18}"
  local cur next
  cur="$(current_color)"
  next="$(other_color)"

  validate_legacy_app_containers
  docker image inspect "${IMAGE}:${tag}" >/dev/null 2>&1 || die "镜像 ${IMAGE}:${tag} 不存在，请先执行 build"

  log "当前颜色=${cur}  目标颜色=${next}"
  start_color "$next" "$tag"       # 1. 起新版本
  switch_traffic "$next"           # 2. 健康后切流量
  cleanup_old_app_containers "$next" # 3. 停止并删除旧应用容器

  log "部署完成 🎉 当前活动颜色=${next}（端口 $(port_of "$next")）"
  log "旧镜像仍保留；需要回滚时可执行： $0 rollback"
}

cmd_rollback() {
  local cur prev
  cur="$(current_color)"
  prev="$([[ "$cur" == "blue" ]] && echo green || echo blue)"
  local name
  name="$(container_of "$prev")"

  validate_legacy_app_containers
  log "回滚：${cur} → ${prev}"
  if docker ps -a --format '{{.Names}}' | grep -Fqx -- "$name"; then
    docker start "$name" >/dev/null
    start_color_wait_ready "$name" || die "回滚目标未就绪"
  else
    local tag
    tag="$(tag_of "$prev")"
    [[ -n "$tag" ]] || die "找不到 ${prev} 的历史镜像标签，无法自动回滚"
    docker image inspect "${IMAGE}:${tag}" >/dev/null 2>&1 || die "历史镜像 ${IMAGE}:${tag} 不存在，无法自动回滚"
    start_color "$prev" "$tag"
  fi
  switch_traffic "$prev"
  cleanup_old_app_containers "$prev"
  log "已回滚到 ${prev}"
}

start_color_wait_ready() {
  local name="$1" waited=0
  until docker exec "$name" wget -q -T 3 -O /dev/null "http://localhost:8080/health" 2>/dev/null; do
    (( waited >= HEALTH_TIMEOUT )) && return 1
    sleep 3; waited=$((waited+3))
  done
  return 0
}

cmd_status() {
  log "当前颜色： $(current_color)"
  for color in blue green; do
    [[ -f "$(tag_file_of "$color")" ]] && printf '%s 镜像标签：%s\n' "$color" "$(tag_of "$color")"
  done
  echo
  docker ps -a --filter "name=${APP_NAME}-" \
    --format 'table {{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}'
  echo
  [[ -f "$NGINX_CONF" ]] && { echo "--- $NGINX_CONF ---"; cat "$NGINX_CONF"; }
}

case "${1:-}" in
  build)    shift; do_build "${1:-0.2.18}" ;;
  deploy)   shift; cmd_deploy "${1:-0.2.18}" ;;
  rollback) cmd_rollback ;;
  status)   cmd_status ;;
  *)
    cat <<EOF
用法: $0 <命令> [tag]

  build [tag]     构建镜像（默认 tag 0.2.18）
  deploy [tag]    零停机发布到备用颜色
  rollback        回滚到上一个颜色
  status          查看状态

首次切换时如需删除旧版应用容器，可在 deploy/.env 设置：
  LEGACY_APP_CONTAINERS=sub2api-next,sub2api-v0216-nodes-20261015
仅填写旧应用容器名（必须是 ${APP_NAME} 或以 ${APP_NAME}- 开头）；不会清理数据库、Redis、代理容器或镜像。

环境变量见脚本头部注释。
EOF
    exit 1 ;;
esac
