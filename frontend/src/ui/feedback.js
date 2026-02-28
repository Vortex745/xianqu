const ensureBody = () => {
  if (typeof document === 'undefined') return false
  return !!document.body
}

const toMessageText = (input) => {
  if (typeof input === 'string') return input
  if (input && typeof input === 'object') return input.message || ''
  return ''
}

let toastHost = null
const ensureToastHost = () => {
  if (!ensureBody()) return null
  if (toastHost) return toastHost
  toastHost = document.createElement('div')
  toastHost.className = 'ui-toast-host'
  document.body.appendChild(toastHost)
  return toastHost
}

const showToast = (type, input, duration = 2200) => {
  const host = ensureToastHost()
  if (!host) return { close: () => {} }

  const message = toMessageText(input)
  const el = document.createElement('div')
  el.className = `ui-toast ui-toast--${type}`
  el.textContent = message
  host.appendChild(el)

  requestAnimationFrame(() => el.classList.add('is-show'))

  let timer = setTimeout(() => {
    close()
  }, duration)

  const close = () => {
    if (!el.parentNode) return
    clearTimeout(timer)
    el.classList.remove('is-show')
    setTimeout(() => {
      if (el.parentNode) el.parentNode.removeChild(el)
    }, 220)
  }

  return { close }
}

export const ElMessage = (input) => showToast('info', input)
ElMessage.success = (input) => showToast('success', input)
ElMessage.error = (input) => showToast('error', input)
ElMessage.warning = (input) => showToast('warning', input)
ElMessage.info = (input) => showToast('info', input)

const createConfirm = (message, title = '提示', options = {}) => {
  if (!ensureBody()) return Promise.reject(new Error('no-dom'))

  const confirmButtonText = options.confirmButtonText || '确定'
  const cancelButtonText = options.cancelButtonText || options.cancelButtonButtonText || '取消'
  const type = options.type || 'default'

  return new Promise((resolve, reject) => {
    const overlay = document.createElement('div')
    overlay.className = 'ui-msgbox-overlay'

    const panel = document.createElement('div')
    panel.className = `ui-msgbox ui-msgbox--${type}`

    const header = document.createElement('div')
    header.className = 'ui-msgbox__header'
    header.textContent = title

    const body = document.createElement('div')
    body.className = 'ui-msgbox__body'
    body.textContent = message

    const footer = document.createElement('div')
    footer.className = 'ui-msgbox__footer'

    const cancelBtn = document.createElement('button')
    cancelBtn.className = 'ui-btn ui-btn--ghost'
    cancelBtn.textContent = cancelButtonText

    const confirmBtn = document.createElement('button')
    confirmBtn.className = 'ui-btn ui-btn--solid'
    confirmBtn.textContent = confirmButtonText

    const close = () => {
      if (!overlay.parentNode) return
      overlay.classList.remove('is-show')
      setTimeout(() => {
        if (overlay.parentNode) overlay.parentNode.removeChild(overlay)
      }, 180)
      document.removeEventListener('keydown', onEsc, true)
    }

    const onEsc = (e) => {
      if (e.key === 'Escape') {
        close()
        reject(new Error('cancel'))
      }
    }

    cancelBtn.onclick = () => {
      close()
      reject(new Error('cancel'))
    }
    confirmBtn.onclick = () => {
      close()
      resolve('confirm')
    }

    if (options.closeOnClickModal !== false) {
      overlay.onclick = (e) => {
        if (e.target === overlay) {
          close()
          reject(new Error('cancel'))
        }
      }
    }

    footer.appendChild(cancelBtn)
    footer.appendChild(confirmBtn)
    panel.appendChild(header)
    panel.appendChild(body)
    panel.appendChild(footer)
    overlay.appendChild(panel)
    document.body.appendChild(overlay)
    requestAnimationFrame(() => overlay.classList.add('is-show'))
    document.addEventListener('keydown', onEsc, true)
  })
}

export const ElMessageBox = {
  confirm: createConfirm
}

export const ElLoading = {
  service(options = {}) {
    if (!ensureBody()) return { close: () => {} }

    const overlay = document.createElement('div')
    overlay.className = 'ui-loading-service'
    overlay.style.background = options.background || 'rgba(17, 18, 21, 0.42)'

    const box = document.createElement('div')
    box.className = 'ui-loading-service__box'
    const dot = document.createElement('div')
    dot.className = 'ui-loading-service__dot'
    const text = document.createElement('div')
    text.className = 'ui-loading-service__text'
    text.textContent = options.text || '加载中...'
    box.appendChild(dot)
    box.appendChild(text)
    overlay.appendChild(box)
    document.body.appendChild(overlay)

    return {
      close() {
        if (overlay.parentNode) overlay.parentNode.removeChild(overlay)
      }
    }
  }
}

export default {
  ElMessage,
  ElMessageBox,
  ElLoading
}
