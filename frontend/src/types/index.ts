export interface Song {
  id: string
  title: string
  artist: string
  album: string
  coverUrl: string
  duration: number
  hash: string
  /** 最高音质标签，如 "Hi-Res"、"无损" */
  quality?: string
  /** 是否为原唱 */
  isOriginal?: boolean
  /** 歌曲标签，如 "昨日超万人播放"、"上周播放top1W"；首页数据源可能不返回，消费端需做空值守卫 */
  tags?: string[]
}

export interface Playlist {
  id: string
  name: string
  coverUrl: string
  description: string
  playCount: number
  songCount?: number
}

/** 用户歌单（侧边栏展示，来自 /v7/get_all_list） */
export interface UserPlaylist {
  id: string
  name: string
  songCount: number
  coverUrl: string
  /** 1=默认收藏，2=我喜欢，0=普通歌单 */
  isDefault: number
  /** 0=创建的歌单，1=收藏的歌单 */
  type: number
}

export interface Artist {
  id: string
  name: string
  avatarUrl: string
  genre: string
  followerCount: number
}

export interface Banner {
  id: string
  imageUrl: string
  title: string
  linkUrl: string
}

export interface PlayHistory {
  song: Song
  playedAt: string
}

/** 扫码登录 - 获取二维码的响应 */
export interface QrLoginResponse {
  /** 二维码唯一标识，用于轮询登录状态 */
  qrKey: string
  /** 二维码图片地址 */
  qrUrl: string
  /** 二维码过期时间戳(ms) */
  expiresAt: number
}

/** 扫码登录状态：未扫码 / 已扫码待确认 / 已确认 / 已过期 */
export type QrLoginStatus = 'pending' | 'scanned' | 'confirmed' | 'expired'

/** 登录成功后的用户信息与凭证 */
export interface LoginResult {
  token: string
  user: {
    nickname: string
    avatar: string
    isVip: boolean
    /** 好友 / 关注 / 粉丝数（来自 user_detail 接口，未返回时缺省） */
    friends?: number
    follows?: number
    fans?: number
  }
}
