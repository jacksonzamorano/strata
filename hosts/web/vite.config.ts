import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import solid from "vite-plugin-solid";
import { viteSingleFile } from "vite-plugin-singlefile";

export default defineConfig({
  plugins: [tailwindcss(), solid(), viteSingleFile()],
  server: {
    port: 5173
  },
  build: {
    target: "esnext",
    minify: false,
    sourcemap: false,
    cssCodeSplit: false,
    assetsInlineLimit: 100000000,
    outDir: "dist",
    emptyOutDir: true,
    rollupOptions: {
      output: {
        inlineDynamicImports: true
      }
    }
  }
});
