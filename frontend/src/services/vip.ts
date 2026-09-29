import { ClaimDailyVip } from '../../wailsjs/go/vip/Service'
import type { vip } from '../../wailsjs/go/models'

/**
 * 每日 VIP 接口（Wails 绑定 window.go.vip.Service.*）。
 *
 * 服务层只负责取数据，动画之类的副作用放在 hooks 里（见 @/hooks/useConfetti）。
 *
 * 领取与升级都由后端做风控：当天已经成功过就直接返回本地记录、不再请求酷狗
 * （DailyVipResult.claimed 为 false 表示这次没走网络）。所以重复调用是安全的，
 * 前端不需要自己做去重；「重复领取」对调用方同样算成功。
 */

export type DailyVipResult = vip.DailyVipResult

/** 领取当天 VIP（后端固定按北京时间当天领取，无需参数） */
export async function claimDailyVip(): Promise<DailyVipResult> {
  return ClaimDailyVip()
}
