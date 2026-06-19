import react from '@vitejs/plugin-react';
import {defineConfig} from 'vite';

const frontendPort = Number.parseInt(process.env.FRONTEND_PORT ?? '5173', 10);
const backendTarget = process.env.VITE_PROXY_TARGET ?? `http://${process.env.HOST ?? '127.0.0.1'}:${process.env.PORT ?? '8642'}`;
const openBrowser = process.env.BROWSER === 'none' || process.env.VITE_OPEN === 'false' ? false : true;

export default defineConfig({
  plugins: [react()],
  build: {
    chunkSizeWarningLimit: 1800,
  },
  server: {
    port: Number.isFinite(frontendPort) ? frontendPort : 5173,
    open: openBrowser,
    proxy: {
      '/api': backendTarget,
    },
  },
  test: {
    environment: 'node',
  },
});
