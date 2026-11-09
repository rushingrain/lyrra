import defaultCover from '@/assets/images/default_cover.jpg'

interface NowPlayingBackdropProps {
  /** 当前歌曲封面；为空时回退默认封面 */
  coverUrl?: string
}

/**
 * 歌词展开时的沉浸背景：封面放大模糊 + 暗色遮罩，覆盖整个窗口（含播放栏区域）。
 *
 * 背景只画一层、且挂在根节点上，歌词内容与播放栏都浮在它上面，
 * 这样上下两段是同一张模糊图，不会在交界处出现接缝。
 */
export function NowPlayingBackdrop({ coverUrl }: NowPlayingBackdropProps) {
  return (
    <div
      className="immersive-fade-in pointer-events-none absolute inset-0 z-20 overflow-hidden"
      aria-hidden
    >
      <img src={coverUrl || defaultCover} alt="" className="h-full w-full scale-125 object-cover blur-2xl" />
      <div className="absolute inset-0 bg-black/60" />
    </div>
  )
}
