export const VERCEL_UPLOAD_LIMIT_BYTES = 4 * 1024 * 1024
export const UPLOAD_LIMIT_LABEL = '4MB'

const TARGET_IMAGE_BYTES = 3.6 * 1024 * 1024
const MAX_IMAGE_EDGE = 1600
const MIN_JPEG_QUALITY = 0.58

const canUseCanvas = () => (
  typeof document !== 'undefined' &&
  typeof FileReader !== 'undefined' &&
  typeof Image !== 'undefined'
)

const fileToDataUrl = (file) => new Promise((resolve, reject) => {
  const reader = new FileReader()
  reader.onload = () => resolve(String(reader.result || ''))
  reader.onerror = () => reject(reader.error || new Error('读取图片失败'))
  reader.readAsDataURL(file)
})

const loadImage = (src) => new Promise((resolve, reject) => {
  const img = new Image()
  img.onload = () => resolve(img)
  img.onerror = () => reject(new Error('图片解析失败'))
  img.src = src
})

const canvasToBlob = (canvas, type, quality) => new Promise((resolve, reject) => {
  canvas.toBlob((blob) => {
    if (blob) resolve(blob)
    else reject(new Error('图片压缩失败'))
  }, type, quality)
})

const toUploadFile = (blob, originalFile) => {
  const sourceName = originalFile.name || 'upload.jpg'
  const safeName = sourceName.replace(/\.[^.]+$/, '') || 'upload'
  const fileName = blob.type === 'image/png' ? `${safeName}.png` : `${safeName}.jpg`
  return new File([blob], fileName, {
    type: blob.type || 'image/jpeg',
    lastModified: Date.now()
  })
}

export async function prepareImageForVercelUpload(file) {
  if (!file) {
    return {
      ok: false,
      message: '请选择要上传的图片'
    }
  }

  if (!String(file.type || '').startsWith('image/')) {
    return {
      ok: false,
      message: '请上传图片文件'
    }
  }

  if (file.size <= TARGET_IMAGE_BYTES || !canUseCanvas()) {
    if (file.size > VERCEL_UPLOAD_LIMIT_BYTES) {
      return {
        ok: false,
        message: `图片不能超过 ${UPLOAD_LIMIT_LABEL}`
      }
    }
    return {
      ok: true,
      file
    }
  }

  try {
    const dataUrl = await fileToDataUrl(file)
    const img = await loadImage(dataUrl)
    const scale = Math.min(1, MAX_IMAGE_EDGE / Math.max(img.width, img.height))
    const width = Math.max(1, Math.round(img.width * scale))
    const height = Math.max(1, Math.round(img.height * scale))

    const canvas = document.createElement('canvas')
    canvas.width = width
    canvas.height = height
    const ctx = canvas.getContext('2d')
    if (!ctx) throw new Error('浏览器不支持图片压缩')
    ctx.drawImage(img, 0, 0, width, height)

    const outputType = file.type === 'image/png' ? 'image/png' : 'image/jpeg'
    let quality = 0.82
    let blob = await canvasToBlob(canvas, outputType, quality)

    while (blob.size > TARGET_IMAGE_BYTES && outputType === 'image/jpeg' && quality > MIN_JPEG_QUALITY) {
      quality = Math.max(MIN_JPEG_QUALITY, quality - 0.08)
      blob = await canvasToBlob(canvas, outputType, quality)
    }

    if (blob.size > VERCEL_UPLOAD_LIMIT_BYTES) {
      return {
        ok: false,
        message: `图片压缩后仍超过 ${UPLOAD_LIMIT_LABEL}，请换一张更小的图片`
      }
    }

    return {
      ok: true,
      file: toUploadFile(blob, file)
    }
  } catch (error) {
    if (file.size > VERCEL_UPLOAD_LIMIT_BYTES) {
      return {
        ok: false,
        message: `图片不能超过 ${UPLOAD_LIMIT_LABEL}`
      }
    }
    return {
      ok: true,
      file
    }
  }
}
