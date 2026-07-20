import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  clearScreen: false,
  server: {
    port: 34115,
    strictPort: false,
    host: "127.0.0.1",
  },
  build: {
    target: "es2020",
    minify: "esbuild",
    sourcemap: false,
  },
});
