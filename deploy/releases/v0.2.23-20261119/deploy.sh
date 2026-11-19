#!/usr/bin/env bash
set -Eeuo pipefail
umask 077

stage=/opt/sub2api/staging/v0.2.23-20261119
backup=/opt/sub2api/backups/pre-v0.2.23-20261119
runner="$stage/deploy.py"

die() {
  printf 'Deployment stopped: %s\n' "$*" >&2
  exit 1
}

[[ -f "$runner" && -f "$stage/rollback.py" && -f "$stage/postflight.py" ]] || die 'release scripts are not staged'
[[ -f "$stage/sub2api.gz" && -f "$stage/SHA256SUMS" && -f "$stage/migrations.json" ]] || die 'build bundle is incomplete'
[[ -f "$backup/active-inspect.json" && -f "$backup/nginx.conf" && -f "$backup/migrations.json" ]] || die 'pre-deploy backup is missing'
[[ -f "$stage/rehearsal.json" ]] || die 'isolated database restore/migration rehearsal is missing'

switched=0
rollback_if_needed() {
  rc=$?
  if (( rc != 0 )); then
    if (( switched )); then
      printf 'A post-switch step failed; running guarded rollback.\n' >&2
      python3 "$stage/rollback.py" || printf 'Automatic rollback did not complete; keep both app containers running and inspect Nginx.\n' >&2
    else
      printf 'Failure occurred before traffic switch; production upstream was not changed.\n' >&2
    fi
  fi
  exit "$rc"
}
trap rollback_if_needed ERR

python3 "$runner" start
python3 "$runner" verify
python3 "$runner" assets
python3 "$runner" switch
switched=1
python3 "$runner" observe 600
python3 "$stage/postflight.py"
python3 "$runner" finalize
python3 "$runner" retire

trap - ERR
printf 'v0.2.23 deployment completed. Evidence: %s/final-report.json\n' "$stage"
