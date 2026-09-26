import type { Banner, Playlist, Song, Artist, PlayHistory } from '@/types'

export const mockBanners: Banner[] = [
  {
    id: '1',
    title: '热门推荐 - 华语经典',
    imageUrl: 'https://picsum.photos/seed/banner1/800/240',
    linkUrl: '/playlist/1',
  },
  {
    id: '2',
    title: '新歌首发 - 本周新曲',
    imageUrl: 'https://picsum.photos/seed/banner2/800/240',
    linkUrl: '/playlist/2',
  },
  {
    id: '3',
    title: '热门歌手 - 周杰伦特辑',
    imageUrl: 'https://picsum.photos/seed/banner3/800/240',
    linkUrl: '/artist/1',
  },
  {
    id: '4',
    title: '怀旧金曲 - 90年代精选',
    imageUrl: 'https://picsum.photos/seed/banner4/800/240',
    linkUrl: '/playlist/3',
  },
]

export const mockPlaylists: Playlist[] = [
  {
    id: 1,
    name: '心动rap：借夏天的名义，暗恋变成了热恋',
    coverUrl: 'https://picsum.photos/seed/playlist1/200/200',
    description: '收录了华语乐坛最经典的百首金曲，从邓丽君到周杰伦，跨越时代的音乐记忆',
    creator: '周杰伦',
    playCount: 12530000,
    songCount: 100,
  },
  {
    id: 2,
    name: '抖音热门BGM',
    coverUrl: 'https://picsum.photos/seed/playlist2/200/200',
    description: '全网最火的短视频背景音乐合集，时刻更新',
    creator: '抖音音乐',
    playCount: 8920000,
    songCount: 50,
  },
  {
    id: 3,
    name: '睡前轻音乐',
    coverUrl: 'https://picsum.photos/seed/playlist3/200/200',
    description: '舒缓的钢琴与自然白噪音，伴你入眠',
    creator: '抖音音乐',
    playCount: 4560000,
    songCount: 60,
  },
  {
    id: 4,
    name: '电子音乐精选',
    coverUrl: 'https://picsum.photos/seed/playlist4/200/200',
    description: '最燃的电子音乐，适合运动和驾驶',
    creator: '抖音音乐',
    playCount: 6780000,
    songCount: 80,
  },
  {
    id: 5,
    name: '民谣在路上',
    coverUrl: 'https://picsum.photos/seed/playlist5/200/200',
    description: '诗和远方都在民谣里，一路走一路听',
    creator: '抖音音乐',
    playCount: 3210000,
    songCount: 45,
  },
  {
    id: 6,
    name: '新歌推荐',
    coverUrl: 'https://picsum.photos/seed/playlist6/200/200',
    description: '每周最新上架的热门单曲',
    creator: '抖音音乐',
    playCount: 2340000,
    songCount: 30,
  },
  {
    id: 7,
    name: '说唱集结号',
    coverUrl: 'https://picsum.photos/seed/playlist7/200/200',
    description: '中文说唱最强阵容，态度与节奏并存',
    creator: '抖音音乐',
    playCount: 5430000,
    songCount: 55,
  },
  {
    id: 8,
    name: '爵士咖啡馆',
    coverUrl: 'https://picsum.photos/seed/playlist8/200/200',
    description: '慵懒午后，一杯咖啡配爵士，刚刚好',
    creator: '抖音音乐',
    playCount: 1890000,
    songCount: 40,
  },
  {
    id: 9,
    name: '摇滚经典永流传',
    coverUrl: 'https://picsum.photos/seed/playlist9/200/200',
    description: 'Beyond、黑豹、唐朝…中国摇滚的黄金年代',
    creator: '抖音音乐',
    playCount: 7650000,
    songCount: 70,
  },
  {
    id: 10,
    name: '古风雅韵',
    coverUrl: 'https://picsum.photos/seed/playlist10/200/200',
    description: '琴瑟和鸣，诗词入曲，品味千年国风之美',
    creator: '抖音音乐',
    playCount: 2870000,
    songCount: 38,
  },
  {
    id: 11,
    name: '欧美流行前线',
    coverUrl: 'https://picsum.photos/seed/playlist11/200/200',
    description: 'Billboard热单全收录，国际流行风向标',
    creator: '抖音音乐',
    playCount: 9340000,
    songCount: 65,
  },
  {
    id: 12,
    name: '治愈系纯音乐',
    coverUrl: 'https://picsum.photos/seed/playlist12/200/200',
    description: '用旋律治愈疲惫的心灵，找回内心的平静',
    creator: '抖音音乐',
    playCount: 4120000,
    songCount: 42,
  },
  {
    id: 13,
    name: 'KPOP热力全开',
    coverUrl: 'https://picsum.photos/seed/playlist13/200/200',
    description: 'BTS、BLACKPINK、NewJeans…韩流最前线',
    creator: '抖音音乐',
    playCount: 15670000,
    songCount: 55,
  },
  {
    id: 14,
    name: '经典粤语金曲',
    coverUrl: 'https://picsum.photos/seed/playlist14/200/200',
    description: '张国荣、梅艳芳、陈奕迅…港乐传奇永不褪色',
    creator: '抖音音乐',
    playCount: 8230000,
    songCount: 48,
  },
  {
    id: 15,
    name: '运动健身燃脂歌单',
    coverUrl: 'https://picsum.photos/seed/playlist15/200/200',
    description: '高BPM节奏引爆运动潜力，汗水与音乐一起燃烧',
    creator: '抖音音乐',
    playCount: 5460000,
    songCount: 35,
  },
]

export const mockSongs: Song[] = [
  {
    id: 1,
    title: '晴天',
    artist: '周杰伦',
    album: '叶惠美',
    coverUrl: 'https://picsum.photos/seed/song1/72/72',
    duration: 269,
    hash: 'a1b2c3d4e5f6a1b2c3d4e5f600000001',
    quality: 'Hi-Res',
    isOriginal: true,
  },
  {
    id: 2,
    title: '七里香',
    artist: '周杰伦',
    album: '七里香',
    coverUrl: 'https://picsum.photos/seed/song2/72/72',
    duration: 299,
    hash: 'a1b2c3d4e5f6a1b2c3d4e5f600000002',
    quality: '无损',
    isOriginal: true,
  },
  {
    id: 3,
    title: '起风了',
    artist: '买辣椒也用券',
    album: '起风了',
    coverUrl: 'https://picsum.photos/seed/song3/72/72',
    duration: 325,
    hash: 'a1b2c3d4e5f6a1b2c3d4e5f600000003',
    quality: 'Hi-Res',
  },
  {
    id: 4,
    title: '平凡之路',
    artist: '朴树',
    album: '平凡之路',
    coverUrl: 'https://picsum.photos/seed/song4/72/72',
    duration: 312,
    hash: 'a1b2c3d4e5f6a1b2c3d4e5f600000004',
    isOriginal: true,
  },
  {
    id: 5,
    title: '孤勇者',
    artist: '陈奕迅',
    album: '孤勇者',
    coverUrl: 'https://picsum.photos/seed/song5/72/72',
    duration: 238,
    hash: 'a1b2c3d4e5f6a1b2c3d4e5f600000005',
    quality: '无损',
  },
  {
    id: 6,
    title: '光年之外',
    artist: '邓紫棋',
    album: '光年之外',
    coverUrl: 'https://picsum.photos/seed/song6/72/72',
    duration: 243,
    hash: 'a1b2c3d4e5f6a1b2c3d4e5f600000006',
    quality: 'Hi-Res',
    isOriginal: true,
  },
  {
    id: 7,
    title: '江南',
    artist: '林俊杰',
    album: '第二天堂',
    coverUrl: 'https://picsum.photos/seed/song7/72/72',
    duration: 280,
    hash: 'a1b2c3d4e5f6a1b2c3d4e5f600000007',
    isOriginal: true,
  },
  {
    id: 8,
    title: '演员',
    artist: '薛之谦',
    album: '演员',
    coverUrl: 'https://picsum.photos/seed/song8/72/72',
    duration: 256,
    hash: 'a1b2c3d4e5f6a1b2c3d4e5f600000008',
  },
  {
    id: 9,
    title: '告白气球',
    artist: '周杰伦',
    album: '周杰伦的床边故事',
    coverUrl: 'https://picsum.photos/seed/song9/72/72',
    duration: 205,
    hash: 'a1b2c3d4e5f6a1b2c3d4e5f600000009',
    quality: 'Hi-Res',
    isOriginal: true,
  },
  {
    id: 10,
    title: '夜曲',
    artist: '周杰伦',
    album: '十一月的萧邦',
    coverUrl: 'https://picsum.photos/seed/song10/72/72',
    duration: 233,
    hash: 'a1b2c3d4e5f6a1b2c3d4e5f600000010',
    quality: '无损',
  },
]

export const mockArtists: Artist[] = [
  {
    id: 1,
    name: '周杰伦',
    songCount: 820,
    avatarUrl: 'https://picsum.photos/seed/artist1/160/160',
    genre: '华语流行',
    fans: 58000000,
  },
  {
    id: 2,
    name: '林俊杰',
    songCount: 412,
    avatarUrl: 'https://picsum.photos/seed/artist2/160/160',
    genre: '华语流行',
    fans: 42000000,
  },
  {
    id: 3,
    name: '陈奕迅',
    songCount: 389,
    avatarUrl: 'https://picsum.photos/seed/artist3/160/160',
    genre: '粤语流行',
    fans: 35000000,
  },
  {
    id: 4,
    name: '邓紫棋',
    songCount: 256,
    avatarUrl: 'https://picsum.photos/seed/artist4/160/160',
    genre: '华语流行',
    fans: 31000000,
  },
  {
    id: 5,
    name: '薛之谦',
    songCount: 298,
    avatarUrl: 'https://picsum.photos/seed/artist5/160/160',
    genre: '华语流行',
    fans: 28000000,
  },
  {
    id: 6,
    name: '朴树',
    songCount: 120,
    avatarUrl: 'https://picsum.photos/seed/artist6/160/160',
    genre: '民谣摇滚',
    fans: 15000000,
  },
  {
    id: 7,
    name: '蔡徐坤',
    songCount: 145,
    avatarUrl: 'https://picsum.photos/seed/artist7/160/160',
    genre: '流行偶像',
    fans: 22000000,
  },
  {
    id: 8,
    name: '汪苏泷',
    songCount: 210,
    avatarUrl: 'https://picsum.photos/seed/artist8/160/160',
    genre: '华语流行',
    fans: 18000000,
  },
]

export const mockPlayHistory: PlayHistory[] = [
  {
    song: mockSongs[0],
    playedAt: '2026-07-30T22:15:00Z',
  },
]

// ===== 综合搜索 mock 数据（后续接入真实接口后替换） =====

export interface MockComplexArtist {
  id: string
  name: string
  avatarUrl: string
  songCount: number
  fans: number
}

export interface MockAlbum {
  id: string
  name: string
  artist: string
  coverUrl: string
  date: string
}

export interface MockComplexSong {
  id: string
  title: string
  artist: string
  coverUrl: string
  quality?: string
  isOriginal?: boolean
  /** 后端以字符串标签形式返回的附加信息，如 "TOP500 第29名"、"评论过万" */
  tagContent: string[]
}

export interface MockMv {
  id: string
  name: string
  artist: string
  coverUrl: string
}

export interface MockSearchUser {
  id: string
  nickname: string
  signature: string
  avatarUrl: string
  gender?: 'male' | 'female'
}

export const mockComplexArtists: MockComplexArtist[] = [
  { id: 'ca1', name: '孙燕姿', avatarUrl: 'https://picsum.photos/seed/artist1/200/200', songCount: 606, fans: 5406000 },
  { id: 'ca2', name: '周杰伦', avatarUrl: 'https://picsum.photos/seed/artist2/200/200', songCount: 820, fans: 58000000 },
  { id: 'ca3', name: '林俊杰', avatarUrl: 'https://picsum.photos/seed/artist3/200/200', songCount: 412, fans: 42000000 },
  { id: 'ca4', name: '陈奕迅', avatarUrl: 'https://picsum.photos/seed/artist4/200/200', songCount: 389, fans: 35000000 },
  { id: 'ca5', name: '邓紫棋', avatarUrl: 'https://picsum.photos/seed/artist5/200/200', songCount: 256, fans: 31000000 },
  { id: 'ca6', name: '薛之谦', avatarUrl: 'https://picsum.photos/seed/artist6/200/200', songCount: 298, fans: 28000000 },
]

export const mockAlbums: MockAlbum[] = [
  { id: 'al1', name: '是时候', artist: '孙燕姿', coverUrl: 'https://picsum.photos/seed/album1/200/200', date: '2011-03-08' },
  { id: 'al2', name: '叶惠美', artist: '周杰伦', coverUrl: 'https://picsum.photos/seed/album2/200/200', date: '2003-07-31' },
  { id: 'al3', name: '编号89757', artist: '林俊杰', coverUrl: 'https://picsum.photos/seed/album3/200/200', date: '2005-04-01' },
  { id: 'al4', name: 'U87', artist: '陈奕迅', coverUrl: 'https://picsum.photos/seed/album4/200/200', date: '2005-06-07' },
  { id: 'al5', name: '新的心跳', artist: '邓紫棋', coverUrl: 'https://picsum.photos/seed/album5/200/200', date: '2015-11-06' },
  { id: 'al6', name: '初学者', artist: '薛之谦', coverUrl: 'https://picsum.photos/seed/album6/200/200', date: '2016-07-18' },
]

export const mockComplexSongs: MockComplexSong[] = [
  { id: 'cs1', title: '是时候', artist: '孙燕姿', coverUrl: 'https://picsum.photos/seed/song1/200/200', quality: 'Hi-Res', isOriginal: true, tagContent: ['TOP500 第29名', '评论过万'] },
  { id: 'cs2', title: '逆光', artist: '孙燕姿', coverUrl: 'https://picsum.photos/seed/song2/200/200', quality: 'Hi-Res', tagContent: ['评论过万'] },
  { id: 'cs3', title: '晴天', artist: '周杰伦', coverUrl: 'https://picsum.photos/seed/song3/200/200', quality: '无损', isOriginal: true, tagContent: ['TOP500 第1名', '万人收藏'] },
  { id: 'cs4', title: '七里香', artist: '周杰伦', coverUrl: 'https://picsum.photos/seed/song4/200/200', quality: 'Hi-Res', isOriginal: true, tagContent: ['TOP500 第3名'] },
  { id: 'cs5', title: '江南', artist: '林俊杰', coverUrl: 'https://picsum.photos/seed/song5/200/200', quality: '无损', tagContent: ['评论过万'] },
  { id: 'cs6', title: '十年', artist: '陈奕迅', coverUrl: 'https://picsum.photos/seed/song6/200/200', quality: 'Hi-Res', isOriginal: true, tagContent: ['TOP500 第16名', '万人收藏'] },
]

export const mockMvs: MockMv[] = [
  { id: 'mv1', name: '是时候 MV', artist: '孙燕姿', coverUrl: 'https://picsum.photos/seed/mv1/400/225' },
  { id: 'mv2', name: '晴天 MV', artist: '周杰伦', coverUrl: 'https://picsum.photos/seed/mv2/400/225' },
  { id: 'mv3', name: '江南 MV', artist: '林俊杰', coverUrl: 'https://picsum.photos/seed/mv3/400/225' },
  { id: 'mv4', name: '富士山下 MV', artist: '陈奕迅', coverUrl: 'https://picsum.photos/seed/mv4/400/225' },
  { id: 'mv5', name: '光年之外 MV', artist: '邓紫棋', coverUrl: 'https://picsum.photos/seed/mv5/400/225' },
  { id: 'mv6', name: '演员 MV', artist: '薛之谦', coverUrl: 'https://picsum.photos/seed/mv6/400/225' },
]

export const mockSearchUsers: MockSearchUser[] = [
  { id: 'u1', nickname: '燕姿后援会', signature: '每一天都是是时候', avatarUrl: 'https://picsum.photos/seed/user1/200/200', gender: 'female' },
  { id: 'u2', nickname: '杰伦的小可爱', signature: '想听你的演唱会', avatarUrl: 'https://picsum.photos/seed/user2/200/200', gender: 'female' },
  { id: 'u3', nickname: 'JMの城堡', signature: '林距离，同林俊杰', avatarUrl: 'https://picsum.photos/seed/user3/200/200', gender: 'male' },
  { id: 'u4', nickname: 'Eason粉丝团', signature: '十年之前我不认识你', avatarUrl: 'https://picsum.photos/seed/user4/200/200', gender: 'male' },
  { id: 'u5', nickname: 'G.E.M.歌迷会', signature: '光年之外等你', avatarUrl: 'https://picsum.photos/seed/user5/200/200', gender: 'female' },
  { id: 'u6', nickname: '谦谦的星河', signature: '我好像在哪见过你', avatarUrl: 'https://picsum.photos/seed/user6/200/200', gender: 'male' },
]
