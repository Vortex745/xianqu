import { prepareImageForVercelUpload } from '@/utils/imageUpload'

export function validatePublishImageUpload(file) {
  if (!file) {
    return {
      allowed: false,
      message: '请选择要上传的图片'
    }
  }

  return {
    allowed: true,
    message: ''
  }
}

export async function preparePublishImageUpload(file) {
  const result = await prepareImageForVercelUpload(file)
  return {
    allowed: result.ok,
    message: result.message || '',
    file: result.file
  }
}
