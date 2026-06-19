import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'), // 配置 @ 指向 src
    },
  },
  json: {
    stringify: false,
  },
  build: {
    minify: 'esbuild',
  },
  server: {
    host: '127.0.0.1',
    port: 8888,
  }
})
