import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  base: '/fauxrpc/',
  build: {
    outDir: 'dashboard_dist',
    emptyOutDir: true,
  },
});
