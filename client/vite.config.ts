/*
 * Copyright (c) 2023-2025 KCai Technology (https://kcaitech.com). All rights reserved.
 *
 * This file is part of the Vextra project, which is licensed under the AGPL-3.0 license.
 * The full license text can be found in the LICENSE file in the root directory of this source tree.
 *
 * For more information about the AGPL-3.0 license, please visit:
 * https://www.gnu.org/licenses/agpl-3.0.html
 */

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