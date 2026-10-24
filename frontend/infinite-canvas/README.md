# Infinite Canvas frontend

This frontend is based on `basketikun/infinite-canvas` v0.20.0 and is built
into `frontend/public/infinite-canvas` for same-origin embedding in Sub2API.

Build it from the repository's `frontend` directory with `pnpm run build:canvas`.
The normal `pnpm run build` runs this build before building the Sub2API frontend.

Model channels are configured by each user in the browser. The canvas does not
import account API keys from Sub2API; generation requires a saved channel with
an HTTP(S) base URL, API key, and selected model. Channel settings are stored in
the browser's local storage.
