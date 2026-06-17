import react from '@vitejs/plugin-react';
import {defineConfig} from 'vite';

export default defineConfig({
  plugins: [react()],
  build: {
    chunkSizeWarningLimit: 1800,
  },
  server: {
    port: 5173,
    open: true,
    proxy: {
      '/api': 'http://127.0.0.1:8642',
    },
  },
  test: {
    environment: 'node',
  },
});
