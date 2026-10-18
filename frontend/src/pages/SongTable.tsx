import { useEffect, useState } from 'react'
import { ResultTable } from '@/components/common/ResultTable'
import { songColumns } from '@/components/search/songColumns'
import {
  DEFAULT_SONG_PAGE,
  DEFAULT_SONG_PAGE_SIZE,
  getDailyRecommendSongs,
  getHotSongs,
  getNewSongs,
} from '@/services/song'
import { usePlayerStore } from '@/stores/playerStore'
import { Play } from 'lucide-react'
import type { Song } from '@/types'

export type SongTableKind = 'new' | 'hot' | 'daily'

const TITLES: Record<SongTableKind, string> = {
  new: '最新音乐',
  hot: '热门音乐',
  daily: '每日推荐',
}

interface SongTableProps {
  kind: SongTableKind
}

/**
 * 歌曲表格页：最新音乐 / 热门音乐 / 每日推荐共用一套表格与播放全部逻辑，数据源由路由决定。
 * 表格复用搜索结果的列配置与可播放行，交互与搜索页保持一致。
 */
export default function SongTable({ kind }: SongTableProps) {
  const [songs, setSongs] = useState<Song[]>([])
  const playAll = usePlayerStore((state) => state.playAll)

  useEffect(() => {
    let cancelled = false
    setSongs([])
    // 与首页使用同一组请求参数：song 服务按 key 缓存 1h，
    // 从首页跳转过来直接命中缓存，不会重复请求后端。
    const request =
      kind === 'new'
        ? getNewSongs(DEFAULT_SONG_PAGE, DEFAULT_SONG_PAGE_SIZE)
        : kind === 'hot'
          ? getHotSongs()
          : getDailyRecommendSongs()
    request.then((data) => {
      if (!cancelled) setSongs(data)
    })
    return () => {
      cancelled = true
    }
  }, [kind])

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold text-[var(--color-text-primary)]">{TITLES[kind]}</h1>
        <button
          type="button"
          disabled={songs.length === 0}
          onClick={() => void playAll(songs)}
          className="flex cursor-pointer items-center gap-1 rounded-md bg-[var(--color-accent)] px-3 py-2 text-sm text-[var(--color-on-accent)] transition-colors hover:bg-[var(--color-accent-hover)] disabled:cursor-not-allowed disabled:opacity-50"
        >
          {/* Play 图标本身是描边图形，fill 后即为实心三角（播放栏的播放按钮同样处理） */}
          <Play size={16} fill="currentColor" />
          播放全部
        </button>
      </div>

      <ResultTable columns={songColumns} rows={songs} rowKey={(song) => song.id} />
    </div>
  )
}
