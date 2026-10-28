# Infinite Canvas frontend

This frontend is based on `basketikun/infinite-canvas` v0.20.0 and is built
into `frontend/public/infinite-canvas` for same-origin embedding in Sub2API.

Build it from the repository's `frontend` directory with `pnpm run build:canvas`.
The normal `pnpm run build` runs this build before building the Sub2API frontend.

Model channels are configured by each user in the browser. The canvas does not
import account API keys from Sub2API; generation requires a saved channel with
an HTTP(S) base URL, API key, and selected model. Channel settings are stored in
the browser's local storage.

Canvas projects, generated media, assets and generation histories are saved in
this browser's IndexedDB (with a localStorage fallback for project state).
They do not automatically follow the Sub2API account to another computer,
browser or origin. The separate server canvas API is not used by this embedded
Infinite Canvas frontend.

To move work between devices, configure the same WebDAV directory and run
“Sync now” on the source device, then on the destination device. Sync includes
projects, assets, image/video histories and referenced media; it does not include
API keys. Export/import configuration separately from the settings dialog to
move channel settings and preferences. Configuration exports include API keys
and WebDAV credentials.

Canvas image generation displays an indeterminate waiting bar and elapsed time.
The stage changes to saving only after the API returns an image. For batches,
the bar counts settled requests and the label counts successfully stored images;
the first completed image stays visible while remaining requests run. Model APIs
do not provide an exact generation percentage or remaining time. The timer is
local to the progress component, so ticking does not rewrite the saved canvas.
