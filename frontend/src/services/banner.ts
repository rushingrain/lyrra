import type { Banner } from '@/types'
import { mockBanners } from '@/data/mock'
import toast from 'react-hot-toast'
import { GetHomeBanners } from '../../wailsjs/go/banner/BannerService'

/**
 * 首页 banner 相关接口
 *
 * 已接入后端（Wails 绑定 window.go.banner.BannerService.*），
 * 后端调酷狗 /pc/diantai 接口并过滤第三方广告后返回。
 *
 * 关于「只跑前端 npm run dev」：
 * Wails 的 window.go 运行时只在 Wails WebView 中注入，普通浏览器里不存在。
 * 此处做 isBackendAvailable 检测：浏览器环境下自动回退到 mock，
 * 便于纯前端 UI 开发时轮播不空白。
 */

/** Wails 运行时是否可用（真实后端是否存在） */
function isBackendAvailable(): boolean {
  return typeof window !== 'undefined' && Boolean((window as any).go?.banner?.BannerService)
}

/**
 * 缓存有效期(ms)：对齐 /pc/diantai 响应里的 timing_update_interval=3600（服务端建议 1h 刷新）。
 * 首页在导航切换时反复挂载，直接拉接口会重复打后端；
 * 用模块级缓存跨导航存活，1h 后自动失效刷新，兼顾新鲜度。
 */
const CACHE_TTL_MS = 60 * 60 * 1000

// 模块级缓存：仅存在于 app 窗口生命周期；请求失败不入缓存，下次挂载自动重试。
let cachedBanners: Banner[] | null = null
let cachedAt = 0
// 进行中的请求去重：首次并发调用共用同一 promise，避免重复拉取
let inflight: Promise<Banner[]> | null = null

/**
 * 获取首页轮播 banner
 * 后端返回已过滤广告的列表，imageUrl 为 https 图床地址。
 * 缓存有效期内多次调用直接返回缓存（不触发网络请求）。
 * @returns 纯浏览器开发回退 mock；Wails 下后端请求失败回退空数组（轮播自动隐藏）
 */
export async function getHomeBanners(): Promise<Banner[]> {
  if (!isBackendAvailable()) {
    return mockBanners
  }
  if (cachedBanners && Date.now() - cachedAt < CACHE_TTL_MS) {
    return cachedBanners
  }
  if (!inflight) {
    inflight = GetHomeBanners()
      .then((data) => {
        cachedBanners = data
        cachedAt = Date.now()
        return data
      })
      .catch(() => {
        // 后端请求失败：提示用户，并回退空数组让轮播自动隐藏，下次挂载自动重试
        toast.error('首页轮播加载失败，请稍后重试')
        return [] as Banner[]
      })
      .finally(() => {
        inflight = null
      })
  }
  return inflight
}
