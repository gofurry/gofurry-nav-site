import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  build: {
    outDir: '../internal/transport/http/webui/dist',
    emptyOutDir: true,
  },
  server: {
    port: 5177,
  },
})
