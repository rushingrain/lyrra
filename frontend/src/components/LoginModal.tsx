import { useCallback, useEffect, useState } from 'react'
import { Loader2, RefreshCw, Smartphone, X } from 'lucide-react'
import toast from 'react-hot-toast'
import {
  checkQrLoginStatus,
  getCurrentUser,
  getLoginQrCode,
  loginWithSmsCode,
  sendSmsCode,
} from '@/services/auth'
import type { LoginResult, QrLoginResponse, QrLoginStatus } from '@/types'
// 左上角"切换到扫码登录"入口按钮固定使用的本地占位二维码，不请求后端
import qrCodeMockImage from '@/assets/images/qr_code.png'

/** 轮询扫码状态的间隔(ms) */
const POLL_INTERVAL = 3000

type LoginMode = 'phone' | 'qr'

/** 手机号 + 验证码登录表单 */
function PhoneLoginForm({ onSuccess }: { onSuccess?: (result: LoginResult) => void }) {
  const [phone, setPhone] = useState('')
  const [code, setCode] = useState('')
  const [countdown, setCountdown] = useState(0)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [codeSent, setCodeSent] = useState(false)

  const phoneValid = /^1[3-9]\d{9}$/.test(phone)
  const canSubmit = phoneValid && /^\d{4,6}$/.test(code)

  // 验证码发送倒计时
  useEffect(() => {
    if (countdown <= 0) return
    const timer = window.setTimeout(() => setCountdown((c) => c - 1), 1000)
    return () => window.clearTimeout(timer)
  }, [countdown])

  const handleSendCode = async () => {
    if (!phoneValid) {
      setError('请输入正确的手机号')
      return
    }
    setError('')
    try {
      await sendSmsCode(phone)
      setCodeSent(true)
      setCountdown(60)
    } catch {
      setError('验证码发送失败，请稍后重试')
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!canSubmit || submitting) return
    setError('')
    setSubmitting(true)
    try {
      const result = await loginWithSmsCode(phone, code)
      onSuccess?.(result)
    } catch (err) {
      setError(err instanceof Error ? err.message : '登录失败，请稍后重试')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      {/* 手机号 */}
      <div className="flex items-center gap-2 rounded-md border border-[var(--color-border)] px-3 py-2 transition-colors duration-200 focus-within:border-[var(--color-accent)]">
        <span className="shrink-0 text-sm text-[var(--color-text-secondary)]">+86</span>
        <span className="h-4 w-px bg-[var(--color-border)]" />
        <input
          type="tel"
          value={phone}
          onChange={(e) => setPhone(e.target.value.replace(/\D/g, '').slice(0, 11))}
          placeholder="请输入手机号"
          className="flex-1 border-none bg-transparent text-sm outline-none placeholder:text-[var(--color-text-secondary)]/50 text-[var(--color-text-primary)]"
        />
      </div>

      {/* 验证码 */}
      <div className="flex items-center gap-2 rounded-md border border-[var(--color-border)] px-3 py-2 transition-colors duration-200 focus-within:border-[var(--color-accent)]">
        <input
          type="text"
          inputMode="numeric"
          value={code}
          onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
          placeholder="请输入验证码"
          className="flex-1 border-none bg-transparent text-sm outline-none placeholder:text-[var(--color-text-secondary)]/50 text-[var(--color-text-primary)]"
        />
        <button
          type="button"
          onClick={handleSendCode}
          disabled={countdown > 0}
          className="shrink-0 cursor-pointer rounded-md px-2 py-1 text-xs text-[var(--color-accent)] transition-colors duration-200 hover:bg-[var(--color-accent-bg)] disabled:cursor-not-allowed disabled:text-[var(--color-text-secondary)]/50"
        >
          {countdown > 0 ? `${countdown}s 后重发` : codeSent ? '重新获取' : '获取验证码'}
        </button>
      </div>

      {/* 错误提示 */}
      {error && <p className="text-xs text-red-500">{error}</p>}

      {/* 登录按钮 */}
      <button
        type="submit"
        disabled={!canSubmit || submitting}
        className="flex w-full cursor-pointer items-center justify-center gap-2 rounded-md bg-[var(--color-accent)] py-2 text-sm font-medium text-white transition-colors duration-200 hover:bg-[var(--color-accent-hover)] disabled:cursor-not-allowed disabled:opacity-50"
      >
        {submitting && <Loader2 size={14} className="animate-spin" />}
        {submitting ? '登录中...' : '登录'}
      </button>

      <p className="text-center text-xs text-[var(--color-text-secondary)]">
        首次登录请先注册「酷狗音乐概念版」账号
      </p>
    </form>
  )
}

interface LoginModalProps {
  open: boolean
  onClose: () => void
  /** 登录成功回调：回填用户信息，由父组件更新登录态 */
  onLoginSuccess?: (user: LoginResult['user']) => void
}

export function LoginModal({ open, onClose, onLoginSuccess }: LoginModalProps) {
  const [mode, setMode] = useState<LoginMode>('phone')
  const [qrInfo, setQrInfo] = useState<QrLoginResponse | null>(null)
  const [loading, setLoading] = useState(false)
  const [status, setStatus] = useState<QrLoginStatus>('pending')

  /** 获取二维码（接入后端后，请求逻辑收敛在 services/auth 中） */
  const loadQrCode = useCallback(async () => {
    setLoading(true)
    setStatus('pending')
    try {
      const info = await getLoginQrCode()
      setQrInfo(info)
    } catch {
      toast.error('获取登录二维码失败，请稍后重试')
    } finally {
      setLoading(false)
    }
  }, [])

  // 打开弹窗时：默认手机号登录，并预加载二维码
  useEffect(() => {
    if (open) {
      setMode('phone')
      loadQrCode()
    }
  }, [open, loadQrCode])

  // 扫码模式下：二维码过期检测 + 轮询扫码状态
  useEffect(() => {
    if (!open || mode !== 'qr' || !qrInfo) return

    const tick = () => {
      if (qrInfo.expiresAt - Date.now() <= 0) {
        setStatus('expired')
      }
    }
    tick()
    const expiryTimer = window.setInterval(tick, 1000)

    // mock 阶段始终返回 pending，接入后端轮询接口后自动生效
    const pollTimer = window.setInterval(async () => {
      try {
        const next = await checkQrLoginStatus(qrInfo.qrKey)
        setStatus(next)
      } catch {
        // 请求失败时忽略，等待下一次轮询
      }
    }, POLL_INTERVAL)

    return () => {
      window.clearInterval(expiryTimer)
      window.clearInterval(pollTimer)
    }
  }, [open, mode, qrInfo])

  // 扫码确认后视为登录成功：获取用户信息 → 通知父组件更新登录态 → 关闭弹窗
  useEffect(() => {
    if (status !== 'confirmed') return
    getCurrentUser()
      .then((user) => {
        if (user) onLoginSuccess?.(user)
      })
      .finally(() => onClose())
  }, [status, onClose, onLoginSuccess])

  // ESC 关闭
  useEffect(() => {
    if (!open) return
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [open, onClose])

  if (!open) return null

  const expired = status === 'expired'
  const hintText = expired
    ? '二维码已失效，请刷新后重新扫码'
    : status === 'scanned'
      ? '扫码成功，请在手机上确认登录'
      : '请使用酷狗音乐概念版手机客户端扫码登录'

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="relative w-[360px] overflow-hidden rounded-2xl bg-[var(--color-background)] p-8 shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        {/* 左上角：二维码一角（直角朝上，仅显示对角线上半部分三角形，与弹窗边缘留出间距） */}
        {mode === 'phone' ? (
          <button
            type="button"
            onClick={() => setMode('qr')}
            aria-label="切换到扫码登录"
            title="二维码登陆"
            className="group absolute left-3 top-3 h-16 w-16 cursor-pointer overflow-hidden"
            style={{ clipPath: 'polygon(0 0, 0 100%, 100% 0)' }}
          >
            {/* 入口按钮固定使用本地占位二维码，不依赖后端，点击后才进入扫码模式 */}
            <img src={qrCodeMockImage} alt="" className="h-full w-full object-cover" />
            <span className="absolute inset-0 flex items-start justify-start bg-black/0 p-1 text-[10px] font-medium leading-none text-white opacity-0 transition-all duration-200 group-hover:bg-black/40 group-hover:opacity-100">
              二维码登陆
            </span>
          </button>
        ) : (
          <button
            type="button"
            onClick={() => setMode('phone')}
            aria-label="切换到手机号登录"
            title="手机号登录"
            className="absolute left-3 top-3 flex h-10 w-10 cursor-pointer items-center justify-center rounded-lg bg-[var(--color-background)] text-[var(--color-text-secondary)] transition-colors duration-200 hover:text-[var(--color-accent)]"
          >
            <Smartphone size={16} />
          </button>
        )}

        {/* 头部：居中标题 / 右侧关闭 */}
        <div className="relative mb-6 flex items-center justify-center">
          <h2 className="text-lg font-semibold text-[var(--color-text-primary)]">
            {mode === 'phone' ? '手机号登录' : '扫码登录'}
          </h2>
          <button
            type="button"
            className="absolute right-0 flex h-7 w-7 cursor-pointer items-center justify-center rounded-md text-[var(--color-text-secondary)] transition-colors duration-200 hover:bg-gray-100"
            onClick={onClose}
            aria-label="关闭"
          >
            <X size={16} />
          </button>
        </div>

        {/* 登录方式内容 */}
        {mode === 'phone' ? (
          <PhoneLoginForm />
        ) : (
          <>
            {/* 二维码 */}
            <div className="relative mx-auto h-[200px] w-[200px] overflow-hidden rounded-xl border border-[var(--color-border)] bg-[var(--color-surface)]">
              {loading || !qrInfo ? (
                <div className="flex h-full w-full items-center justify-center">
                  <Loader2 size={24} className="animate-spin text-[var(--color-text-secondary)]" />
                </div>
              ) : (
                <>
                  <img src={qrInfo.qrUrl} alt="登录二维码" className="h-full w-full object-cover" />
                  {expired && (
                    <div className="absolute inset-0 flex flex-col items-center justify-center gap-2 rounded-xl bg-white/80 backdrop-blur-sm">
                      <span className="text-sm text-[var(--color-text-secondary)]">二维码已过期</span>
                      <button
                        className="cursor-pointer rounded-md bg-[var(--color-accent)] px-3 py-1 text-xs font-medium text-white transition-colors duration-200 hover:bg-[var(--color-accent-hover)]"
                        onClick={loadQrCode}
                      >
                        点击刷新
                      </button>
                    </div>
                  )}
                </>
              )}
            </div>

            {/* 提示文案 */}
            <p className="mt-5 text-center text-sm text-[var(--color-text-secondary)]">{hintText}</p>

            {/* 刷新 */}
            {qrInfo && (
              <div className="mt-4 flex items-center justify-center gap-4 text-xs text-[var(--color-text-secondary)]">
                <button
                  className="flex cursor-pointer items-center gap-1 rounded-md px-2 py-1 transition-colors duration-200 hover:bg-gray-100 disabled:cursor-not-allowed disabled:opacity-50"
                  onClick={loadQrCode}
                  disabled={loading}
                >
                  <RefreshCw size={12} />
                  刷新
                </button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  )
}
