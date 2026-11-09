import { useEffect, useState } from 'react'

const DEFAULT_COVER_COLOR = '#4B5563'
const SAMPLE_SIZE = 24

const colorCache = new Map<string, string>()
const colorInflight = new Map<string, Promise<string>>()

/** 从封面提取代表色，并限制亮度，确保白色歌单名始终可读。 */
export function useCoverColor(url: string, fallback = DEFAULT_COVER_COLOR): string {
  const [color, setColor] = useState(() => colorCache.get(url) ?? fallback)

  useEffect(() => {
    if (!url) {
      setColor(fallback)
      return
    }

    const cached = colorCache.get(url)
    if (cached) {
      setColor(cached)
      return
    }

    let active = true
    loadCoverColor(url)
      .then((nextColor) => {
        if (active) setColor(nextColor)
      })
      .catch(() => {
        if (active) setColor(fallback)
      })

    return () => {
      active = false
    }
  }, [url, fallback])

  return color
}

function loadCoverColor(url: string): Promise<string> {
  const cached = colorCache.get(url)
  if (cached) return Promise.resolve(cached)

  const existing = colorInflight.get(url)
  if (existing) return existing

  const request = new Promise<string>((resolve, reject) => {
    const image = new Image()
    image.crossOrigin = 'anonymous'
    image.decoding = 'async'
    image.onload = () => {
      try {
        resolve(extractRepresentativeColor(image))
      } catch (error) {
        reject(error)
      }
    }
    image.onerror = () => reject(new Error('封面取色失败'))
    image.src = url
  })
    .then((color) => {
      colorCache.set(url, color)
      return color
    })
    .finally(() => {
      colorInflight.delete(url)
    })

  colorInflight.set(url, request)
  return request
}

function extractRepresentativeColor(image: HTMLImageElement): string {
  const canvas = document.createElement('canvas')
  canvas.width = SAMPLE_SIZE
  canvas.height = SAMPLE_SIZE
  const context = canvas.getContext('2d', { willReadFrequently: true })
  if (!context) throw new Error('无法创建取色画布')

  context.drawImage(image, 0, 0, SAMPLE_SIZE, SAMPLE_SIZE)
  const pixels = context.getImageData(0, 0, SAMPLE_SIZE, SAMPLE_SIZE).data

  let red = 0
  let green = 0
  let blue = 0
  let totalWeight = 0

  for (let index = 0; index < pixels.length; index += 4) {
    if (pixels[index + 3] < 128) continue

    const redValue = pixels[index]
    const greenValue = pixels[index + 1]
    const blueValue = pixels[index + 2]
    const maxChannel = Math.max(redValue, greenValue, blueValue)
    const minChannel = Math.min(redValue, greenValue, blueValue)
    const saturation = maxChannel === 0 ? 0 : (maxChannel - minChannel) / maxChannel
    const brightness = maxChannel / 255

    // 降低过暗、过曝和低饱和像素的权重，避免黑边或白底主导取色。
    const visibility = Math.min(1, brightness * 4) * Math.min(1, (1 - brightness) * 6)
    const weight = visibility * (0.2 + saturation * 2)
    if (weight <= 0.01) continue

    red += redValue * weight
    green += greenValue * weight
    blue += blueValue * weight
    totalWeight += weight
  }

  if (totalWeight === 0) return DEFAULT_COVER_COLOR

  const { hue, saturation, lightness } = rgbToHsl(
    red / totalWeight,
    green / totalWeight,
    blue / totalWeight,
  )
  // 灰度封面保持中性色，避免低饱和度噪点被误判成某个色相。
  const backgroundColorSaturation = saturation < 0.08
    ? saturation
    : Math.min(0.72, Math.max(0.34, saturation * 1.15))
  const backgroundColorLightness = Math.min(0.42, Math.max(0.28, lightness * 0.72))
  const channels = hslToRgb(hue, backgroundColorSaturation, backgroundColorLightness)

  return rgbToHex(ensureWhiteTextContrast(channels))
}

function rgbToHsl(
  red: number,
  green: number,
  blue: number,
): { hue: number; saturation: number; lightness: number } {
  const redRatio = red / 255
  const greenRatio = green / 255
  const blueRatio = blue / 255
  const maxChannel = Math.max(redRatio, greenRatio, blueRatio)
  const minChannel = Math.min(redRatio, greenRatio, blueRatio)
  const lightness = (maxChannel + minChannel) / 2

  if (maxChannel === minChannel) return { hue: 0, saturation: 0, lightness }

  const delta = maxChannel - minChannel
  const saturation = lightness > 0.5
    ? delta / (2 - maxChannel - minChannel)
    : delta / (maxChannel + minChannel)
  let hue = 0
  if (maxChannel === redRatio) {
    hue = (greenRatio - blueRatio) / delta + (greenRatio < blueRatio ? 6 : 0)
  } else if (maxChannel === greenRatio) {
    hue = (blueRatio - redRatio) / delta + 2
  } else {
    hue = (redRatio - greenRatio) / delta + 4
  }

  return { hue: hue / 6, saturation, lightness }
}

function hslToRgb(hue: number, saturation: number, lightness: number): number[] {
  const hueToChannel = (first: number, second: number, position: number): number => {
    let value = position
    if (value < 0) value += 1
    if (value > 1) value -= 1
    if (value < 1 / 6) return first + (second - first) * 6 * value
    if (value < 1 / 2) return second
    if (value < 2 / 3) return first + (second - first) * (2 / 3 - value) * 6
    return first
  }

  const second = lightness < 0.5
    ? lightness * (1 + saturation)
    : lightness + saturation - lightness * saturation
  const first = 2 * lightness - second
  return [
    Math.round(hueToChannel(first, second, hue + 1 / 3) * 255),
    Math.round(hueToChannel(first, second, hue) * 255),
    Math.round(hueToChannel(first, second, hue - 1 / 3) * 255),
  ]
}

function ensureWhiteTextContrast(channels: number[]): number[] {
  let adjusted = channels
  while (relativeLuminance(adjusted) > 0.18) {
    adjusted = adjusted.map((channel) => Math.round(channel * 0.85))
  }
  return adjusted
}

function relativeLuminance(channels: number[]): number {
  const [red, green, blue] = channels.map((channel) => {
    const ratio = channel / 255
    return ratio <= 0.04045 ? ratio / 12.92 : ((ratio + 0.055) / 1.055) ** 2.4
  })
  return red * 0.2126 + green * 0.7152 + blue * 0.0722
}

function rgbToHex(channels: number[]): string {
  return `#${channels.map((channel) => channel.toString(16).padStart(2, '0')).join('')}`
}
