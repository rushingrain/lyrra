import type { Song } from '@/types'
import { mockSongs } from '@/data/mock'
import toast from 'react-hot-toast'
import { GetLatestSongs, GetHotSongs } from '../../wailsjs/go/song/Service'

/**
 * 首页最新音乐接口
 *
 * 已接入后端（Wails 绑定 window.go.song.Service.*），
 * 后端调酷狗 TopSong（新歌首发榜）接口返回最新单曲。
 *
 * 关于「只跑前端 npm run dev」：
 * Wails 的 window.go 运行时只在 Wails WebView 中注入，普通浏览器里不存在。
 * 此处做 isBackendAvailable 检测：浏览器环境下自动回退到 mock，
 * 便于纯前端 UI 开发时板块不空白。
 */

/** Wails 运行时是否可用（真实后端是否存在） */
function isBackendAvailable(): boolean {
  return typeof window !== 'undefined' && Boolean((window as any).go?.song?.Service)
}

/**
 * 缓存有效期(ms)：新歌榜单更新不频繁，缓存 1h 减少重复请求。
 * 首页在导航切换时反复挂载，直接拉接口会重复打后端。
 */
const CACHE_TTL_MS = 60 * 60 * 1000

/** 前端请求最新音乐的默认参数：page=1、pagesize=15（卡片网格约两行，不多不少） */
export const DEFAULT_SONG_PAGE = 1
export const DEFAULT_SONG_PAGE_SIZE = 15

// 模块级缓存：仅存在于 app 窗口生命周期；请求失败不入缓存，下次挂载自动重试。
// 以 page:pagesize 为 key 区分不同参数的请求，避免不同参数互相串缓存。
interface SongCacheEntry {
  data: Song[]
  at: number
  inflight: Promise<Song[]> | null
}
const songCache = new Map<string, SongCacheEntry>()

/**
 * 带缓存的请求：命中缓存直接返回；进行中请求去重；失败 toast 并回退空数组。
 * 缓存按 key 区分（最新音乐按 page:pagesize，热门音乐固定 "hot"），条目有界，无泄漏。
 */
async function cachedRequest(
  key: string,
  errorMsg: string,
  fetcher: () => Promise<Song[]>,
): Promise<Song[]> {
  let entry = songCache.get(key)
  if (entry && Date.now() - entry.at < CACHE_TTL_MS) {
    return entry.data
  }
  if (!entry) {
    entry = { data: [], at: 0, inflight: null }
    songCache.set(key, entry)
  }
  if (!entry.inflight) {
    entry.inflight = fetcher()
      .then((data) => {
        entry!.data = data
        entry!.at = Date.now()
        return data
      })
      .catch(() => {
        toast.error(errorMsg)
        return [] as Song[]
      })
      .finally(() => {
        entry!.inflight = null
      })
  }
  return entry.inflight
}

/**
 * 获取首页最新音乐
 * @param page 页码，默认 1
 * @param pagesize 每页条数，默认 15
 * 缓存有效期内多次调用直接返回缓存。
 * @returns 纯浏览器开发回退 mock；Wails 下后端请求失败回退空数组（板块自动隐藏）
 */
export function getNewSongs(
  page = DEFAULT_SONG_PAGE,
  pagesize = DEFAULT_SONG_PAGE_SIZE,
): Promise<Song[]> {
  if (!isBackendAvailable()) {
    return Promise.resolve(mockSongs)
  }
  return cachedRequest(`${page}:${pagesize}`, '最新音乐加载失败，请稍后重试', () =>
    GetLatestSongs(page, pagesize),
  )
}

/**
 * 获取首页热门音乐（TopCard 单曲卡片推荐，后端固定返回一批）
 * 缓存有效期内多次调用直接返回缓存。
 * @returns 纯浏览器开发回退 mock；Wails 下后端请求失败回退空数组（板块自动隐藏）
 */
export function getHotSongs(): Promise<Song[]> {
  if (!isBackendAvailable()) {
    return Promise.resolve(mockSongs)
  }
  return cachedRequest('hot', '热门音乐加载失败，请稍后重试', () => GetHotSongs())
}
