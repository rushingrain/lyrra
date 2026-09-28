import { GetSongPlayURL } from '../../wailsjs/go/play/Service'
import type { play } from '../../wailsjs/go/models'

/**
 * 播放相关接口
 *
 * 已接入后端（Wails 绑定 window.go.play.Service.*）：
 * 后端按当前播放品质偏好解析出播放地址，并校验返回的文件是否命中目标档位。
 * 目前只在 `wails dev` / 打包后的应用里使用（浏览器没有 window.go 运行时）。
 */

export type PlayURL = play.PlayURL

/** 播放品质档位（与后端 internal/play 的档位键一致） */
export type QualityLevel = 'res' | 'sq' | 'hq' | 'standard'

/** 档位 → 界面文案；与歌曲列表里的音质标签口径一致 */
const QUALITY_LABELS: Record<QualityLevel, string> = {
  res: 'Hi-Res',
  sq: '无损',
  hq: 'HQ',
  standard: '标准',
}

/**
 * 当前播放音质的展示文案。
 * degraded 表示上游把该档位降级成了别的文件（例如歌曲没有 Hi-Res 时 high 落到标准文件），
 * 这时标注一下，避免"显示 Hi-Res 实际放的是标准音质"造成误解。
 */
export function qualityLabel(level: string | null | undefined, degraded = false): string {
  const label = QUALITY_LABELS[level as QualityLevel]
  if (!label) return '音质'
  return degraded ? `${label} · 降级` : label
}

/**
 * 获取歌曲播放地址。
 * hash 传 Song.hash（后端按同一品质偏好解析出来的播放用文件 hash），
 * albumAudioID 传 Song.id（MixSongID / album_audio_id），可为 0。
 *
 * 返回的 degraded 表示上游把该档位降级成了别的文件（例如歌曲没有 Hi-Res 时 high 降级），
 * 地址仍可用，本轮按"静默照播"处理。
 */
export function getSongPlayURL(hash: string, albumAudioID: number): Promise<PlayURL> {
  return GetSongPlayURL(hash, albumAudioID)
}
