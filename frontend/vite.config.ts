import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { fileURLToPath, URL } from 'node:url'
import { readFileSync } from 'node:fs'

// 发布构建会先把 tag 版本写入 wails.json，页面与安装包使用同一版本来源。
const wailsConfig = JSON.parse(readFileSync(new URL('../wails.json', import.meta.url), 'utf-8'))

// https://vitejs.dev/config/
export default defineConfig({
  define: {
    __APP_VERSION__: JSON.stringify(wailsConfig.info?.productVersion ?? '开发版'),
  },
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    hmr: {
      host: 'localhost',
      protocol: 'ws',
    }
  }
})
