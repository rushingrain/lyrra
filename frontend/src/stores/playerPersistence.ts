import { usePlayerStore } from '@/stores/playerStore'
import toast from 'react-hot-toast'
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
 * - 设备数据：播放模式、音量、音质偏好；
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
let deviceSaveInFlight = false
let pendingAccount: PlayerAccountState | null = null
/** 最近一次提交（或刚恢复）的内容指纹，用于跳过重复写盘 */
let savedDevice = ''
let savedAccount = ''

function devicePayload(device: PlayerDeviceState): string {
  return `${device.playMode}:${device.volume}:${device.qualityPreference}`
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
  const beforeRestore = usePlayerStore.getState()

  usePlayerStore.subscribe((state, prev) => {
    // 先恢复整份设备配置，避免启动时的默认值覆盖磁盘上的其它字段。
    if (state.deviceSettingsStatus === 'ready' && (
      prev.deviceSettingsStatus !== 'ready'
      || state.playMode !== prev.playMode
      || state.volume !== prev.volume
      || state.qualityPreference !== prev.qualityPreference
    )) {
      scheduleDeviceSave({
        playMode: state.playMode,
        volume: state.volume,
        qualityPreference: state.qualityPreference,
      })
      // 音质是离散选项，选中后立即提交；音量仍沿用防抖。
      if (state.qualityPreference !== prev.qualityPreference) void flushDevice()
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
      const current = usePlayerStore.getState()
      current.restoreState({
        ...snapshot,
        device: {
          ...snapshot.device,
          // 恢复期间已手动调整的模式、音量优先，不被旧值覆盖。
          playMode: current.playMode !== beforeRestore.playMode ? current.playMode : snapshot.device.playMode,
          volume: current.volume !== beforeRestore.volume ? current.volume : snapshot.device.volume,
        },
      })
    })
    .catch((error) => {
      usePlayerStore.setState({ deviceSettingsStatus: 'error' })
      console.warn('[player] 本地播放状态恢复失败', error)
    })
}

function scheduleDeviceSave(device: PlayerDeviceState): void {
  pendingDevice = device
  if (deviceTimer !== null) clearTimeout(deviceTimer)
  deviceTimer = setTimeout(() => void flushDevice(), SAVE_DEBOUNCE_MS)
}

function scheduleAccountSave(account: PlayerAccountState): void {
  pendingAccount = account
  if (accountTimer !== null) clearTimeout(accountTimer)
  accountTimer = setTimeout(flushAccount, SAVE_DEBOUNCE_MS)
}

async function flushDevice(): Promise<void> {
  if (deviceTimer !== null) {
    clearTimeout(deviceTimer)
    deviceTimer = null
  }
  if (deviceSaveInFlight) return
  deviceSaveInFlight = true
  try {
    // 所有设备字段串行写入同一个文件，避免旧的音量写入覆盖新音质。
    while (pendingDevice) {
      const device = pendingDevice
      pendingDevice = null
      const payload = devicePayload(device)
      if (payload === savedDevice) continue
      try {
        await savePlayerDeviceState(device)
        savedDevice = payload
      } catch (error) {
        savedDevice = ''
        // 保留最新快照，下一次变更或关窗时重试。
        pendingDevice ??= device
        console.warn('[player] 播放设置保存失败', error)
        toast.error('播放设置保存失败，请重试', { id: 'player-device-save-error' })
        return
      }
    }
  } finally {
    deviceSaveInFlight = false
  }
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
  void flushDevice()
  flushAccount()
}
