import { create } from 'zustand'
import toast from 'react-hot-toast'
import type { Song } from '@/types'
import { getSongPlayURL } from '@/services/play'

/**
 * 播放器状态（全局唯一）
 *
 * 只维护一个 HTMLAudioElement：src 直接指向后端返回的 CDN 地址，
 * 由 WebView 用 HTTP Range 流式播放，不下载、不落盘。
 * 时间统一用毫秒（与后端 dto.Song.duration 一致），UI 层负责格式化。
 */

export type PlayerStatus = 'idle' | 'loading' | 'playing' | 'paused' | 'error'

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

  /** 播放指定歌曲：同一首 → 播放/暂停切换；不同首 → 解析地址并播放 */
  playSong: (song: Song) => Promise<void>
  /** 播放/暂停切换（无曲目时不动作） */
  toggle: () => void
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
 * 播放失败只弹一条提示：同一个 id 的 toast 会被替换而不是叠加。
 * 一次失败可能同时触发 audio.play() 的 rejection 和 <audio> 的 error 事件。
 */
const TOAST_ID_PLAYBACK = 'player-playback-error'

function isSameSong(a: Song | null, b: Song): boolean {
  if (!a) return false
  return a.id === b.id && (a.hash ?? '') === (b.hash ?? '')
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
    usePlayerStore.setState({ status: 'paused', currentTime: 0 })
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

  playSong: async (song) => {
    const element = getAudio()
    const state = get()

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
    })

    // 第一步：解析地址
    let playURL
    try {
      playURL = await getSongPlayURL(song.hash, song.id)
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
    const safe = Math.min(100, Math.max(0, Math.round(volume)))
    getAudio().volume = safe / 100
    set({ volume: safe })
  },
}))
