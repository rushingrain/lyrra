import { usePlayerStore } from '@/stores/playerStore'
import {
  getPlayerState,
  savePlayerAccountState,
  savePlayerDeviceState,
  type PlayerAccountState,
  type PlayerDeviceState,
} from '@/services/playerState'

/**
 * 播放状态落盘
 *
 * 订阅播放器 store 的变化，把两类数据分别写回后端（后端再落 JSON 文件）：
 * - 设备数据：播放模式、音量；
 * - 账号数据：播放队列、当前歌曲。
 *
 * 写盘频率控制有三道：
 *   1. 这里做 1s 尾部防抖（音量拖动、连续切歌只在停下来后写一次）；
 *   2. 内容与"已提交内容"一致时直接跳过；
 *   3. 后端还会再比对一次最近写入的内容，相同就不落盘。
 *
 * 恢复只带回歌曲信息：播放地址不持久化，第一次播放时重新向后端请求。
 */

/** 落盘防抖时长(ms) */
const SAVE_DEBOUNCE_MS = 1000

let started = false
let deviceTimer: ReturnType<typeof setTimeout> | null = null
let accountTimer: ReturnType<typeof setTimeout> | null = null
let pendingDevice: PlayerDeviceState | null = null
let pendingAccount: PlayerAccountState | null = null
/** 最近一次提交（或刚恢复）的内容指纹，用于跳过重复写盘 */
let savedDevice = ''
let savedAccount = ''

function devicePayload(device: PlayerDeviceState): string {
  return `${device.playMode}:${device.volume}`
}

function accountPayload(account: PlayerAccountState): string {
  return JSON.stringify(account)
}

/** 当前 store 对应的账号数据快照 */
function currentAccountState(): PlayerAccountState {
  const { playlist, song } = usePlayerStore.getState()
  return { playlist, currentSong: song ?? undefined }
}

/**
 * 启动时恢复本地状态并订阅后续变更；重复调用只生效一次（StrictMode 下 effect 会跑两次）。
 */
export function initPlayerPersistence(): void {
  if (started) return
  started = true

  usePlayerStore.subscribe((state, prev) => {
    if (state.playMode !== prev.playMode || state.volume !== prev.volume) {
      scheduleDeviceSave({ playMode: state.playMode, volume: state.volume })
    }
    if (state.playlist !== prev.playlist || state.song !== prev.song) {
      scheduleAccountSave(currentAccountState())
    }
  })

  // 关窗前尽力补写一次：防抖窗口内直接退出也不至于丢数据
  window.addEventListener('pagehide', flushAll)

  void getPlayerState()
    .then((snapshot) => {
      // 先把"磁盘里已有的内容"记为基线，再恢复：
      // 这样恢复本身不会触发一次回写，而用户在恢复期间做过的改动（与基线不同）仍会被写盘。
      savedDevice = devicePayload(snapshot.device)
      savedAccount = accountPayload({
        playlist: snapshot.account.playlist ?? [],
        currentSong: snapshot.account.currentSong,
      })
      usePlayerStore.getState().restoreState(snapshot)
    })
    .catch((error) => {
      // 恢复失败不弹提示：播放器按默认值启动，用户仍可正常播放
      console.warn('[player] 本地播放状态恢复失败', error)
    })
}

function scheduleDeviceSave(device: PlayerDeviceState): void {
  pendingDevice = device
  if (deviceTimer !== null) clearTimeout(deviceTimer)
  deviceTimer = setTimeout(flushDevice, SAVE_DEBOUNCE_MS)
}

function scheduleAccountSave(account: PlayerAccountState): void {
  pendingAccount = account
  if (accountTimer !== null) clearTimeout(accountTimer)
  accountTimer = setTimeout(flushAccount, SAVE_DEBOUNCE_MS)
}

function flushDevice(): void {
  if (deviceTimer !== null) {
    clearTimeout(deviceTimer)
    deviceTimer = null
  }
  const device = pendingDevice
  pendingDevice = null
  if (!device) return

  const payload = devicePayload(device)
  if (payload === savedDevice) return
  savedDevice = payload
  void savePlayerDeviceState(device).catch((error) => {
    // 写失败不记账，下次变更时会重试
    savedDevice = ''
    console.warn('[player] 播放设置保存失败', error)
  })
}

function flushAccount(): void {
  if (accountTimer !== null) {
    clearTimeout(accountTimer)
    accountTimer = null
  }
  const account = pendingAccount
  pendingAccount = null
  if (!account) return

  const payload = accountPayload(account)
  if (payload === savedAccount) return
  savedAccount = payload
  void savePlayerAccountState(account).catch((error) => {
    savedAccount = ''
    console.warn('[player] 播放列表保存失败', error)
  })
}

function flushAll(): void {
  flushDevice()
  flushAccount()
}
