import { create } from 'zustand'
import type { Song, UserPlaylist } from '@/types'
import { getUserPlaylists, playlistError } from '@/services/userPlaylist'
import { useAuthStore } from '@/stores/authStore'
import { invalidatePlaylistSongs } from '@/services/playlist'

interface UserPlaylistState {
  playlists: UserPlaylist[]
  loading: boolean
  error: string | null
  revisions: Record<string, number>
  createOpen: boolean
  songToCollect: Song | null
  load: () => Promise<void>
  reset: () => void
  changed: (globalId?: string) => void
  openCreate: () => void
  closeCreate: () => void
  collectSong: (song: Song | null) => void
}

let generation = 0

/** 侧边栏与歌曲收藏弹窗共用数据；修改后刷新，账号切换时丢弃旧请求。 */
export const useUserPlaylistStore = create<UserPlaylistState>((set, get) => ({
  playlists: [], loading: false, error: null, revisions: {}, createOpen: false, songToCollect: null,
  load: async () => {
    const user = useAuthStore.getState().user
    if (!user) return
    const request = ++generation
    set({ loading: true, error: null })
    try {
      const playlists = await getUserPlaylists()
      if (request === generation && useAuthStore.getState().user === user) set({ playlists })
    } catch (error) {
      if (request === generation && useAuthStore.getState().user === user) {
        set({ error: playlistError(error, '获取歌单失败，请稍后重试') })
      }
    } finally {
      if (request === generation) set({ loading: false })
    }
  },
  reset: () => {
    generation += 1
    invalidatePlaylistSongs()
    set({ playlists: [], loading: false, error: null, revisions: {}, createOpen: false, songToCollect: null })
  },
  changed: (globalId) => {
    invalidatePlaylistSongs(globalId)
    if (globalId) set((state) => ({ revisions: { ...state.revisions, [globalId]: (state.revisions[globalId] ?? 0) + 1 } }))
    void get().load()
  },
  openCreate: () => set({ createOpen: true }),
  closeCreate: () => set({ createOpen: false }),
  collectSong: (songToCollect) => set({ songToCollect }),
}))
