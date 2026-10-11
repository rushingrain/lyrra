import React from 'react'
import { createRoot } from 'react-dom/client'
import './style.css'
import App from './App'

// 前端 console 转发到后端日志：wails runtime.Log* 路由到 options.Logger，
// 与后端 slog 日志落入同一文件 logs/lyrra.log。
// runtime 缺失（如浏览器调试环境）时静默跳过，保留原 console 输出。
const rt = (window as { runtime?: Record<string, (...a: unknown[]) => unknown> }).runtime
if (rt && typeof rt.LogDebug === 'function') {
    const fmt = (a: unknown): string => {
        if (a instanceof Error) return a.stack ?? a.message
        if (typeof a === 'string') return a
        try {
            return JSON.stringify(a)
        } catch {
            return String(a)
        }
    }
    const route = (method: string, args: unknown[]) => {
        const line = args.map(fmt).join(' ')
        try {
            rt[method](line)
        } catch {
            // 转发失败不打断前端
        }
    }
    const orig = { debug: console.debug, info: console.info, warn: console.warn, error: console.error }
    console.debug = (...args) => { route('LogDebug', args); orig.debug(...args) }
    console.info = (...args) => { route('LogInfo', args); orig.info(...args) }
    console.warn = (...args) => { route('LogWarning', args); orig.warn(...args) }
    console.error = (...args) => { route('LogError', args); orig.error(...args) }
}

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <React.StrictMode>
        <App />
    </React.StrictMode>
)
