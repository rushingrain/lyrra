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
}

export interface Playlist {
  id: string
  name: string
  coverUrl: string
  description: string
  playCount: number
  songCount: number
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
