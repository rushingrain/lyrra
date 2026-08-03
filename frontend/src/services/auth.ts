import type { LoginResult, QrLoginResponse, QrLoginStatus } from '@/types'
// mock 二维码占位图：仅用于纯浏览器开发（window.go 不存在时），Wails 环境下不生效
import qrCodeMockImage from '@/assets/images/qr_code.png'
import {
  CheckQrLoginStatus,
  GetCurrentUser,
  GetLoginQrCode,
  Logout,
} from '../../wailsjs/go/auth/AuthService'

/**
 * 登录相关接口
 *
 * 二维码登录已接入后端（Wails 绑定 window.go.auth.AuthService.*）。
 * 手机号登录（sendSmsCode / loginWithSmsCode）暂未接入后端，保持 mock 实现。
 *
 * 关于「只跑前端 npm run dev」：
 * Wails 的 window.go 运行时只在 Wails WebView 中注入，普通浏览器里不存在。
 * 因此真实登录流程必须用 `wails dev` 启动（前端代码会走 GetLoginQrCode 等真实绑定）。
 * 此处做 isBackendAvailable 检测：浏览器环境下自动回退到 mock，
 * 便于纯前端 UI 开发时登录弹窗不崩溃、能看到假二维码和 pending 状态。
 */

/** 二维码有效期（毫秒），与后端 5 分钟过期保持一致，仅 mock 使用 */
const QR_EXPIRES_MS = 5 * 60 * 1000

/** Wails 运行时是否可用（真实后端是否存在） */
function isBackendAvailable(): boolean {
  return typeof window !== 'undefined' && Boolean((window as any).go?.auth?.AuthService)
}

/**
 * 获取扫码登录二维码
 * 后端返回 qrKey（轮询用）+ qrUrl（data:image/png;base64,... 图片地址）+ 过期时间
 */
export async function getLoginQrCode(): Promise<QrLoginResponse> {
  if (!isBackendAvailable()) {
    // 纯浏览器 mock：返回本地占位二维码
    return {
      qrKey: `mock-qr-${Date.now()}`,
      qrUrl: qrCodeMockImage,
      expiresAt: Date.now() + QR_EXPIRES_MS,
    }
  }
  return GetLoginQrCode()
}

/**
 * 轮询查询扫码登录状态（单次查询，由调用方按固定间隔轮询）
 * @param qrKey 二维码唯一标识（来自 getLoginQrCode）
 * @returns 'pending' | 'scanned' | 'confirmed' | 'expired'
 */
export async function checkQrLoginStatus(qrKey: string): Promise<QrLoginStatus> {
  if (!isBackendAvailable()) {
    // 纯浏览器 mock：始终返回未扫码，由前端在二维码过期时切换为 'expired'
    return qrKey ? 'pending' : 'expired'
  }
  // 后端绑定返回 string（Wails 绑定生成器不产出命名 string 类型），这里断言为联合类型
  return (await CheckQrLoginStatus(qrKey)) as QrLoginStatus
}

/**
 * 获取当前登录用户信息
 * 应用启动时用于恢复会话；扫码确认后用于回填用户信息。
 * @returns 未登录 / 会话过期时返回 null
 */
export async function getCurrentUser(): Promise<LoginResult['user'] | null> {
  if (!isBackendAvailable()) {
    // 纯浏览器 mock：返回固定测试用户
    return {
      nickname: '测试用户',
      avatar: 'https://imge.kugou.com/common/3_0_0.jpg',
      isVip: true,
      friends: 3,
      follows: 10,
      fans: 100,
    }
  }
  try {
    return await GetCurrentUser()
  } catch {
    return null
  }
}

/** 登出：清空后端登录态与本地会话 */
export async function logout(): Promise<void> {
  if (!isBackendAvailable()) return
  await Logout()
}

/**
 * 发送短信验证码
 * TODO: 手机号登录暂未接入后端
 * @param phone 手机号
 */
export async function sendSmsCode(phone: string): Promise<void> {
  // TODO: 接入后端后替换为：
  // await SendSmsCode(phone)
  void phone
  await new Promise((resolve) => setTimeout(resolve, 300))
}

/**
 * 手机号 + 短信验证码登录
 * TODO: 手机号登录暂未接入后端
 * @returns 登录成功后的用户信息与凭证
 */
export async function loginWithSmsCode(phone: string, code: string): Promise<LoginResult> {
  // TODO: 接入后端后替换为：
  // const data = await LoginWithSmsCode(phone, code)
  // return data
  // mock：模拟请求耗时后抛错，提示后端尚未接入
  await new Promise((resolve) => setTimeout(resolve, 600))
  throw new Error(`登录接口尚未接入（phone: ${phone}，code: ${code}）`)
}
