import { useEffect, useRef } from 'react'
import { Trash2 } from 'lucide-react'
import { PlayableSong } from '@/components/business/PlayableSong'
import { usePlayerStore } from '@/stores/playerStore'

interface PlaylistPanelProps {
  /** 点击面板外部时关闭 */
  onClose: () => void
}

/**
 * 触发按钮的选择器：面板外面被按下时关闭，但触发按钮自己负责开关，
 * 这里要跳过它，否则会出现"关一下又开"。
 */
const TOGGLE_SELECTOR = '[data-playlist-toggle]'

/** 时长格式化为 mm:ss；时长缺失时用占位符（与搜索表格的口径一致） */
const MS_PER_SECOND = 1000

function formatDuration(ms?: number): string {
  if (!ms || ms <= 0) return '--:--'
  const seconds = Math.floor(ms / MS_PER_SECOND)
  const minutes = Math.floor(seconds / 60)
  return `${String(minutes).padStart(2, '0')}:${String(seconds % 60).padStart(2, '0')}`
}

/**
 * 播放列表面板
 *
 * 由 MiniPlayer 的播放列表按钮唤出，贴窗口右侧展开成一条竖栏（顶栏下沿到播放栏上沿），
 * 展示当前播放队列：
 * 每行左边是歌曲条目（点封面切歌，正在播放的那首标题与封面按钮高亮），右边是时长。
 */
export function PlaylistPanel({ onClose }: PlaylistPanelProps) {
  const playlist = usePlayerStore((state) => state.playlist)
  const clearPlaylist = usePlayerStore((state) => state.clearPlaylist)
  const removeFromPlaylist = usePlayerStore((state) => state.removeFromPlaylist)
  const panelRef = useRef<HTMLDivElement>(null)

  // 点击面板外部关闭
  useEffect(() => {
    const onMouseDown = (event: MouseEvent) => {
      const target = event.target
      if (!(target instanceof Element)) return
      if (panelRef.current?.contains(target)) return
      if (target.closest(TOGGLE_SELECTOR)) return
      onClose()
    }
    document.addEventListener('mousedown', onMouseDown)
    return () => document.removeEventListener('mousedown', onMouseDown)
  }, [onClose])

  return (
    <div
      ref={panelRef}
      className="absolute inset-y-3 right-0 z-40 flex w-[380px] max-w-[80vw] flex-col overflow-hidden rounded-l-xl bg-[var(--color-background)] shadow-[-8px_0_24px_rgba(0,0,0,0.12)]"
    >
      {/* 标题栏：左标题 + 总数，右清空 */}
      <div className="flex items-center justify-between px-4 py-2">
        <div className="flex items-start gap-1">
          <h3 className="text-xl font-medium text-[var(--color-text-primary)]">播放列表</h3>
          <span className="text-[var(--color-text-secondary)]">{playlist.length}</span>
        </div>
        <button
          type="button"
          onClick={clearPlaylist}
          disabled={playlist.length === 0}
          className="flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 text-sm text-[var(--color-text-secondary)] transition-colors hover:bg-gray-100 disabled:cursor-not-allowed disabled:opacity-50"
          aria-label="清空播放列表"
          title="清空播放列表"
        >
          <Trash2 size={14} />
          清空
        </button>
      </div>

      {/* 列表：左歌曲条目、右时长 */}
      <div className="min-h-0 flex-1 overflow-y-auto">
        {playlist.length === 0 ? (
          <p className="px-4 py-8 text-center text-[var(--color-text-secondary)]">
            播放列表为空
          </p>
        ) : (
          playlist.map((song) => (
            <div
              key={`${song.id}:${song.hash ?? ''}`}
              className="group flex items-center gap-3 px-4 py-2 transition-colors hover:bg-gray-100"
            >
              <PlayableSong song={song} className="min-w-0 flex-1" markCurrent />
              {/* 时长与删除按钮共用同一位置：悬浮行时切换为删除按钮，不产生布局抖动 */}
              <div className="group/actions relative flex w-14 shrink-0 items-center justify-end">
                <span className="text-sm tabular-nums text-[var(--color-text-secondary)] transition-opacity group-hover:opacity-0 group-focus-within/actions:opacity-0">
                  {formatDuration(song.duration)}
                </span>
                <button
                  type="button"
                  onClick={() => removeFromPlaylist(song)}
                  className="absolute inset-y-0 right-0 flex cursor-pointer items-center rounded-md p-1 text-[var(--color-text-secondary)] opacity-0 transition-[color,opacity] hover:text-[var(--color-text-primary)] group-hover:opacity-100 group-focus-within/actions:opacity-100"
                  aria-label={`移出播放列表：${song.name}`}
                  title="移出播放列表"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
