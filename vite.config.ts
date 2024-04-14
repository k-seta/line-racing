import { defineConfig } from 'vite'

export default defineConfig({
  server: {
    proxy: {
      "/p1": {
        target: "http://localhost:8001",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/p1/, ''),
      },
      "/p2": {
        target: "http://localhost:8002",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/p2/, ''),
      },
      "/p3": {
        target: "http://localhost:8003",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/p3/, ''),
      },
      "/p4": {
        target: "http://localhost:8004",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/p4/, ''),
      },
      "/p5": {
        target: "http://localhost:8005",
        changeOrigin: true,
        rewrite: (path) => path.replace(/^\/p5/, ''),
      }
    },
  },
});
