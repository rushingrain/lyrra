import { create } from 'zustand'
import toast from 'react-hot-toast'
import type { Song } from '@/types'
import { getSongPlayURL } from '@/services/play'
import { getSongStats } from '@/services/song'

/**
 * 播放器状态（全局唯一）
 *
 * 只维护一个 HTMLAudioElement：src 直接指向后端返回的 CDN 地址，
 * 由 WebView 用 HTTP Range 流式播放，不下载、不落盘。
 * 时间统一用毫秒（与后端 dto.Song.duration 一致），UI 层负责格式化。
 */

export type PlayerStatus = 'idle' | 'loading' | 'playing' | 'paused' | 'error'

/** 播放模式：列表循环 / 顺序播放 / 单曲循环 / 随机播放 */
export type PlayMode = 'list-loop' | 'order' | 'single' | 'shuffle'

/** 播放模式的切换顺序（与界面上的按钮顺序一致） */
export const PLAY_MODE_ORDER: PlayMode[] = ['list-loop', 'order', 'single', 'shuffle']

/**
 * 播放队列上限：与后端 playerstate.MaxPlaylistSize 保持一致。
 * 队列满时不再接纳新歌（提示用户先清空），避免悄悄丢掉队列里已有的条目。
 */
export const MAX_PLAYLIST_SIZE = 500

/** 播放模式归一化：不在白名单里的值（本地文件被手工改过等）回落到列表循环 */
function normalizePlayMode(value: string): PlayMode {
  return (PLAY_MODE_ORDER as string[]).includes(value) ? (value as PlayMode) : 'list-loop'
}

/** 音量归一化到 0-100 的整数 */
function clampVolume(volume: number): number {
  if (!Number.isFinite(volume)) return 0
  return Math.min(100, Math.max(0, Math.round(volume)))
}

/** 单条统计：count 是原始数量，text 是后端格式化后的展示文案（k / w / 亿） */
export interface SongStat {
  count: number
  text: string
}

/** 当前歌曲的统计数据；某项为 null 表示取不到，UI 显示占位符 */
export interface PlayerSongStats {
  favorite: SongStat | null
  comment: SongStat | null
}

/**
 * 本地持久化的播放状态（结构对应后端的 playerstate.State）。
 * 只带歌曲信息，不带播放地址：恢复后第一次播放时重新解析地址。
 */
export interface StoredPlayerState {
  device: { playMode: string; volume: number }
  account: { playlist?: Song[]; currentSong?: Song }
}

interface PlayerState {
  song: Song | null
  /** 当前曲目的播放地址（带签名，会过期；出错时清空以便重新解析） */
  url: string | null
  /** 后端实际使用的音质档位（res / sq / hq / standard） */
  quality: string | null
  /** 上游把该档位降级了（例如没有 Hi-Res 时 high 落到标准文件） */
  degraded: boolean
  status: PlayerStatus
  /** 已播放时长(ms) */
  currentTime: number
  /** 总时长(ms)，优先用 audio 元数据，取不到时用 Song.duration */
  duration: number
  /** 音量 0-100 */
  volume: number
  error: string | null
  /** 当前歌曲的收藏数 / 评论数，取不到时为 null（切歌时先清空再按缓存或请求填充） */
  stats: PlayerSongStats | null
  /** 当前播放模式 */
  playMode: PlayMode
  /** 播放列表：点击播放按钮的歌曲按顺序入列，同一首不重复 */
  playlist: Song[]

  /** 播放指定歌曲：同一首 → 播放/暂停切换；不同首 → 解析地址并播放 */
  playSong: (song: Song) => Promise<void>
  /** 播放/暂停切换（无曲目时不动作） */
  toggle: () => void
  /** 切换播放模式（列表循环 → 顺序 → 单曲 → 随机） */
  cyclePlayMode: () => void
  /** 清空播放列表（正在播放的歌曲不受影响，播完即停） */
  clearPlaylist: () => void
  /** 播放下一首（按当前模式决定目标；顺序播放到末尾则停下） */
  playNext: () => void
  /** 播放上一首（随机模式下按列表顺序回退） */
  playPrev: () => void
  /** 一曲播放结束：按当前模式续播 */
  handleEnded: () => void
  /** 启动时恢复本地保存的播放状态（播放模式 / 音量 / 播放队列 / 当前歌曲） */
  restoreState: (snapshot: StoredPlayerState) => void
  /** 登出：清掉账号级数据（播放队列 / 当前歌曲），设备设置保留 */
  resetAccountData: () => void
  /** 跳转到指定位置(ms) */
  seek: (ms: number) => void
  setVolume: (volume: number) => void
}

/** 唯一的 audio 实例，首次播放时创建 */
let audio: HTMLAudioElement | null = null
/** 请求序号：切歌期间旧请求返回后丢弃，避免乱序覆盖 */
let requestSeq = 0

const MS_PER_SECOND = 1000

/**
 * 统计数据缓存有效期：收藏数 / 评论数短时间内不会变，
 * 切回同一首歌直接命中缓存，不必重新请求。
 */
const STATS_TTL_MS = 15 * 60 * 1000

/** 缓存与在途请求都按「song.id + hash」建键：同一首歌的重播共用一次请求 */
const statsCache = new Map<string, { data: PlayerSongStats; at: number }>()
const statsInflight = new Map<string, Promise<PlayerSongStats>>()

function statsCacheKey(song: Song): string {
  return `${song.id}:${song.hash ?? ''}`
}

function readStatsCache(key: string): PlayerSongStats | null {
  const entry = statsCache.get(key)
  if (!entry) return null
  if (Date.now() - entry.at >= STATS_TTL_MS) {
    statsCache.delete(key)
    return null
  }
  return entry.data
}

function toStat(count: { count: number; text: string } | undefined): SongStat | null {
  if (!count) return null
  return { count: count.count, text: count.text }
}

/**
 * 拉取统计数据并缓存。统计数据属于可降级信息：失败不弹提示、也不写缓存，
 * 下次切回这首歌会重试；同一首歌的并发调用共用同一个请求。
 */
function fetchSongStats(song: Song): Promise<PlayerSongStats> {
  const key = statsCacheKey(song)
  const pending = statsInflight.get(key)
  if (pending) return pending

  const task = getSongStats(song.id, song.hash ?? '')
    .then((stats) => {
      const data: PlayerSongStats = {
        favorite: toStat(stats.favorite),
        comment: toStat(stats.comment),
      }
      statsCache.set(key, { data, at: Date.now() })
      return data
    })
    .finally(() => {
      statsInflight.delete(key)
    })

  statsInflight.set(key, task)
  return task
}

/**
 * 把统计数据写进 store。缓存命中时同步写入（切回时不留空档），
 * 请求失败静默降级为 null，绝不打断播放。
 */
function applySongStats(song: Song, seq: number): void {
  const cached = readStatsCache(statsCacheKey(song))
  if (cached) {
    usePlayerStore.setState({ stats: cached })
    return
  }

  void fetchSongStats(song)
    .then((stats) => {
      if (seq !== requestSeq) return
      usePlayerStore.setState({ stats })
    })
    .catch((error) => {
      if (seq !== requestSeq) return
      console.warn('[player] song stats unavailable', error)
      usePlayerStore.setState({ stats: null })
    })
}

/**
 * 播放失败只弹一条提示：同一个 id 的 toast 会被替换而不是叠加。
 * 一次失败可能同时触发 audio.play() 的 rejection 和 <audio> 的 error 事件。
 */
const TOAST_ID_PLAYBACK = 'player-playback-error'

/** 播放列表写满的提示：同样用固定 id，连续入列时替换而不是叠加 */
const TOAST_ID_PLAYLIST_FULL = 'player-playlist-full'

function isSameSong(a: Song | null, b: Song): boolean {
  if (!a) return false
  return a.id === b.id && (a.hash ?? '') === (b.hash ?? '')
}

/**
 * 播放列表里的歌曲标识：优先用 MixSongID（同一首歌在不同接口里 hash 可能不同，
 * 按 id 去重才不会出现重复行），没有 id 时退回 hash。
 */
function songKey(song: Song): string {
  return song.id !== 0 ? `id:${song.id}` : `hash:${song.hash ?? ''}`
}

/** 当前曲目在播放列表里的下标；不在列表里（例如清空后仍在播）返回 -1 */
function currentIndex(playlist: Song[], song: Song | null): number {
  if (!song) return -1
  return playlist.findIndex((item) => songKey(item) === songKey(song))
}

/**
 * 计算下一首的下标，越界表示"没有下一首"（顺序播放到最后）：
 * 随机播放换一首，列表循环回到第一首，顺序 / 单曲循环按列表顺序前进。
 * 单曲循环的"重播当前曲目"只作用于自动续播（见 handleEnded），手动切歌不受影响。
 */
function nextIndex(mode: PlayMode, current: number, total: number): number {
  if (total <= 0) return -1
  const from = current < 0 ? -1 : current
  if (mode === 'shuffle') {
    if (total === 1) return 0
    let target = from
    while (target === from) {
      target = Math.floor(Math.random() * total)
    }
    return target
  }
  if (mode === 'list-loop') return (from + 1) % total
  return from + 1
}

/** 计算上一首的下标：随机模式也按列表顺序回退；列表循环从末尾绕回 */
function prevIndex(mode: PlayMode, current: number, total: number): number {
  if (total <= 0) return -1
  const from = current < 0 ? 0 : current
  if (mode === 'list-loop') return (from - 1 + total) % total
  return Math.max(0, from - 1)
}

function errorMessage(error: unknown): string {
  if (error instanceof Error && error.message) return error.message
  return String(error)
}

/**
 * 播放地址里带账号标识（文件名形如 ..._ap3116_us863006722_df2bea..._pi411_..._qu128_...），
 * 日志里只保留 host，并把 userid / dfid 段打码，避免把账号标识写进控制台。
 */
function maskMediaURL(raw?: string): string | undefined {
  if (!raw) return undefined
  try {
    const parsed = new URL(raw)
    const path = parsed.pathname.replace(/_us[^_]+_/g, '_us***_').replace(/_df[^_]+_/g, '_df***_')
    return `${parsed.origin}${path}`
  } catch {
    return undefined
  }
}

function getAudio(): HTMLAudioElement {
  if (audio) return audio

  audio = new Audio()
  audio.preload = 'none'

  audio.addEventListener('loadedmetadata', () => {
    if (!audio) return
    const duration = Number.isFinite(audio.duration) ? audio.duration * MS_PER_SECOND : 0
    if (duration > 0) usePlayerStore.setState({ duration })
  })
  audio.addEventListener('timeupdate', () => {
    if (!audio) return
    usePlayerStore.setState({ currentTime: audio.currentTime * MS_PER_SECOND })
  })
  audio.addEventListener('play', () => usePlayerStore.setState({ status: 'playing' }))
  audio.addEventListener('pause', () => {
    // ended 会紧接着触发 pause，这里不覆盖 ended 的归位状态；
    // 切歌时设置新 src 也会触发 pause，此时还在 loading，保持 loading 避免图标闪烁
    if (audio?.ended) return
    if (usePlayerStore.getState().status === 'loading') return
    usePlayerStore.setState({ status: 'paused' })
  })
  audio.addEventListener('ended', () => {
    // 按播放模式续播：单曲循环重播当前曲目，列表循环回到开头，顺序播放到末尾停下
    usePlayerStore.getState().handleEnded()
  })
  audio.addEventListener('error', () => {
    // 签名过期 / 网络异常：清掉地址，下次点击会重新解析
    // MEDIA_ERR_NETWORK=2（网络/被拦），MEDIA_ERR_SRC_NOT_SUPPORTED=4（地址不可用）
    console.error('[player] audio error', {
      code: audio?.error?.code,
      message: audio?.error?.message,
      src: maskMediaURL(audio?.src),
    })
    usePlayerStore.setState({ status: 'error', url: null, error: '播放失败，请重试' })
    toast.error('播放失败，请重试', { id: TOAST_ID_PLAYBACK })
  })

  return audio
}

/**
 * 从头重播当前曲目。
 * 单曲循环与"上一首=当前曲目"都走这里：不能复用 playSong（同曲会被当成暂停/继续）。
 */
function restartCurrent(): void {
  const element = getAudio()
  element.currentTime = 0
  usePlayerStore.setState({ currentTime: 0 })
  void element.play().catch((error) => {
    console.error('[player] restart failed', error)
    usePlayerStore.setState({ status: 'error', url: null, error: '播放失败，请重试' })
    toast.error('播放失败，请重试', { id: TOAST_ID_PLAYBACK })
  })
}

export const usePlayerStore = create<PlayerState>((set, get) => ({
  song: null,
  url: null,
  quality: null,
  degraded: false,
  status: 'idle',
  currentTime: 0,
  duration: 0,
  volume: 60,
  error: null,
  stats: null,
  playMode: 'list-loop',
  playlist: [],

  playSong: async (song) => {
    const element = getAudio()
    const state = get()

    // 点击播放即入列：同一首歌只保留一条（位置不变，内容更新为最新一次拿到的数据）
    const index = state.playlist.findIndex((item) => songKey(item) === songKey(song))
    if (index < 0) {
      // 队列已满：拒绝入列并明确提示，不悄悄淘汰旧条目（已在队列里的歌仍可正常切换播放）
      if (state.playlist.length >= MAX_PLAYLIST_SIZE) {
        toast.error('播放列表已满，请先清空列表', { id: TOAST_ID_PLAYLIST_FULL })
        return
      }
      set({ playlist: [...state.playlist, song] })
    } else if (state.playlist[index] !== song) {
      const playlist = [...state.playlist]
      playlist[index] = song
      set({ playlist })
    }

    // 同一首：有地址就播放/暂停切换；仍在解析则忽略重复点击；上次失败则往下重新解析
    if (isSameSong(state.song, song)) {
      if (state.status === 'loading') return
      if (state.url) {
        try {
          if (element.paused) {
            await element.play()
          } else {
            element.pause()
          }
        } catch (error) {
          set({ status: 'error', url: null, error: errorMessage(error) })
          console.error('[player] play() rejected', error)
          toast.error('播放失败，请重试', { id: TOAST_ID_PLAYBACK })
        }
        return
      }
    }

    if (!song.hash) {
      set({
        song,
        url: null,
        status: 'error',
        error: '该歌曲暂无可用音质',
        currentTime: 0,
        duration: song.duration ?? 0,
        stats: null,
      })
      toast.error('该歌曲暂无可用音质')
      return
    }

    const seq = ++requestSeq
    set({
      song,
      url: null,
      status: 'loading',
      error: null,
      currentTime: 0,
      duration: song.duration ?? 0,
      stats: null,
    })

    // 第一步：解析地址（播放优先）。统计数据并发拉取，既不阻塞播放，失败也不影响播放。
    const playRequest = getSongPlayURL(song.hash, song.id, 'standard')
    applySongStats(song, seq)

    let playURL
    try {
      playURL = await playRequest
    } catch (error) {
      if (seq !== requestSeq) return
      set({ status: 'error', url: null, error: errorMessage(error) })
      toast.error('获取播放地址失败，请重试')
      return
    }
    if (seq !== requestSeq) return // 期间用户又点了别的歌

    // 第二步：交给 audio 流式播放（不下载文件）
    element.src = playURL.url
    element.volume = state.volume / 100
    try {
      await element.play()
    } catch (error) {
      if (seq !== requestSeq) return
      set({ status: 'error', url: null, error: errorMessage(error) })
      console.error('[player] play() rejected', error)
      toast.error('播放失败，请重试', { id: TOAST_ID_PLAYBACK })
      return
    }
    if (seq !== requestSeq) return

    set({
      url: playURL.url,
      quality: playURL.quality,
      degraded: playURL.degraded,
      status: 'playing',
    })
  },

  toggle: () => {
    const { song, status } = get()
    if (!song || status === 'loading') return
    if (status === 'playing') {
      getAudio().pause()
      return
    }
    void get().playSong(song)
  },

  cyclePlayMode: () => {
    const index = PLAY_MODE_ORDER.indexOf(get().playMode)
    set({ playMode: PLAY_MODE_ORDER[(index + 1) % PLAY_MODE_ORDER.length] })
  },

  clearPlaylist: () => {
    // 只清列表，不打断正在播放的歌曲；播完或手动切歌时自然停下
    set({ playlist: [] })
  },

  playNext: () => {
    const { playlist, playMode, song } = get()
    const index = currentIndex(playlist, song)
    const target = nextIndex(playMode, index, playlist.length)
    // 顺序播放到末尾：不跳转，也不改动播放状态（当前曲目继续播）
    if (target < 0 || target >= playlist.length) return
    if (target === index) {
      restartCurrent()
      return
    }
    void get().playSong(playlist[target])
  },

  playPrev: () => {
    const { playlist, playMode, song } = get()
    const index = currentIndex(playlist, song)
    const target = prevIndex(playMode, index, playlist.length)
    if (target < 0 || target >= playlist.length) return
    // 已经在第一首（或列表里只有一首）时，等价于从头重播当前曲目
    if (target === index) {
      restartCurrent()
      return
    }
    void get().playSong(playlist[target])
  },

  handleEnded: () => {
    const { playlist, playMode, song } = get()
    if (!song) {
      set({ status: 'paused', currentTime: 0 })
      return
    }
    // 单曲循环：直接重播，不受列表影响
    if (playMode === 'single') {
      restartCurrent()
      return
    }
    const index = currentIndex(playlist, song)
    const target = nextIndex(playMode, index, playlist.length)
    if (target < 0 || target >= playlist.length) {
      // 顺序播放到末尾（或列表已清空）：停在当前曲目
      set({ status: 'paused', currentTime: 0 })
      return
    }
    if (target === index) {
      // 列表循环只剩一首曲目：重播而不是暂停
      restartCurrent()
      return
    }
    set({ currentTime: 0 })
    void get().playSong(playlist[target])
  },

  restoreState: (snapshot) => {
    const device = {
      playMode: normalizePlayMode(snapshot.device.playMode),
      volume: clampVolume(snapshot.device.volume),
    }

    // 恢复是异步的：期间用户可能已经开始播放，这时只补设备设置，不覆盖正在播放的内容
    if (get().song) {
      set(device)
      return
    }

    const currentSong = snapshot.account.currentSong ?? null
    set({
      ...device,
      // 历史文件可能超过上限，这里再裁一次（与播放时入列的规则一致：留最新的）
      playlist: (snapshot.account.playlist ?? []).slice(-MAX_PLAYLIST_SIZE),
      song: currentSong,
      // 播放地址与进度都不落盘：恢复后停在曲目开始处，点播放时重新解析地址
      url: null,
      quality: null,
      degraded: false,
      stats: null,
      error: null,
      currentTime: 0,
      duration: currentSong?.duration ?? 0,
      status: currentSong ? 'paused' : 'idle',
    })
  },

  resetAccountData: () => {
    // 账号数据随登出清空；设备设置（播放模式 / 音量）保留
    set({
      playlist: [],
      song: null,
      url: null,
      quality: null,
      degraded: false,
      stats: null,
      error: null,
      currentTime: 0,
      duration: 0,
      status: 'idle',
    })
  },

  seek: (ms) => {
    const element = getAudio()
    const { duration } = get()
    const target = Math.min(Math.max(0, ms), duration || ms)
    if (Number.isFinite(element.duration)) {
      element.currentTime = target / MS_PER_SECOND
    }
    set({ currentTime: target })
  },

  setVolume: (volume) => {
    const safe = clampVolume(volume)
    getAudio().volume = safe / 100
    set({ volume: safe })
  },
}))
