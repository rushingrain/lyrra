import type { Playlist, Song } from '@/types'
import { mockPlaylists, mockSongs } from '@/data/mock'
import toast from 'react-hot-toast'
import {
  GetAllPlaylistSongs as GetAllPlaylistSongsAPI,
  GetHotPlaylists as GetHotPlaylistsAPI,
  GetPlaylistSongs as GetPlaylistSongsAPI,
} from '../../wailsjs/go/playlist/Service'

/**
 * 官方歌单接口
 *
 * 已接入后端（Wails 绑定 window.go.playlist.Service.*），
 * 后端调酷狗 /top/playlist 接口返回官方推荐歌单。
 *
 * 关于「只跑前端 npm run dev」：
 * Wails 的 window.go 运行时只在 Wails WebView 中注入，普通浏览器里不存在。
 * 此处做 isBackendAvailable 检测：浏览器环境下自动回退到 mock，
 * 便于纯前端 UI 开发时板块不空白。
 */

/** Wails 运行时是否可用（真实后端是否存在） */
function isBackendAvailable(): boolean {
  return typeof window !== 'undefined' && Boolean((window as any).go?.playlist?.Service)
}

/**
 * 缓存有效期(ms)：官方歌单缓存 20 分钟，减少页面切换时的重复请求。
 * 首页在导航切换时反复挂载，直接拉接口会重复打后端。
 */
const CACHE_TTL_MS = 20 * 60 * 1000

/**
 * 官方歌单统一请求条数：首页与列表页共用同一批数据，避免推荐接口分两次请求时
 * 返回不同集合，导致首页出现的歌单在列表页缺失。
 */
export const OFFICIAL_PLAYLIST_PAGE_SIZE = 15
export const DEFAULT_PLAYLIST_WITH_SONG = 0

/** 歌单歌曲分页大小：与后端默认页大小一致 */
export const PLAYLIST_PAGE_SIZE = 30

/** 歌单歌曲分页结果；info 由后端从歌单接口的 list_info 解析 */
export interface PlaylistSongsPage {
  info: Playlist
  songs: Song[]
  total: number
  hasMore: boolean
}

// 模块级缓存：仅存在于 app 窗口生命周期；请求失败不入缓存，下次挂载自动重试。
// 以 pageSize:withSong 为 key 区分不同参数的请求，避免不同参数互相串缓存。
interface PlaylistCacheEntry {
  data: Playlist[]
  at: number
  inflight: Promise<Playlist[]> | null
}
const playlistCache = new Map<string, PlaylistCacheEntry>()

/**
 * 获取官方歌单
 * @param pageSize 期望返回条数，默认 15
 * @param withSong 是否携带歌曲（0 不携带 / 1 携带），默认 0
 * 后端固定返回一批歌单，页面按各自的卡片布局渲染，不重复请求。
 * 缓存有效期内多次调用直接返回缓存。
 * @returns 纯浏览器开发回退 mock；Wails 下后端请求失败回退空数组（板块自动隐藏）
 */
export async function GetHotPlaylists(
  pageSize = OFFICIAL_PLAYLIST_PAGE_SIZE,
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
    entry.inflight = GetHotPlaylistsAPI(pageSize, withSong)
      .then((data) => {
        entry!.data = data
        entry!.at = Date.now()
        return data
      })
      .catch(() => {
        // 后端请求失败：提示用户，并回退空数组让板块自动隐藏，下次挂载自动重试
        toast.error('官方歌单加载失败，请稍后重试')
        return [] as Playlist[]
      })
      .finally(() => {
        entry!.inflight = null
      })
  }
  return entry.inflight
}

// 歌单歌曲只在途请求去重，不缓存已完成的页：歌单内容可能变化，也避免多份列表常驻内存。
const songsInflight = new Map<string, Promise<PlaylistSongsPage>>()

/**
 * 获取歌单歌曲（分页）。
 * @param globalCollectionId 歌单的 global_collection_id
 * @param page 页码，从 1 开始
 * @param pageSize 每页条数，默认 30
 */
export function getPlaylistSongs(
  globalCollectionId: string,
  page = 1,
  pageSize = PLAYLIST_PAGE_SIZE,
): Promise<PlaylistSongsPage> {
  if (!isBackendAvailable()) {
    // 纯浏览器 mock：用本地歌单/歌曲模拟分页
    const info =
      mockPlaylists.find((item) => item.globalCollectionId === globalCollectionId) ?? mockPlaylists[0]
    const start = (page - 1) * pageSize
    const songs = mockSongs.slice(start, start + pageSize)
    return Promise.resolve({
      info,
      songs,
      total: mockSongs.length,
      hasMore: start + songs.length < mockSongs.length,
    })
  }

  const key = `${globalCollectionId}:${page}:${pageSize}`
  const existing = songsInflight.get(key)
  if (existing) return existing

  const request = GetPlaylistSongsAPI(globalCollectionId, page, pageSize).finally(() => {
    songsInflight.delete(key)
  })
  songsInflight.set(key, request)
  return request
}

/**
 * 获取歌单全部歌曲（“播放全部”使用）。
 * 后端按上游单页上限分页拉取并去重，前端只需等待一次结果。
 */
export function getAllPlaylistSongs(globalCollectionId: string): Promise<Song[]> {
  if (!isBackendAvailable()) {
    return Promise.resolve(mockSongs)
  }
  return GetAllPlaylistSongsAPI(globalCollectionId)
}
