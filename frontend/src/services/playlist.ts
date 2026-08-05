import type { Playlist } from '@/types'
import { mockPlaylists } from '@/data/mock'
import toast from 'react-hot-toast'
import { GetHomePlaylists } from '../../wailsjs/go/playlist/PlaylistService'

/**
 * 首页官方歌单接口
 *
 * 已接入后端（Wails 绑定 window.go.playlist.PlaylistService.*），
 * 后端调酷狗 /top/playlist 接口返回官方推荐歌单。
 *
 * 关于「只跑前端 npm run dev」：
 * Wails 的 window.go 运行时只在 Wails WebView 中注入，普通浏览器里不存在。
 * 此处做 isBackendAvailable 检测：浏览器环境下自动回退到 mock，
 * 便于纯前端 UI 开发时板块不空白。
 */

/** Wails 运行时是否可用（真实后端是否存在） */
function isBackendAvailable(): boolean {
  return typeof window !== 'undefined' && Boolean((window as any).go?.playlist?.PlaylistService)
}

/**
 * 缓存有效期(ms)：官方歌单更新不频繁，缓存 1h 减少重复请求。
 * 首页在导航切换时反复挂载，直接拉接口会重复打后端。
 */
const CACHE_TTL_MS = 60 * 60 * 1000

/** 前端请求官方歌单的默认参数：与后端约定 pageSize=15、withSong=0 */
export const DEFAULT_PLAYLIST_PAGE_SIZE = 15
export const DEFAULT_PLAYLIST_WITH_SONG = 0

// 模块级缓存：仅存在于 app 窗口生命周期；请求失败不入缓存，下次挂载自动重试。
// 以 pageSize:withSong 为 key 区分不同参数的请求，避免不同参数互相串缓存。
interface PlaylistCacheEntry {
  data: Playlist[]
  at: number
  inflight: Promise<Playlist[]> | null
}
const playlistCache = new Map<string, PlaylistCacheEntry>()

/**
 * 获取首页官方歌单
 * @param pageSize 期望返回条数，默认 15
 * @param withSong 是否携带歌曲（0 不携带 / 1 携带），默认 0
 * 后端固定返回一批歌单，列数由 PlaylistRecommend 的 CSS Grid 自适应，不重复请求。
 * 缓存有效期内多次调用直接返回缓存。
 * @returns 纯浏览器开发回退 mock；Wails 下后端请求失败回退空数组（板块自动隐藏）
 */
export async function getHomePlaylists(
  pageSize = DEFAULT_PLAYLIST_PAGE_SIZE,
  withSong = DEFAULT_PLAYLIST_WITH_SONG,
): Promise<Playlist[]> {
  if (!isBackendAvailable()) {
    return mockPlaylists
  }
  const key = `${pageSize}:${withSong}`
  let entry = playlistCache.get(key)
  if (entry && Date.now() - entry.at < CACHE_TTL_MS) {
    return entry.data
  }
  if (!entry) {
    entry = { data: [], at: 0, inflight: null }
    playlistCache.set(key, entry)
  }
  if (!entry.inflight) {
    entry.inflight = GetHomePlaylists(pageSize, withSong)
      .then((data) => {
        entry!.data = data
        entry!.at = Date.now()
        return data
      })
      .catch(() => {
        // 后端请求失败：提示用户，并回退空数组让板块自动隐藏，下次挂载自动重试
        toast.error('首页歌单加载失败，请稍后重试')
        return [] as Playlist[]
      })
      .finally(() => {
        entry!.inflight = null
      })
  }
  return entry.inflight
}
