import tailwindcss from '@tailwindcss/vite'
import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'

export default defineConfig({
  base: '/admin/',
  plugins: [vue(), tailwindcss()],
  build: {
    outDir: '../internal/web/dist',
    emptyOutDir: true,
  },
  server: {
    host: '127.0.0.1',
    port: 5179,
    proxy: {
      '/api': 'http://127.0.0.1:8080',
    },
  },
})
