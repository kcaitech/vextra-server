import { defineConfig } from 'vite'
import { resolve } from 'path'
import dts from 'vite-plugin-dts'

export default defineConfig({
  plugins: [
    dts({
      include: ['.'],
      exclude: ['**/__tests__/**', '**/tests/**'],
      outDir: 'dist/types'
    }),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, '.')
    }
  },
  build: {
    lib: {
      entry: resolve(__dirname, 'index.ts'),
      name: 'KcServerClient',
      fileName: (format) => {
        if (format === 'es') return 'index.js'
        if (format === 'cjs') return 'index.cjs'
        return `index.${format}.js`
      },
      formats: ['es', 'cjs']
    },
    rollupOptions: {
      external: ['axios', 'uuid', 'zod'],
      output: {
        globals: {
          axios: 'axios',
          uuid: 'uuid',
          zod: 'zod'
        }
      }
    },
    sourcemap: true,
    minify: 'terser',
    target: 'esnext'
  }
}) 