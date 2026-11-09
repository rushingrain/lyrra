import type { UserPlaylist } from '@/types'
import { GetUserPlaylists } from '../../wailsjs/go/auth/Service'

/** 前端请求用户歌单的默认参数 */
export const DEFAULT_USER_PLAYLIST_PAGE = 1
export const DEFAULT_USER_PLAYLIST_PAGE_SIZE = 30

/**
 * 获取当前登录用户的歌单列表。
 * SideBar 常驻挂载，无需缓存；未登录时后端返回错误，由调用方决定是否提示。
 */
export async function getUserPlaylists(
  page = DEFAULT_USER_PLAYLIST_PAGE,
  pagesize = DEFAULT_USER_PLAYLIST_PAGE_SIZE,
): Promise<UserPlaylist[]> {
  return GetUserPlaylists(page, pagesize)
}
