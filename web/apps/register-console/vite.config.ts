import babel from "@rolldown/plugin-babel";
import tailwindcss from "@tailwindcss/vite";
import react, { reactCompilerPreset } from "@vitejs/plugin-react";
import { defineConfig } from "vite";

// The Register embeds the build and serves it with a strict CSP.
export default defineConfig({
  plugins: [react(), babel({ presets: [reactCompilerPreset()] }), tailwindcss()],
  build: {
    // The Register serves each file in assets/ as immutable, so each file name there must have a content hash.
    assetsDir: "assets",
    // No data: URIs: the CSP permits them for images only.
    assetsInlineLimit: 0,
    // All target browsers support modulepreload.
    modulePreload: { polyfill: false },
  },
});
