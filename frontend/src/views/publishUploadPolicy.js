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
