import { ClaimDailyVip, GetDailyVipStatus, UpgradeDailyVip } from '../../wailsjs/go/vip/Service'
import type { vip } from '../../wailsjs/go/models'

/**
 * 每日 VIP 接口（Wails 绑定 window.go.vip.Service.*）。
 *
 * 服务层只负责取数据，动画之类的副作用放在 hooks 里（见 @/hooks/useConfetti）。
 *
 * 领取与升级都由后端做风控：当天已经成功过就直接返回本地记录、不再请求酷狗
 * 本地缺少领取记录或升级前，会查询服务端同步其它设备的操作。
 * DailyVipResult.claimed 为 false 表示没有发送领取请求，可能进行了只读查询。
 * 前端不需要判断领取前置条件；「重复领取」对调用方同样算成功。
 */

export type DailyVipResult = vip.DailyVipResult
export type UpgradeVipResult = vip.UpgradeVipResult

/** 领取当天 VIP（后端固定按北京时间当天领取，无需参数） */
export async function claimDailyVip(): Promise<DailyVipResult> {
  return ClaimDailyVip()
}

/** 查询当天领取状态，本地缺少记录时由后端核对服务端并同步 */
export async function getDailyVipStatus(): Promise<DailyVipResult> {
  return GetDailyVipStatus()
}

/** 升级当天领取的 VIP（后端校验领取状态并限制重复请求） */
export async function upgradeDailyVip(): Promise<UpgradeVipResult> {
  return UpgradeDailyVip()
}
