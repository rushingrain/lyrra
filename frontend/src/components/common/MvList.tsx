import type { Mv as MvType } from '@/types'
import { Grid } from '@/components/common/Grid'

interface MvListProps {
  mvs: MvType[]
}

/**
 * MV 列表：3:2 封面 + 左对齐文案，布局复用 Grid
 */
export function MvList({ mvs }: MvListProps) {
  return (
    <Grid
      items={mvs}
      getItemKey={(mv) => mv.id}
      getCover={(mv) => ({ src: mv.coverUrl, alt: mv.name })}
      aspect="3 / 2"
      coverClassName="rounded-lg"
      minWidth={154}
      maxWidth={280}
      ratio={0.14}
      getInfo={(mv) => ({ title: mv.name, subtitle: mv.artist })}
    />
  )
}
