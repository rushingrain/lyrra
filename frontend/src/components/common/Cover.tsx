import defaultCover from '@/assets/images/default_cover.jpg'

interface CoverProps {
  src?: string
  alt: string
  /** 尺寸类，默认 h-12 w-12 */
  className?: string
  /** 圆角风格：rounded 圆角 / circle 圆形 */
  variant?: 'rounded' | 'circle'
}

/** 通用封面图：空地址回退默认封面 */
export function Cover({ src, alt, className = 'h-12 w-12', variant = 'rounded' }: CoverProps) {
  return (
    <img
      src={src || defaultCover}
      alt={alt}
      className={`${className} shrink-0 ${variant === 'circle' ? 'rounded-full' : 'rounded-lg'} object-cover`}
      loading="lazy"
    />
  )
}
