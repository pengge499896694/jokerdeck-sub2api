import { defineConfig } from "vitest/config";
import { resolve } from "node:path";

export default defineConfig({
    root: resolve(__dirname),
    resolve: { alias: { "@": resolve(__dirname, "src") } },
    esbuild: { jsx: "automatic" },
    test: { environment: "jsdom", include: ["src/**/*.spec.tsx"] },
});
