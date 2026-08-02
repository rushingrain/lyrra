import {
  SkipBack,
  Play,
  SkipForward,
  Pause,
  Repeat,
  Shuffle,
  Heart,
  MessageSquare,
  Volume2,
  Volume1,
  VolumeX,
  MoreHorizontal,
  ListVideo,
  RotateCw,
  MoveRight
} from 'lucide-react'
import { useEffect, useRef, useState } from 'react'
import { mockPlayHistory } from '../../data/mock'

type PlayMode = 'list-loop' | 'order' | 'single' | 'shuffle'

const PLAY_MODE_ORDER: PlayMode[] = ['list-loop', 'order', 'single', 'shuffle']
const PLAY_MODE_LABEL: Record<PlayMode, string> = {
  'list-loop': '列表循环',
  order: '顺序播放',
  single: '单曲循环',
  shuffle: '随机播放',
}

export function MiniPlayer() {
  const currentSong = mockPlayHistory[0]?.song
  const duration = currentSong?.duration ?? 0

  const [isPlaying, setIsPlaying] = useState(false)
  const [currentTime, setCurrentTime] = useState(Math.floor(duration * 0.3))
  const [playMode, setPlayMode] = useState<PlayMode>('list-loop')
  const [volume, setVolume] = useState(60)
  const [isLiked, setIsLiked] = useState(false)
  const [showLyrics, setShowLyrics] = useState(false)

  // 进度条拖拽与悬浮
  const trackRef = useRef<HTMLDivElement>(null)
  const [isDragging, setIsDragging] = useState(false)
  const [hoverTime, setHoverTime] = useState<number | null>(null)
  const [hoverX, setHoverX] = useState(0)

  // 播放进度模拟：播放时每秒推进
  useEffect(() => {
    if (!isPlaying || duration === 0) return
    const timer = setInterval(() => {
      setCurrentTime((t) => (t >= duration ? 0 : t + 1))
    }, 1000)
    return () => clearInterval(timer)
  }, [isPlaying, duration])

  // 拖拽时跟随鼠标更新进度
  useEffect(() => {
    if (!isDragging) return
    const onMove = (e: MouseEvent) => {
      const track = trackRef.current
      if (!track || duration === 0) return
      const rect = track.getBoundingClientRect()
      const ratio = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width))
      setCurrentTime(ratio * duration)
      setHoverTime(ratio * duration)
      setHoverX(Math.min(Math.max(e.clientX - rect.left, 24), rect.width - 24))
    }
    const onUp = () => {
      setIsDragging(false)
      setHoverTime(null)
    }
    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
    return () => {
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
    }
  }, [isDragging, duration])

  // 音量竖向滑块拖拽
  const volumeTrackRef = useRef<HTMLDivElement>(null)
  const [isVolumeDragging, setIsVolumeDragging] = useState(false)

  useEffect(() => {
    if (!isVolumeDragging) return
    const onMove = (e: MouseEvent) => {
      const track = volumeTrackRef.current
      if (!track) return
      const rect = track.getBoundingClientRect()
      // 竖向：底部为最小值，顶部为最大值
      const ratio = Math.min(1, Math.max(0, (rect.bottom - e.clientY) / rect.height))
      setVolume(Math.round(ratio * 100))
    }
    const onUp = () => setIsVolumeDragging(false)
    window.addEventListener('mousemove', onMove)
    window.addEventListener('mouseup', onUp)
    return () => {
      window.removeEventListener('mousemove', onMove)
      window.removeEventListener('mouseup', onUp)
    }
  }, [isVolumeDragging])

  if (!currentSong) return null

  const formatDuration = (seconds: number) => {
    const safe = Math.max(0, Math.floor(seconds))
    const m = Math.floor(safe / 60)
    const s = safe % 60
    return `${m}:${s.toString().padStart(2, '0')}`
  }

  const formatCount = (n: number) =>
    n >= 10000 ? `${(n / 10000).toFixed(1)}w` : `${n}`

  const progressPercent = duration > 0 ? (currentTime / duration) * 100 : 0
  const VolumeIcon = volume === 0 ? VolumeX : volume < 50 ? Volume1 : Volume2

  const handleTrackMouseMove = (e: React.MouseEvent) => {
    if (isDragging) return
    const track = trackRef.current
    if (!track || duration === 0) return
    const rect = track.getBoundingClientRect()
    const ratio = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width))
    setHoverTime(ratio * duration)
    setHoverX(Math.min(Math.max(e.clientX - rect.left, 24), rect.width - 24))
  }

  const handleTrackMouseDown = (e: React.MouseEvent) => {
    const track = trackRef.current
    if (!track || duration === 0) return
    const rect = track.getBoundingClientRect()
    const ratio = Math.min(1, Math.max(0, (e.clientX - rect.left) / rect.width))
    setIsDragging(true)
    setCurrentTime(ratio * duration)
  }

  const handleVolumeMouseDown = (e: React.MouseEvent) => {
    const track = volumeTrackRef.current
    if (!track) return
    const rect = track.getBoundingClientRect()
    const ratio = Math.min(1, Math.max(0, (rect.bottom - e.clientY) / rect.height))
    setIsVolumeDragging(true)
    setVolume(Math.round(ratio * 100))
  }

  // 通用图标按钮样式
  const iconBtn =
    'flex h-8 w-8 cursor-pointer items-center justify-center rounded-md text-[var(--color-text-secondary)] transition-colors hover:bg-gray-100'

  return (
    <div className="flex h-full w-full flex-col">
      {/* 顶部播放进度条（可拖拽 / 悬浮显示时间） */}
      <div
        className="group relative w-full cursor-pointer select-none"
        onMouseMove={handleTrackMouseMove}
        onMouseLeave={() => !isDragging && setHoverTime(null)}
        onMouseDown={handleTrackMouseDown}
      >
        <div
          ref={trackRef}
          className="relative h-[3px] w-full rounded-full bg-[var(--color-border)]"
        >
          {/* 已播放进度 */}
          <div
            className="absolute inset-y-0 left-0 rounded-full bg-[var(--color-accent)]"
            style={{ width: `${progressPercent}%` }}
          >
            {/* 拖拽手柄 */}
            <div
              className={`absolute right-0 top-1/2 h-3 w-3 -translate-y-1/2 translate-x-1/2 rounded-full bg-[var(--color-accent)] shadow transition-opacity ${isDragging || hoverTime !== null
                ? 'opacity-100'
                : 'opacity-0 group-hover:opacity-100'
                }`}
            />
          </div>
        </div>

        {/* 悬浮时间提示：当前时间 / 总时间 */}
        {hoverTime !== null && (
          <div
            className="pointer-events-none absolute -top-7 -translate-x-1/2 whitespace-nowrap rounded bg-black/75 px-1.5 py-0.5 text-[10px] text-white"
            style={{ left: hoverX }}
          >
            {formatDuration(hoverTime)} / {formatDuration(duration)}
          </div>
        )}
      </div>

      {/* 底部控制栏 */}
      <div className="flex flex-1 items-center gap-4 px-10">
        {/* 左：封面 + 歌曲信息 + 点赞 + 评论 */}
        <div className="flex w-[340px] items-center gap-3">
          {/* 唱片样式：黑圈包裹圆形图片 */}
          <div className="flex cursor-pointer items-center justify-center rounded-full bg-black p-2.5 transition-transform hover:scale-105">
            <img
              src={currentSong.coverUrl}
              alt={currentSong.title}
              className="h-10 w-10 rounded-full object-cover"
            />
          </div>
          <div className="min-w-0 flex-1">
            <p className="truncate text-sm font-medium text-[var(--color-text-primary)]">
              {currentSong.title}
            </p>
            <p className="truncate text-xs text-[var(--color-text-secondary)]">
              {currentSong.artist}
            </p>
          </div>
          <button
            className="flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-xs text-[var(--color-text-secondary)] transition-colors hover:bg-gray-100"
            aria-label="点赞"
            title="点赞"
          >
            <Heart size={16} />
            <span>{formatCount(1234567)}</span>
          </button>
          <button
            className="flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-xs text-[var(--color-text-secondary)] transition-colors hover:bg-gray-100"
            aria-label="评论"
            title="评论"
          >
            <MessageSquare size={16} />
            <span>{formatCount(86234)}</span>
          </button>
        </div>

        {/* 中：播放控制 */}
        <div className="flex flex-1 items-center justify-center gap-3">
          <div className="group relative">
            <button
              className={iconBtn}
              onClick={() =>
                setPlayMode((m) => {
                  const idx = PLAY_MODE_ORDER.indexOf(m)
                  return PLAY_MODE_ORDER[(idx + 1) % PLAY_MODE_ORDER.length]
                })
              }
              aria-label={PLAY_MODE_LABEL[playMode]}
            >
              {playMode === 'list-loop' && <Repeat size={18} />}
              {playMode === 'order' && (
                <span className="flex flex-col items-center leading-none">
                  <MoveRight size={18} />
                  <MoveRight size={18} className="-mt-2.5" />
                </span>
              )}
              {playMode === 'shuffle' && <Shuffle size={18} />}
              {playMode === 'single' && (
                <span className="relative inline-flex items-center justify-center">
                  <RotateCw size={18} />
                  <span className="absolute inset-0 flex items-center justify-center text-[8px] font-bold leading-none">
                    1
                  </span>
                </span>
              )}
            </button>
            {/* 悬浮显示当前播放方式 */}
            <span className="pointer-events-none absolute bottom-full left-1/2 z-30 mb-2 -translate-x-1/2 whitespace-nowrap rounded bg-black/75 px-2 py-1 text-[10px] text-white opacity-0 transition-opacity duration-150 group-hover:opacity-100">
              {PLAY_MODE_LABEL[playMode]}
            </span>
          </div>
          <button className={iconBtn} aria-label="上一首" title="上一首">
            <SkipBack size={20} fill="currentColor" />
          </button>
          <button
            className="flex h-9 w-9 cursor-pointer items-center justify-center rounded-full bg-[var(--color-accent)] text-[var(--color-on-accent)] transition-colors hover:bg-[var(--color-accent-hover)]"
            onClick={() => setIsPlaying((p) => !p)}
            aria-label={isPlaying ? '暂停' : '播放'}
            title={isPlaying ? '暂停' : '播放'}
          >
            {isPlaying ? (
              <Pause size={18} fill="currentColor" />
            ) : (
              <Play size={18} fill="currentColor" className="ml-0.5" />
            )}
          </button>
          <button className={iconBtn} aria-label="下一首" title="下一首">
            <SkipForward size={20} fill="currentColor" />
          </button>

          <button
            className="relative flex h-8 w-8 cursor-pointer items-center justify-center rounded-md text-[var(--color-text-secondary)] transition-colors hover:bg-gray-100"
            aria-label="播放列表"
            title="播放列表"
          >
            <ListVideo size={18} />
          </button>
        </div>

        {/* 右：播放列表 + 音质 + 收藏 + 桌面歌词 + 音量 + 更多 */}
        <div className="flex w-[340px] items-center justify-end gap-1">
          <button
            className="flex h-7 cursor-pointer items-center rounded px-1.5 text-sm font-medium text-[var(--color-text-secondary)] transition-colors hover:bg-gray-100"
            title="音质选择"
          >
            无损
          </button>
          <button
            className={`${iconBtn} ${isLiked ? 'text-red-400' : ''}`}
            onClick={() => setIsLiked((v) => !v)}
            aria-label="收藏"
            title="收藏"
          >
            <Heart size={18} fill={isLiked ? 'currentColor' : 'none'} />
          </button>
          <button
            className={`flex h-8 w-8 cursor-pointer items-center justify-center rounded-md text-base font-medium transition-colors hover:bg-gray-100 ${showLyrics ? 'text-[var(--color-accent)]' : 'text-[var(--color-text-secondary)]'
              }`}
            onClick={() => setShowLyrics((v) => !v)}
            aria-label="桌面歌词"
            title="桌面歌词"
          >
            词
          </button>

          {/* 音量调节（竖向滑块，悬浮展开） */}
          <div className="group relative flex items-center">
            <button
              className={iconBtn}
              onClick={() => setVolume((v) => (v === 0 ? 60 : 0))}
              aria-label="音量"
              title="音量"
            >
              <VolumeIcon size={18} />
            </button>
            <div
              className={`absolute bottom-full left-1/2 z-30 -translate-x-1/2 flex-col items-center gap-2 rounded-lg bg-[var(--color-background)] p-2 shadow-lg ${isVolumeDragging ? 'flex' : 'hidden group-hover:flex'
                }`}
            >
              <div
                className="relative flex h-24 w-6 cursor-pointer items-center justify-center"
                onMouseDown={handleVolumeMouseDown}
              >
                <div
                  ref={volumeTrackRef}
                  className="relative h-full w-1 rounded-full bg-[var(--color-border)]"
                >
                  {/* 已调节音量（从下往上） */}
                  <div
                    className="absolute bottom-0 left-0 w-full rounded-full bg-[var(--color-accent)]"
                    style={{ height: `${volume}%` }}
                  >
                    <div className="absolute left-1/2 top-0 h-3 w-3 -translate-x-1/2 -translate-y-1/2 rounded-full bg-[var(--color-accent)] shadow" />
                  </div>
                </div>
              </div>
              {/* 百分比 */}
              <span className="text-[10px] tabular-nums text-[var(--color-text-secondary)]">
                {volume}%
              </span>
            </div>
          </div>

          <button className={iconBtn} aria-label="更多" title="更多">
            <MoreHorizontal size={18} />
          </button>
        </div>
      </div>
    </div>
  )
}
