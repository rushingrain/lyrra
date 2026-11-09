import { useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import { ChevronDown, Play } from 'lucide-react'
import { Cover } from '@/components/common/Cover'
import { usePlayerStore } from '@/stores/playerStore'
import { getLyrics, type LyricLine } from '@/services/song'

interface LyricsViewProps {
  /** 收起歌词页，返回主界面 */
  onClose: () => void
}

type LoadState = 'loading' | 'ready' | 'error'

/** 顶部 / 底部渐隐区高度占比：越靠边缘的歌词越淡，突出中间当前句 */
const FADE_MASK = 'linear-gradient(to bottom, transparent, #000 18%, #000 82%, transparent)'

/** 毫秒 → mm:ss，用于歌词行的跳转按钮 */
function formatTime(ms: number): string {
  const totalSeconds = Math.max(0, Math.floor(ms / 1000))
  const minutes = Math.floor(totalSeconds / 60)
  const seconds = totalSeconds % 60
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
}

/**
 * 歌词详情页
 *
 * 由 MiniPlayer 的歌曲封面唤出，覆盖侧边栏 / 顶栏 / 内容区，底边紧贴 MiniPlayer。
 * 沉浸背景由 MainLayout 的 NowPlayingBackdrop 统一绘制（覆盖整窗，含播放栏），
 * 这里只负责内容，透明浮在它上面。
 * 歌词列跟随播放进度把当前句垂直居中高亮，可见行数随窗口高度自适应；点击某句可跳转播放。
 */
export function LyricsView({ onClose }: LyricsViewProps) {
  const song = usePlayerStore((s) => s.song)
  const currentTime = usePlayerStore((s) => s.currentTime)
  const seek = usePlayerStore((s) => s.seek)

  const [lines, setLines] = useState<LyricLine[]>([])
  const [state, setState] = useState<LoadState>('loading')

  const hash = song?.hash ?? ''
  const title = song?.name ?? '未播放'

  // 打开歌词页（或切歌）时拉取当前歌曲的歌词；期间用户又切歌则丢弃旧结果。
  useEffect(() => {
    if (!hash) {
      setLines([])
      setState('ready')
      return
    }
    let cancelled = false
    setLines([])
    setState('loading')
    getLyrics(hash)
      .then((data) => {
        if (cancelled) return
        setLines(data)
        setState('ready')
      })
      .catch((error) => {
        if (cancelled) return
        // 歌词属于可降级信息：静默降级为占位文案，不弹提示
        console.warn('[lyrics] unavailable', error)
        setState('error')
      })
    return () => {
      cancelled = true
    }
  }, [hash])

  // 当前句 = 最后一个时间点不大于播放进度的行（歌词按时间升序，可直接顺序扫描）
  const activeIndex = useMemo(() => {
    let index = -1
    for (let i = 0; i < lines.length; i++) {
      if (lines[i].time > currentTime) break
      index = i
    }
    return index
  }, [lines, currentTime])

  const listRef = useRef<HTMLDivElement | null>(null)
  const lineRefs = useRef<Array<HTMLDivElement | null>>([])
  // 上下留白 = 容器高度的一半，让首尾行也能滚到正中间
  const [pad, setPad] = useState(0)

  // 容器高度随窗口变化，用 ResizeObserver 同步留白与滚动基准
  const observeList = useCallback((node: HTMLDivElement | null) => {
    listRef.current = node
    if (!node) return
    const update = () => setPad(node.clientHeight / 2)
    update()
    const observer = new ResizeObserver(update)
    observer.observe(node)
    return () => observer.disconnect()
  }, [])

  // 当前句垂直居中：offsetTop 相对滚动容器，减去半屏偏移即得目标 scrollTop
  useLayoutEffect(() => {
    const list = listRef.current
    const line = activeIndex >= 0 ? lineRefs.current[activeIndex] : null
    if (!list || !line) return
    const top = line.offsetTop - (list.clientHeight - line.clientHeight) / 2
    list.scrollTo({ top: Math.max(0, top), behavior: 'smooth' })
  }, [activeIndex, lines, pad])

  const placeholder = state === 'loading' ? '歌词加载中…' : state === 'error' ? '歌词加载失败' : '暂无歌词'

  return (
    <div className="immersive-fade-in absolute inset-0 z-30 overflow-hidden">
      <div className="relative flex h-full flex-col">
        {/* 头部：仅收起按钮（歌名 / 歌手移到歌词列表上方） */}
        <header className="flex items-center px-8 py-4">
          <button
            type="button"
            onClick={onClose}
            className="flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-sm text-white/70 transition-colors hover:text-white"
            aria-label="收起歌词"
            title="收起歌词"
          >
            <ChevronDown size={20} />
          </button>
        </header>

        {/* 正文：左封面 + 右歌词列，宽度 45:55（歌词起点约在窗口 47% 处） */}
        <div className="flex min-h-0 flex-1 gap-12 px-12 pb-10">
          <div className="hidden min-w-0 flex-[45] items-center justify-center lg:flex">
            <div
              className="shrink-0"
              style={{ width: 'clamp(200px, 32vh, 340px)', height: 'clamp(200px, 32vh, 340px)' }}
            >
              <Cover
                src={song?.coverUrl}
                alt={title}
                variant="rounded"
                className="h-full w-full shadow-2xl"
              />
            </div>
          </div>

          <div className="flex h-full min-w-0 flex-[55] flex-col">
            {/* 歌名 / 歌手：放在歌词列表上方 */}
            <div className="mb-3 shrink-0">
              <h2 className="truncate text-2xl font-semibold text-white">{title}</h2>
              <p className="mt-1 truncate text-sm text-white/60">歌手：{song?.artist || '未知歌手'}</p>
            </div>

            <div className="relative min-h-0 flex-1">
              {lines.length === 0 ? (
                <div className="flex h-full items-center justify-center text-lg text-white/60">
                  {placeholder}
                </div>
              ) : (
                <div
                  ref={observeList}
                  className="hide-scrollbar h-full overflow-y-auto"
                  style={{
                    paddingTop: pad,
                    paddingBottom: pad,
                    maskImage: FADE_MASK,
                    WebkitMaskImage: FADE_MASK,
                  }}
                >
                  {lines.map((line, index) => (
                    <div
                      key={`${line.time}-${index}`}
                      ref={(el) => {
                        lineRefs.current[index] = el
                      }}
                      className="group flex w-full items-center gap-3 py-2.5"
                    >
                      <span
                        className={`min-w-0 flex-1 leading-relaxed transition-colors duration-300 ${
                          index === activeIndex
                            ? 'text-2xl font-bold text-[var(--color-accent)]'
                            : 'text-xl font-medium text-white/55 group-hover:text-white'
                        }`}
                      >
                        {line.text}
                      </span>
                      {/* 鼠标移到该行才出现：跳到这句的时间点播放 */}
                      <button
                        type="button"
                        onClick={() => seek(line.time)}
                        className="pointer-events-none flex shrink-0 cursor-pointer items-center gap-1 rounded-full bg-white/10 px-2 py-0.5 text-xs font-medium text-white/70 opacity-0 transition-opacity hover:bg-white/20 hover:text-white focus-visible:opacity-100 group-hover:pointer-events-auto group-hover:opacity-100"
                        title={`播放到 ${formatTime(line.time)}`}
                        aria-label={`播放到 ${formatTime(line.time)}`}
                      >
                        <Play size={12} fill="currentColor" />
                        <span className="tabular-nums">{formatTime(line.time)}</span>
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
