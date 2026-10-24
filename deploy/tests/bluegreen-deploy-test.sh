#!/bin/bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/sub2api-bluegreen-test.XXXXXX")"
cleanup() { rm -rf "$tmp_dir"; }
trap cleanup EXIT HUP INT TERM

fail() {
  printf 'blue-green deployment test failed: %s\n' "$1" >&2
  exit 1
}

assert_container() {
  grep -Fqx -- "$1" "$FAKE_DOCKER_STATE/containers" || fail "expected container $1"
}

assert_no_container() {
  if grep -Fqx -- "$1" "$FAKE_DOCKER_STATE/containers"; then
    fail "unexpected container $1"
  fi
}

mkdir -p "$tmp_dir/bin" "$tmp_dir/state"
cp "$repo_root/deploy/tests/fixtures/bin/docker" "$tmp_dir/bin/docker"
cp "$repo_root/deploy/tests/fixtures/bin/nginx" "$tmp_dir/bin/nginx"
chmod +x "$tmp_dir/bin/docker" "$tmp_dir/bin/nginx"

export FAKE_DOCKER_STATE="$tmp_dir/state"
export FAKE_DOCKER_EVENTS="$tmp_dir/events"
export PATH="$tmp_dir/bin:$PATH"
export APP_NAME=sub2api
export IMAGE=sub2api
export NETWORK=sub2api-net
export STATE_FILE="$tmp_dir/state/active-color"
export NGINX_CONF="$tmp_dir/nginx/sub2api-upstream.conf"
export ENV_FILE="$tmp_dir/deploy.env"
export HEALTH_TIMEOUT=0
export LEGACY_APP_CONTAINERS=sub2api-old-release

mkdir -p "$(dirname "$NGINX_CONF")"
touch "$ENV_FILE"
printf 'blue\n' > "$STATE_FILE"
printf 'v1\n' > "$STATE_FILE.blue.tag"
printf 'sub2api-blue\nsub2api-canary\nsub2api-old-release\nsub2api-postgres\nedge-nginx\n' > "$FAKE_DOCKER_STATE/containers"
printf 'sub2api-blue\nsub2api-canary\nsub2api-old-release\nsub2api-postgres\nedge-nginx\n' > "$FAKE_DOCKER_STATE/running"
printf 'sub2api-canary\n' > "$FAKE_DOCKER_STATE/managed"
: > "$FAKE_DOCKER_EVENTS"

bash "$repo_root/deploy/bluegreen-deploy.sh" deploy v2 >/dev/null

assert_container sub2api-green
assert_no_container sub2api-blue
assert_no_container sub2api-canary
assert_no_container sub2api-old-release
assert_container sub2api-postgres
assert_container edge-nginx
grep -Fqx 'green' "$STATE_FILE" || fail 'active color was not saved'
grep -Fqx 'v2' "$STATE_FILE.green.tag" || fail 'new image tag was not saved'

reload_line=$(grep -nF 'nginx -s reload' "$FAKE_DOCKER_EVENTS" | cut -d: -f1)
old_stop_line=$(grep -nF 'docker stop -t 30 sub2api-blue' "$FAKE_DOCKER_EVENTS" | cut -d: -f1)
old_rm_line=$(grep -nF 'docker rm sub2api-blue' "$FAKE_DOCKER_EVENTS" | cut -d: -f1)
[[ -n "$reload_line" && -n "$old_stop_line" && -n "$old_rm_line" ]] || fail 'missing traffic switch or old-container cleanup event'
(( reload_line < old_stop_line && old_stop_line < old_rm_line )) || fail 'old container cleanup happened before traffic switch or graceful stop'

bash "$repo_root/deploy/bluegreen-deploy.sh" rollback >/dev/null
assert_container sub2api-blue
assert_no_container sub2api-green
assert_container sub2api-postgres
assert_container edge-nginx
grep -Fqx 'blue' "$STATE_FILE" || fail 'rollback did not restore the prior color'
grep -Fq 'docker run sub2api-blue' "$FAKE_DOCKER_EVENTS" || fail 'rollback did not recreate the removed container'
grep -Fq 'docker image inspect sub2api:v1' "$FAKE_DOCKER_EVENTS" || fail 'rollback did not use the saved previous image tag'

printf 'blue-green deployment test passed\n'
