import type { Song, UserPlaylist } from '@/types'
import {
  GetUserPlaylists,
  GetUserPlaylist,
  CreateUserPlaylist,
  CollectPlaylist,
  DeleteUserPlaylist,
  UpdateUserPlaylist,
  AddPlaylistSong,
  RemovePlaylistSong,
} from '../../wailsjs/go/playlist/Service'

/** 前端请求用户歌单的默认参数 */
export const DEFAULT_USER_PLAYLIST_PAGE = 1
export const DEFAULT_USER_PLAYLIST_PAGE_SIZE = 150

/**
 * 获取当前登录用户的歌单列表。
 * 未登录时后端返回错误，由调用方决定是否提示。
 */
export async function getUserPlaylists(
  page = DEFAULT_USER_PLAYLIST_PAGE,
  pagesize = DEFAULT_USER_PLAYLIST_PAGE_SIZE,
): Promise<UserPlaylist[]> {
  return GetUserPlaylists(page, pagesize)
}

export function getUserPlaylist(globalId: string): Promise<UserPlaylist | null> {
  return GetUserPlaylist(globalId)
}

export const createUserPlaylist = (name: string) => CreateUserPlaylist(name)
export const collectPlaylist = (globalId: string) => CollectPlaylist(globalId)
export const deleteUserPlaylist = (globalId: string) => DeleteUserPlaylist(globalId)

export interface PlaylistEdit {
  name: string
  description: string
  tags: string[]
  coverData: string
}

export const updateUserPlaylist = (globalId: string, edit: PlaylistEdit) => UpdateUserPlaylist(globalId, edit)
export const addPlaylistSong = (globalId: string, song: Song) => AddPlaylistSong(globalId, { ...song, isOriginal: song.isOriginal ?? false })
export const removePlaylistSong = (globalId: string, fileId: string) => RemovePlaylistSong(globalId, fileId)

/** Wails 的拒绝值可能是字符串，统一保留后端的可读错误提示。 */
export function playlistError(error: unknown, fallback: string): string {
  return typeof error === 'string' && error ? error : error instanceof Error ? error.message : fallback
}
