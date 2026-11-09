import {
  GetPlayerState,
  SavePlayerAccountState,
  SavePlayerDeviceState,
} from '../../wailsjs/go/play/Service'
import { playerstate, type dto } from '../../wailsjs/go/models'
import type { Song } from '@/types'

/**
 * 播放器本地状态的读写（后端 window.go.play.Service.*）。
 *
 * 数据落在用户配置目录：
 * - 设备数据（播放模式、音量）：与账号无关，登出保留；
 * - 账号数据（播放队列、当前歌曲）：随登出删除，未登录时后端直接忽略写入。
 *
 * 播放地址与播放进度都不持久化：恢复只带回歌曲信息，播放时再向后端请求地址。
 * 写盘频率由调用方控制（见 stores/playerPersistence.ts 的防抖）。
 */

export type PlayerStateSnapshot = playerstate.State
export type PlayerDeviceState = playerstate.DeviceState

/** 账号级数据用前端自己的 Song 类型，写盘前再转成后端 DTO */
export interface PlayerAccountState {
  playlist: Song[]
  currentSong?: Song
}

/** 读取本地保存的播放状态（启动时调用一次） */
export function getPlayerState(): Promise<PlayerStateSnapshot> {
  return GetPlayerState()
}

/** 保存设备级设置（播放模式 / 音量） */
export function savePlayerDeviceState(device: PlayerDeviceState): Promise<void> {
  return SavePlayerDeviceState(device)
}

/** 保存账号级播放数据（播放队列 / 当前歌曲） */
export function savePlayerAccountState(account: PlayerAccountState): Promise<void> {
  return SavePlayerAccountState(
    new playerstate.AccountState({
      playlist: account.playlist.map(toDtoSong),
      currentSong: account.currentSong ? toDtoSong(account.currentSong) : undefined,
    }),
  )
}

/** 后端 dto.Song 的 isOriginal 是必填布尔，前端类型里是可选的，这里补上默认值 */
function toDtoSong(song: Song): dto.Song {
  return { ...song, isOriginal: song.isOriginal ?? false }
}
