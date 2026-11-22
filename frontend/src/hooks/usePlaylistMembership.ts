import { useEffect, useState } from 'react'
import { useAuthStore } from '@/stores/authStore'
import { useUserPlaylistStore } from '@/stores/userPlaylistStore'
import { getUserPlaylist, playlistError } from '@/services/userPlaylist'
import type { LoginResult, UserPlaylist } from '@/types'

interface MembershipResult {
  user: LoginResult['user']
  globalId: string
  revision: number
  playlist: UserPlaylist | null
  error: string | null
}

/** 直接打开详情/编辑路由时也向后端确认归属，不依赖侧边栏或路由 state。 */
export function usePlaylistMembership(globalId: string, watchChanges = true) {
  const user = useAuthStore((state) => state.user)
  const revision = useUserPlaylistStore((state) => watchChanges ? state.revisions[globalId] ?? 0 : 0)
  const [result, setResult] = useState<MembershipResult | null>(null)
  const [retry, setRetry] = useState(0)
  useEffect(() => {
    if (!user || !globalId) return
    let cancelled = false
    setResult(null)
    getUserPlaylist(globalId)
      .then((playlist) => { if (!cancelled) setResult({ user, globalId, revision, playlist, error: null }) })
      .catch((error) => { if (!cancelled) setResult({ user, globalId, revision, playlist: null, error: playlistError(error, '获取歌单信息失败') }) })
    return () => { cancelled = true }
  }, [user, globalId, revision, retry])
  const current = result?.user === user && result?.globalId === globalId && result?.revision === revision ? result : null
  return {
    playlist: current?.playlist ?? null,
    loading: Boolean(user && !current),
    error: current?.error ?? null,
    retry: () => setRetry((value) => value + 1),
  }
}
