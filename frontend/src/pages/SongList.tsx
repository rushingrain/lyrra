import { useEffect, useState } from 'react'
import { Navigate, useParams } from 'react-router-dom'
import { PlayAllButton } from '@/components/business/PlayAllButton'
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
import type { Song } from '@/types'

const METAS = {
  new: { title: '最新音乐', desc: '新鲜出炉，别错过每一首新歌' },
  hot: { title: '热门音乐', desc: '高播放量的歌曲，看看大家都在听什么' },
  daily: { title: '每日推荐', desc: '根据你的音乐口味生成，每天更新' },
} as const

type SongListKind = keyof typeof METAS

function isSongListKind(value: string | undefined): value is SongListKind {
  return value === 'new' || value === 'hot' || value === 'daily'
}

function formatDayMonth(date: Date): string {
  const dd = String(date.getDate()).padStart(2, '0')
  const mm = String(date.getMonth() + 1).padStart(2, '0') // getMonth 是 0-11
  return `${dd} / ${mm}`
}

/**
 * 歌曲列表页（最新音乐 / 热门音乐 / 每日推荐）：
 * 标题、描述、播放全部各占一行，不使用封面。
 * 板块由路由参数 /songs/:kind 决定。
 */
export default function SongList() {
  const { kind } = useParams<{ kind: string }>()
  const [songs, setSongs] = useState<Song[]>([])
  const playAll = usePlayerStore((state) => state.playAll)

  useEffect(() => {
    if (!isSongListKind(kind)) return
    let cancelled = false
    setSongs([])
    // 与首页使用同一组请求参数：song 服务按 key 缓存 20 分钟，
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

  // 未知板块（如手工输入的 /songs/xxx）直接回首页
  if (!isSongListKind(kind)) return <Navigate to="/" replace />

  const { desc } = METAS[kind]
  const title =
    kind === 'daily'
      ? `${formatDayMonth(new Date())} ${METAS[kind].title}`
      : METAS[kind].title

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-3">
        <h1 className="text-2xl font-semibold text-[var(--color-text-primary)]">{title}</h1>
        <p className="text-sm text-[var(--color-text-secondary)]">{desc}</p>
        <div>
          <PlayAllButton onClick={() => void playAll(songs)} disabled={songs.length === 0} />
        </div>
      </div>

      <ResultTable columns={songColumns} rows={songs} rowKey={(song) => song.id} />
    </div>
  )
}
