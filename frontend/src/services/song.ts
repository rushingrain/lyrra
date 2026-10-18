import type { Song } from '@/types'
import { mockSongs } from '@/data/mock'
import toast from 'react-hot-toast'
import {
  GetDailyRecommendSongs,
  GetLatestSongs,
  GetRecommendedSongs,
  GetSongStats,
  GetLyrics,
} from '../../wailsjs/go/song/Service'
import type { song } from '../../wailsjs/go/models'

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

/** 前端请求最新音乐的默认参数 */
export const DEFAULT_SONG_PAGE = 1
export const DEFAULT_SONG_PAGE_SIZE = 30
/** 热门音乐对应的上游卡片 ID（潮流尝鲜），后端 GetRecommendedSongs 的 typeId */
const HOT_SONGS_TYPE_ID = 3005

/** 歌曲统计数据（收藏数 / 评论数）；单项缺失表示该项取不到 */
export type SongStats = song.SongStats

/** 一行时间轴歌词；time 为毫秒 */
export type LyricLine = song.LyricLine

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
 * 获取首页热门音乐（TopCardYouth 卡片推荐，后端固定返回一批）
 * 后端方法为 GetRecommendedSongs(typeId)，这里固定传热门卡片 ID。
 * 缓存有效期内多次调用直接返回缓存。
 * @returns 纯浏览器开发回退 mock；Wails 下后端请求失败回退空数组（板块自动隐藏）
 */
export function getHotSongs(): Promise<Song[]> {
  if (!isBackendAvailable()) {
    return Promise.resolve(mockSongs)
  }
  return cachedRequest('hot', '热门音乐加载失败，请稍后重试', () =>
    GetRecommendedSongs(HOT_SONGS_TYPE_ID),
  )
}

/**
 * 获取每日推荐（EverydayRecommend，按登录账号返回个性化推荐）。
 *
 * 只在登录后调用；不做本地缓存（避免切换账号后命中上一个账号的数据），
 * 仅在途请求去重（开发模式下 StrictMode 会重复触发 effect）。
 * @returns 纯浏览器开发回退 mock；Wails 下请求失败提示后返回空数组（板块不展示内容）
 */
let dailyRecommendInflight: Promise<Song[]> | null = null

export function getDailyRecommendSongs(): Promise<Song[]> {
  if (!isBackendAvailable()) {
    return Promise.resolve(mockSongs)
  }
  if (dailyRecommendInflight) return dailyRecommendInflight

  dailyRecommendInflight = GetDailyRecommendSongs()
    .catch(() => {
      toast.error('每日推荐加载失败，请稍后重试')
      return [] as Song[]
    })
    .finally(() => {
      dailyRecommendInflight = null
    })
  return dailyRecommendInflight
}

/**
 * 获取歌曲统计数据（收藏数 / 评论数）。
 *
 * mixSongID 传 Song.id，hash 传 Song.hash；后端把数量格式化成展示文案（k / w / 亿）放在 text 里。
 * 单项取不到时该字段为空，两项都取不到时后端才 reject。
 * 统计数据属于可降级信息：调用方不要弹错误提示，展示占位符即可。
 */
export function getSongStats(mixSongID: number, hash: string): Promise<SongStats> {
  return GetSongStats(mixSongID, hash)
}

/**
 * 歌词缓存有效期(ms)：同一首歌的歌词基本不变，缓存久一点减少重复请求。
 * 歌词只在用户打开歌词页时才拉取，缓存仅在 app 窗口生命周期内有效。
 */
const LYRIC_CACHE_TTL_MS = 60 * 60 * 1000

const lyricCache = new Map<string, { data: LyricLine[]; at: number }>()
const lyricInflight = new Map<string, Promise<LyricLine[]>>()

/**
 * 获取歌曲歌词（按时间升序的毫秒时间轴）。
 *
 * hash 传 Song.hash；后端把 LRC 解析成结构化列表，上游没有歌词时返回空数组（不是错误）。
 * 请求本身失败会 reject——歌词属于可降级信息，调用方静默处理即可，不要弹 toast。
 * 只缓存成功结果；空数组同样入缓存，避免每次打开都重试没有歌词的歌曲。
 */
export function getLyrics(hash: string): Promise<LyricLine[]> {
  const key = hash.trim()
  if (!key) return Promise.resolve([])

  const cached = lyricCache.get(key)
  if (cached && Date.now() - cached.at < LYRIC_CACHE_TTL_MS) {
    return Promise.resolve(cached.data)
  }

  const pending = lyricInflight.get(key)
  if (pending) return pending

  const task = GetLyrics(key)
    .then((lines) => {
      const data = lines ?? []
      lyricCache.set(key, { data, at: Date.now() })
      return data
    })
    .finally(() => {
      lyricInflight.delete(key)
    })

  lyricInflight.set(key, task)
  return task
}
