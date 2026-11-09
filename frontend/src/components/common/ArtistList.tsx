import type { Artist as ArtistType } from '@/types'
import { Grid } from '@/components/common/Grid'

interface ArtistListProps {
  artists: ArtistType[]
}

/**
 * 歌手列表：圆形头像 + 居中文案，布局复用 Grid
 */
export function ArtistList({ artists }: ArtistListProps) {
  return (
    <Grid
      items={artists}
      getItemKey={(artist) => artist.id}
      getCover={(artist) => ({
        src: artist.avatarUrl || 'src/assets/images/default_cover.jpg',
        alt: artist.name,
      })}
      aspect="1 / 1"
      coverClassName="rounded-full"
      minWidth={136}
      maxWidth={240}
      ratio={0.12}
      getInfo={(artist) => ({ title: artist.name, subtitle: `单曲: ${artist.songCount}` })}
      center
    />
  )
}
