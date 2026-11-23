# Jokerdeck v0.2.24 guarded deployment

This release targets the production topology recorded in `DEPLOY_20261112.md`:
the current application container is `sub2api-v0.2.23-20261119`, Nginx is the
`edge-nginx` container, and its config is a single-file bind mount at
`/opt/edge/nginx.conf`.

`deploy.sh` is the server-side phase driver. It does not build or transfer the
release bundle. Before running it, stage the Linux amd64 embedded-frontend
binary, its SHA256 file, runtime pricing resources and manifest, complete
migration manifest/SQL files, `build-info.json`, `rehearsal.json`, and the
release Python scripts under `/opt/sub2api/staging/v0.2.24-20261124/`.
Create the timestamped backup directory and isolated full-database restore
rehearsal before invoking the driver. The driver refuses to proceed if those
gates are absent.

Run on the server as the deployment account with Docker access:

```sh
sudo bash /opt/sub2api/staging/v0.2.24-20261124/deploy.sh
```

The sequence is candidate build and additive migrations, candidate health and
API/assets verification, Nginx config validation and graceful reload, ten
minutes of health/log observation, final verification, Compose state sync, and
old-worker drain. It never restarts Nginx or recreates the production app
container. If a post-switch phase fails, the driver invokes the guarded rollback
and retains both app containers if worker drain is incomplete.

Do not run this release against a topology that differs from the recorded
container names, network, Nginx config path, 11 upstream entries, or public
health-check domains without reviewing and updating the release scripts.
