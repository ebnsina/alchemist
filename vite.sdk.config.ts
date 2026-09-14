import { defineConfig } from 'vite';

// The SDK for hosts that want their own chrome. Shaka stays a dynamic import so it
// lands in its own chunk and is not paid for until a video actually loads.
export default defineConfig({
  build: {
    outDir: 'dist/sdk',
    emptyOutDir: true,
    target: ['es2019', 'chrome70', 'safari12'],
    lib: { entry: 'src/index.ts', formats: ['es'], fileName: () => 'alchemist-player.js' },
    rollupOptions: { output: { chunkFileNames: '[name]-[hash].js' } },
  },
});
