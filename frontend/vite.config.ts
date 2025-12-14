import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000, // 前端运行在 3000 端口
    proxy: {
      '/api': {
        target: 'http://localhost:8080', // 转发给 Go 后端
        changeOrigin: true,
      }
    }
  }
})