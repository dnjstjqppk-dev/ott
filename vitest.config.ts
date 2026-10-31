import { defineConfig } from 'vitest/config';
import { resolve } from 'node:path';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    include: ['**/*.spec.ts', '**/*.test.ts'],
    exclude: ['**/node_modules/**', '**/.nuxt/**', '**/.output/**']
  },
  resolve: {
    alias: {
      '~': resolve(__dirname, './clients/web-user'),
      '@': resolve(__dirname, './clients/web-user')
    }
  }
});
