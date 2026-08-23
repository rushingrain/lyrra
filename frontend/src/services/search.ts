import { Search, SearchComplex } from '../../wailsjs/go/search/SearchService'

type ApiRecord = Record<string, unknown>

export interface ComplexSearchSong {
  id: string
  title: string
  artist: string
  coverUrl: string
  quality?: string
  isOriginal: boolean
  tags: string[]
}

export interface ComplexSearchPlaylist {
  id: string
  name: string
  coverUrl: string
  description: string
  playCount: number
}

export interface ComplexSearchAlbum {
  id: string
  name: string
  artist: string
  coverUrl: string
  date: string
}

export interface ComplexSearchArtist {
  id: string
  name: string
  avatarUrl: string
  songCount: number
  fans: number
}

export interface ComplexSearchMv {
  id: string
  name: string
  artist: string
  coverUrl: string
}

export interface SearchLyric {
  id: string
  title: string
  artist: string
  album: string
}

export interface SearchResult {
  songs: ComplexSearchSong[]
  playlists: ComplexSearchPlaylist[]
  albums: ComplexSearchAlbum[]
  artists: ComplexSearchArtist[]
  mvs: ComplexSearchMv[]
  lyrics: SearchLyric[]
}

export type NormalSearchType = 'song' | 'special' | 'album' | 'author' | 'mv' | 'lyric'
export type SearchTabType = 'complex' | NormalSearchType

const EMPTY_RESULT: SearchResult = {
  songs: [],
  playlists: [],
  albums: [],
  artists: [],
  mvs: [],
  lyrics: [],
}

const CACHE_LIMIT = 20
const cache = new Map<string, Promise<SearchResult>>()

function isBackendAvailable(): boolean {
  return typeof window !== 'undefined' && Boolean((window as any).go?.search?.SearchService)
}

function asRecord(value: unknown): ApiRecord | null {
  if (value && typeof value === 'object' && !Array.isArray(value)) return value as ApiRecord
  return null
}

function asRecords(value: unknown): ApiRecord[] {
  return Array.isArray(value)
    ? value.map(asRecord).filter((item): item is ApiRecord => item !== null)
    : []
}

function asString(value: unknown): string {
  return typeof value === 'string' ? value : value == null ? '' : String(value)
}

function asNumber(value: unknown): number {
  const number = typeof value === 'number' ? value : Number(value)
  return Number.isFinite(number) ? number : 0
}

function imageUrl(value: unknown): string {
  return asString(value).replaceAll('{size}', '240')
}

function firstValue(item: ApiRecord, ...keys: string[]): unknown {
  return keys.map((key) => item[key]).find((value) => value != null && value !== '')
}

function groupsByType(data: unknown): Map<string, ApiRecord[]> {
  const groups = asRecords(asRecord(data)?.lists)
  return new Map(groups.map((group) => [asString(group.type), asRecords(group.lists)]))
}

function normalSearchItems(data: unknown): ApiRecord[] {
  const record = asRecord(data)
  if (!record) return []

  for (const key of ['lists', 'info', 'list', 'data']) {
    const items = asRecords(record[key])
    if (items.length > 0) return items
  }
  return []
}

function mapSong(item: ApiRecord): ComplexSearchSong {
  const tagContent = asString(firstValue(item, 'TagContent', 'tag_content'))
  const sqBitrate = asNumber(firstValue(item, 'SQBitrate', 'sq_bitrate'))
  const hqBitrate = asNumber(firstValue(item, 'HQBitrate', 'hq_bitrate'))
  return {
    id: asString(firstValue(item, 'ID', 'MixSongID', 'FileHash', 'hash', 'audio_id')),
    title: asString(firstValue(item, 'SongName', 'FileName', 'songname', 'filename')),
    artist: asString(firstValue(item, 'SingerName', 'singername', 'singer')),
    coverUrl: imageUrl(firstValue(item, 'Image', 'img', 'imgurl', 'cover')),
    quality: sqBitrate > 0 ? 'SQ' : hqBitrate >= 320 ? 'HQ' : undefined,
    isOriginal: asNumber(item.IsOriginal) === 1,
    tags: tagContent ? tagContent.split(/[|,]/).filter(Boolean) : [],
  }
}

function mapPlaylist(item: ApiRecord): ComplexSearchPlaylist {
  return {
    id: asString(firstValue(item, 'specialid', 'special_id', 'gid', 'id')),
    name: asString(firstValue(item, 'specialname', 'special_name', 'name')),
    coverUrl: imageUrl(firstValue(item, 'img', 'imgurl', 'cover')),
    description: asString(firstValue(item, 'intro', 'nickname', 'description')),
    playCount: asNumber(firstValue(item, 'total_play_count', 'play_count')),
  }
}

function mapAlbum(item: ApiRecord): ComplexSearchAlbum {
  return {
    id: asString(firstValue(item, 'albumid', 'album_id', 'id')),
    name: asString(firstValue(item, 'albumname', 'album_name', 'title', 'name')),
    artist: asString(firstValue(item, 'singer', 'SingerName', 'singername')),
    coverUrl: imageUrl(firstValue(item, 'img', 'imgurl', 'cover')),
    date: asString(firstValue(item, 'publish_time', 'publishtime', 'publish_date')).slice(0, 10),
  }
}

function mapArtist(item: ApiRecord): ComplexSearchArtist {
  return {
    id: asString(firstValue(item, 'AuthorId', 'UserId', 'author_id', 'singerid', 'id')),
    name: asString(firstValue(item, 'AuthorName', 'author_name', 'singername', 'singer', 'name')),
    avatarUrl: imageUrl(firstValue(item, 'Avatar', 'img', 'imgurl', 'pic')),
    songCount: asNumber(firstValue(item, 'AudioCount', 'audio_count', 'song_count')),
    fans: asNumber(firstValue(item, 'FansNum', 'fans_num', 'fans')),
  }
}

function mapMv(item: ApiRecord): ComplexSearchMv {
  return {
    id: asString(firstValue(item, 'MvID', 'MvHash', 'mv_id', 'hash', 'id')),
    name: asString(firstValue(item, 'MvName', 'mvname', 'songname', 'filename', 'name')),
    artist: asString(firstValue(item, 'SingerName', 'singername', 'singer')),
    coverUrl: imageUrl(firstValue(item, 'Pic', 'ErectPic', 'img', 'imgurl', 'cover')),
  }
}

function mapLyric(item: ApiRecord): SearchLyric {
  return {
    id: asString(firstValue(item, 'id', 'ID', 'hash', 'FileHash')),
    title: asString(firstValue(item, 'songname', 'SongName', 'filename', 'FileName', 'name')),
    artist: asString(firstValue(item, 'singername', 'SingerName', 'singer')),
    album: asString(firstValue(item, 'album_name', 'albumname', 'album')),
  }
}

function mapComplexSearch(data: unknown): SearchResult {
  const groups = groupsByType(data)
  return {
    songs: (groups.get('song') ?? []).map(mapSong),
    playlists: (groups.get('collect') ?? []).map(mapPlaylist),
    albums: (groups.get('album') ?? []).map(mapAlbum),
    artists: (groups.get('author') ?? []).map(mapArtist),
    mvs: (groups.get('mv') ?? []).map(mapMv),
    lyrics: [],
  }
}

function mapNormalSearch(searchType: NormalSearchType, data: unknown): SearchResult {
  const items = normalSearchItems(data)
  const result: SearchResult = { ...EMPTY_RESULT }

  switch (searchType) {
    case 'song':
      result.songs = items.map(mapSong)
      break
    case 'special':
      result.playlists = items.map(mapPlaylist)
      break
    case 'album':
      result.albums = items.map(mapAlbum)
      break
    case 'author':
      result.artists = items.map(mapArtist)
      break
    case 'mv':
      result.mvs = items.map(mapMv)
      break
    case 'lyric':
      result.lyrics = items.map(mapLyric)
      break
  }
  return result
}

/** 获取综合搜索结果，并将 SDK 的动态响应转换为页面使用的字段。 */
export function getComplexSearch(keyword: string, pageNum = 1, pageSize = 10): Promise<SearchResult> {
  if (!isBackendAvailable()) return Promise.reject(new Error('搜索服务不可用'))

  const key = `complex:${keyword}:${pageNum}:${pageSize}`
  const cached = cache.get(key)
  if (cached) return cached

  const request = SearchComplex(keyword, pageNum, pageSize)
    .then((response) => mapComplexSearch(response.data))
    .catch((error) => {
      cache.delete(key)
      throw error
    })

  cache.set(key, request)
  if (cache.size > CACHE_LIMIT) {
    const oldestKey = cache.keys().next().value
    if (oldestKey) cache.delete(oldestKey)
  }
  return request
}

/** 根据搜索分类调用对应后端接口，并转换为页面所需的数据结构。 */
export function getSearch(keyword: string, searchType: SearchTabType, pageNum = 1, pageSize = 10): Promise<SearchResult> {
  if (searchType === 'complex') return getComplexSearch(keyword, pageNum, pageSize)
  if (!isBackendAvailable()) return Promise.reject(new Error('搜索服务不可用'))

  const key = `${searchType}:${keyword}:${pageNum}:${pageSize}`
  const cached = cache.get(key)
  if (cached) return cached

  const request = Search(keyword, pageNum, pageSize, searchType)
    .then((response) => mapNormalSearch(searchType, response.data))
    .catch((error) => {
      cache.delete(key)
      throw error
    })

  cache.set(key, request)
  if (cache.size > CACHE_LIMIT) {
    const oldestKey = cache.keys().next().value
    if (oldestKey) cache.delete(oldestKey)
  }
  return request
}

export { EMPTY_RESULT }
