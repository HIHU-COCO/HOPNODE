// Purpose: local desktop frontend tooling. Creator: HIHU-COCO. Created: 2026-10-19.
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  clearScreen: false,
  server: { port: 1420, strictPort: true },
  test: { include: ['src/**/*.test.ts'] },
});
