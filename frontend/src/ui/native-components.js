import {
  computed,
  defineComponent,
  h,
  inject,
  onBeforeUnmount,
  onMounted,
  provide,
  ref,
  Teleport,
  Transition
} from 'vue'
import { useRouter } from 'vue-router'
import { View, Hide } from '@element-plus/icons-vue'
import { ElMessage } from './feedback'

const px = (v) => (typeof v === 'number' ? `${v}px` : v)

const collectVNodes = (nodes, checker, out = []) => {
  ;(nodes || []).forEach((node) => {
    if (!node) return
    if (Array.isArray(node)) {
      collectVNodes(node, checker, out)
      return
    }
    if (checker(node)) out.push(node)
    if (Array.isArray(node.children)) {
      collectVNodes(node.children, checker, out)
    }
  })
  return out
}

const ElIcon = defineComponent({
  name: 'ElIcon',
  props: { size: [Number, String] },
  setup(props, { slots, attrs }) {
    return () =>
      h(
        'span',
        {
          class: ['el-icon', attrs.class],
          style: [{ fontSize: props.size ? px(props.size) : undefined }, attrs.style]
        },
        slots.default ? slots.default() : []
      )
  }
})

const ElButton = defineComponent({
  name: 'ElButton',
  props: {
    type: { type: String, default: 'default' },
    size: String,
    round: Boolean,
    circle: Boolean,
    plain: Boolean,
    link: Boolean,
    disabled: Boolean,
    loading: Boolean,
    color: String,
    icon: [Object, Function]
  },
  emits: ['click'],
  setup(props, { emit, slots, attrs }) {
    return () => {
      const classList = [
        'el-button',
        props.type ? `el-button--${props.type}` : '',
        props.size ? `el-button--${props.size}` : '',
        {
          'is-round': props.round,
          'is-circle': props.circle,
          'is-plain': props.plain,
          'is-link': props.link,
          'is-loading': props.loading
        },
        attrs.class
      ]

      return h(
        'button',
        {
          ...attrs,
          disabled: props.disabled || props.loading,
          class: classList,
          style: [
            attrs.style,
            props.color
              ? {
                  '--ui-btn-color': props.color
                }
              : {}
          ],
          onClick: (e) => {
            if (props.disabled || props.loading) return
            emit('click', e)
          }
        },
        [
          props.loading ? h('span', { class: 'ui-spinner' }) : null,
          props.icon ? h(props.icon, { class: 'el-button__icon' }) : null,
          h('span', { class: 'el-button__text' }, slots.default ? slots.default() : [])
        ]
      )
    }
  }
})

const ElAvatar = defineComponent({
  name: 'ElAvatar',
  props: { size: [Number, String], src: String },
  setup(props, { slots, attrs }) {
    const failed = ref(false)
    return () =>
      h(
        'span',
        {
          class: ['el-avatar', attrs.class],
          style: [
            {
              width: props.size ? px(props.size) : '40px',
              height: props.size ? px(props.size) : '40px'
            },
            attrs.style
          ]
        },
        [
          props.src && !failed.value
            ? h('img', {
                class: 'el-avatar__img',
                src: props.src,
                onError: () => {
                  failed.value = true
                }
              })
            : h('span', { class: 'el-avatar__placeholder' }, slots.default ? slots.default() : ' ')
        ]
      )
  }
})

const ElImage = defineComponent({
  name: 'ElImage',
  props: {
    src: String,
    fit: { type: String, default: 'cover' },
    previewSrcList: { type: Array, default: () => [] },
    hideOnClickModal: Boolean
  },
  setup(props, { slots, attrs }) {
    const failed = ref(false)
    const hasPreview = computed(() => Array.isArray(props.previewSrcList) && props.previewSrcList.length > 0)
    const openPreview = () => {
      if (!hasPreview.value) return
      const target = props.previewSrcList[0]
      if (!target) return
      window.open(target, '_blank')
    }
    return () =>
      h(
        'div',
        {
          class: ['el-image', attrs.class],
          style: attrs.style
        },
        failed.value && slots.error
          ? slots.error()
          : h('img', {
              class: 'el-image__inner',
              src: props.src,
              style: { objectFit: props.fit },
              onClick: hasPreview.value ? openPreview : undefined,
              onError: () => {
                failed.value = true
              }
            })
      )
  }
})

const ElEmpty = defineComponent({
  name: 'ElEmpty',
  props: {
    description: { type: String, default: '暂无数据' },
    imageSize: [Number, String]
  },
  setup(props, { attrs }) {
    return () =>
      h('div', { class: ['el-empty', attrs.class] }, [
        h('div', {
          class: 'el-empty__image',
          style: { width: props.imageSize ? px(props.imageSize) : '110px', height: props.imageSize ? px(props.imageSize) : '110px' }
        }),
        h('p', { class: 'el-empty__description' }, props.description)
      ])
  }
})

const ElBadge = defineComponent({
  name: 'ElBadge',
  props: { value: [Number, String], max: Number, hidden: Boolean },
  setup(props, { slots, attrs }) {
    const shown = computed(() => {
      if (props.hidden) return ''
      const num = Number(props.value)
      if (Number.isNaN(num)) return props.value ?? ''
      if (props.max && num > props.max) return `${props.max}+`
      return num
    })
    return () =>
      h('span', { class: ['el-badge', attrs.class] }, [
        slots.default ? slots.default() : [],
        !props.hidden && shown.value !== ''
          ? h('sup', { class: 'el-badge__content' }, String(shown.value))
          : null
      ])
  }
})

const callBeforeClose = (fn, done) => {
  if (!fn) return done()
  if (fn.length > 0) return fn(done)
  const res = fn()
  if (res === false) return
  done()
}

const ElDialog = defineComponent({
  name: 'ElDialog',
  props: {
    modelValue: Boolean,
    title: String,
    width: [Number, String],
    alignCenter: Boolean,
    showClose: { type: Boolean, default: true },
    destroyOnClose: Boolean,
    closeOnClickModal: { type: Boolean, default: true },
    beforeClose: Function,
    maxWidth: [Number, String]
  },
  emits: ['update:modelValue'],
  setup(props, { emit, slots, attrs }) {
    const close = () => emit('update:modelValue', false)
    const onClose = () => callBeforeClose(props.beforeClose, close)

    return () => {
      if (!props.modelValue && props.destroyOnClose) return null
      return h(Teleport, { to: 'body' }, [
        h(
          'div',
          {
            class: ['el-overlay', { 'is-show': props.modelValue }],
            style: { display: props.modelValue ? 'flex' : 'none', alignItems: props.alignCenter ? 'center' : 'flex-start' },
            onClick: (e) => {
              if (!props.closeOnClickModal) return
              if (e.target === e.currentTarget) onClose()
            }
          },
          [
            h('div', {
              class: ['el-dialog', attrs.class],
              style: {
                width: props.width ? px(props.width) : undefined,
                maxWidth: props.maxWidth ? px(props.maxWidth) : undefined
              }
            }, [
              h('div', { class: 'el-dialog__header' }, [
                h('span', { class: 'el-dialog__title' }, props.title || ''),
                props.showClose
                  ? h(
                      'button',
                      { class: 'el-dialog__headerbtn', type: 'button', onClick: onClose },
                      '×'
                    )
                  : null
              ]),
              h('div', { class: 'el-dialog__body' }, slots.default ? slots.default() : []),
              slots.footer ? h('div', { class: 'el-dialog__footer' }, slots.footer()) : null
            ])
          ]
        )
      ])
    }
  }
})

const ElForm = defineComponent({
  name: 'ElForm',
  props: {
    model: { type: Object, default: () => ({}) },
    rules: { type: Object, default: () => ({}) }
  },
  setup(props, { slots, expose, attrs }) {
    const validate = (cb) => {
      let valid = true
      Object.entries(props.rules || {}).forEach(([field, fieldRules]) => {
        ;(fieldRules || []).forEach((rule) => {
          if (rule.required && (props.model[field] === undefined || props.model[field] === null || String(props.model[field]).trim() === '')) {
            valid = false
            if (rule.message) ElMessage.warning(rule.message)
          }
        })
      })
      cb && cb(valid)
      return Promise.resolve(valid)
    }
    expose({ validate })
    return () =>
      h(
        'form',
        {
          class: ['el-form', attrs.class],
          onSubmit: (e) => e.preventDefault()
        },
        slots.default ? slots.default() : []
      )
  }
})

const ElFormItem = defineComponent({
  name: 'ElFormItem',
  setup(_, { slots, attrs }) {
    return () =>
      h('div', { class: ['el-form-item', attrs.class] }, [
        attrs.label ? h('label', { class: 'el-form-item__label' }, String(attrs.label)) : null,
        h('div', { class: 'el-form-item__content' }, slots.default ? slots.default() : [])
      ])
  }
})

const ElInput = defineComponent({
  name: 'ElInput',
  props: {
    modelValue: [String, Number],
    type: { type: String, default: 'text' },
    placeholder: String,
    prefixIcon: [Object, Function],
    showPassword: Boolean,
    maxlength: [Number, String],
    disabled: Boolean,
    rows: [Number, String],
    size: String,
    autocomplete: String
  },
  emits: ['update:modelValue', 'input', 'change'],
  setup(props, { emit, attrs }) {
    const focused = ref(false)
    const reveal = ref(false)

    const currentType = computed(() => {
      if (props.type !== 'password') return props.type
      if (!props.showPassword) return 'password'
      return reveal.value ? 'text' : 'password'
    })

    const onInput = (e) => {
      const val = e.target.value
      emit('update:modelValue', val)
      emit('input', val)
    }

    if (props.type === 'textarea') {
      return () =>
        h('textarea', {
          ...attrs,
          class: ['el-textarea__inner', attrs.class],
          rows: props.rows || 3,
          disabled: props.disabled,
          maxlength: props.maxlength,
          placeholder: props.placeholder,
          value: props.modelValue ?? '',
          onInput,
          onChange: (e) => emit('change', e.target.value)
        })
    }

    return () =>
      h('div', { class: ['el-input', attrs.class, props.size ? `el-input--${props.size}` : ''] }, [
        h('div', { class: ['el-input__wrapper', { 'is-focus': focused.value }] }, [
          props.prefixIcon ? h(props.prefixIcon, { class: 'el-input__prefix-icon' }) : null,
          h('input', {
            class: 'el-input__inner',
            type: currentType.value,
            disabled: props.disabled,
            maxlength: props.maxlength,
            placeholder: props.placeholder,
            autocomplete: props.autocomplete || 'off',
            value: props.modelValue ?? '',
            onFocus: () => {
              focused.value = true
            },
            onBlur: () => {
              focused.value = false
            },
            onInput,
            onChange: (e) => emit('change', e.target.value)
          }),
          props.showPassword && props.type === 'password'
            ? h(
                'button',
                {
                  class: 'el-input__pwd-btn',
                  type: 'button',
                  'aria-label': reveal.value ? '隐藏密码' : '显示密码',
                  title: reveal.value ? '隐藏密码' : '显示密码',
                  onClick: () => {
                    reveal.value = !reveal.value
                  }
                },
                [h(reveal.value ? Hide : View, { class: 'el-input__pwd-icon' })]
              )
            : null
        ])
      ])
  }
})

const ElUpload = defineComponent({
  name: 'ElUpload',
  props: {
    action: { type: String, required: true },
    name: { type: String, default: 'file' },
    headers: { type: Object, default: () => ({}) },
    showFileList: Boolean,
    beforeUpload: Function,
    onSuccess: Function,
    onError: Function
  },
  setup(props, { slots, attrs }) {
    const inputRef = ref(null)

    const doUpload = async (file) => {
      if (!file) return
      if (props.beforeUpload) {
        const allowed = await props.beforeUpload(file)
        if (allowed === false) return
      }

      const body = new FormData()
      body.append(props.name, file)

      try {
        const res = await fetch(props.action, {
          method: 'POST',
          headers: props.headers || {},
          body
        })
        const data = await res.json()
        if (!res.ok) throw new Error(data?.error || '上传失败')
        props.onSuccess && props.onSuccess(data, file)
      } catch (err) {
        props.onError && props.onError(err, file)
      }
    }

    return () =>
      h(
        'div',
        {
          class: ['el-upload', attrs.class],
          onClick: () => {
            if (inputRef.value) inputRef.value.click()
          }
        },
        [
          h('input', {
            ref: inputRef,
            class: 'el-upload__input',
            type: 'file',
            onClick: (e) => e.stopPropagation(),
            onChange: (e) => {
              const file = e.target.files && e.target.files[0]
              doUpload(file)
              e.target.value = ''
            }
          }),
          slots.default ? slots.default() : []
        ]
      )
  }
})

const ElSwitch = defineComponent({
  name: 'ElSwitch',
  props: { modelValue: Boolean, disabled: Boolean },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit, attrs }) {
    return () =>
      h(
        'button',
        {
          type: 'button',
          class: ['el-switch', { 'is-checked': props.modelValue, 'is-disabled': props.disabled }, attrs.class],
          onClick: () => {
            if (props.disabled) return
            const next = !props.modelValue
            emit('update:modelValue', next)
            emit('change', next)
          }
        },
        [h('span', { class: 'el-switch__core' })]
      )
  }
})

const ElBacktop = defineComponent({
  name: 'ElBacktop',
  props: { right: [Number, String], bottom: [Number, String] },
  setup(props, { slots, attrs }) {
    const shown = ref(false)
    const onScroll = () => {
      shown.value = window.scrollY > 220
    }
    onMounted(() => {
      onScroll()
      window.addEventListener('scroll', onScroll, { passive: true })
    })
    onBeforeUnmount(() => window.removeEventListener('scroll', onScroll))
    return () =>
      shown.value
        ? h(
            'button',
            {
              type: 'button',
              class: ['el-backtop', attrs.class],
              style: { right: px(props.right || 40), bottom: px(props.bottom || 40) },
              onClick: () => window.scrollTo({ top: 0, behavior: 'smooth' })
            },
            slots.default ? slots.default() : '↑'
          )
        : null
  }
})

const ElDrawer = defineComponent({
  name: 'ElDrawer',
  props: {
    modelValue: Boolean,
    title: String,
    direction: { type: String, default: 'rtl' },
    size: { type: [Number, String], default: '320px' }
  },
  emits: ['update:modelValue'],
  setup(props, { emit, slots, attrs }) {
    const close = () => emit('update:modelValue', false)
    return () =>
      h(Teleport, { to: 'body' }, [
        h(
          'div',
          {
            class: ['el-drawer__overlay', { 'is-show': props.modelValue }],
            style: { display: props.modelValue ? 'block' : 'none' },
            onClick: (e) => {
              if (e.target === e.currentTarget) close()
            }
          },
          [
            h(
              'aside',
              {
                class: ['el-drawer', `is-${props.direction}`, attrs.class],
                style: { width: px(props.size) }
              },
              [
                h('header', { class: 'el-drawer__header' }, [
                  h('span', props.title || ''),
                  h('button', { type: 'button', class: 'el-drawer__close', onClick: close }, '×')
                ]),
                h('div', { class: 'el-drawer__body' }, slots.default ? slots.default() : [])
              ]
            )
          ]
        )
      ])
  }
})

const dropdownContextKey = Symbol('dropdown')
const ElDropdown = defineComponent({
  name: 'ElDropdown',
  props: {
    trigger: { type: String, default: 'hover' },
    showTimeout: { type: [Number, String], default: 90 },
    hideTimeout: { type: [Number, String], default: 120 }
  },
  setup(props, { slots, attrs }) {
    const open = ref(false)
    const wrapRef = ref(null)
    let showTimer = null
    let hideTimer = null

    const clearTimers = () => {
      if (showTimer) {
        clearTimeout(showTimer)
        showTimer = null
      }
      if (hideTimer) {
        clearTimeout(hideTimer)
        hideTimer = null
      }
    }

    const close = () => {
      clearTimers()
      open.value = false
    }
    const show = () => {
      clearTimers()
      showTimer = setTimeout(() => {
        open.value = true
      }, Number(props.showTimeout) || 0)
    }
    const hide = () => {
      clearTimers()
      hideTimer = setTimeout(() => {
        open.value = false
      }, Number(props.hideTimeout) || 0)
    }

    const onDocClick = (e) => {
      if (!wrapRef.value) return
      if (!wrapRef.value.contains(e.target)) close()
    }
    onMounted(() => document.addEventListener('click', onDocClick, true))
    onBeforeUnmount(() => {
      document.removeEventListener('click', onDocClick, true)
      clearTimers()
    })
    provide(dropdownContextKey, { close })

    return () =>
      h('div', {
        ref: wrapRef,
        class: ['el-dropdown', attrs.class],
        onMouseenter: () => {
          if (props.trigger === 'hover') show()
        },
        onMouseleave: () => {
          if (props.trigger === 'hover') hide()
        }
      }, [
        h(
          'div',
          {
            class: 'el-dropdown__trigger',
            onClick: (e) => {
              if (props.trigger !== 'click') return
              e.stopPropagation()
              open.value = !open.value
            }
          },
          slots.default ? slots.default() : []
        ),
        open.value ? h('div', { class: 'el-dropdown__menu-wrap' }, slots.dropdown ? slots.dropdown() : []) : null
      ])
  }
})

const ElDropdownMenu = defineComponent({
  name: 'ElDropdownMenu',
  setup(_, { slots, attrs }) {
    return () => h('ul', { class: ['el-dropdown-menu', attrs.class] }, slots.default ? slots.default() : [])
  }
})

const ElDropdownItem = defineComponent({
  name: 'ElDropdownItem',
  props: { divided: Boolean },
  emits: ['click'],
  setup(props, { slots, emit, attrs }) {
    const ctx = inject(dropdownContextKey, { close: () => {} })
    return () =>
      h(
        'li',
        {
          class: ['el-dropdown-menu__item', { 'is-divided': props.divided }, attrs.class],
          onClick: (e) => {
            emit('click', e)
            ctx.close()
          }
        },
        slots.default ? slots.default() : []
      )
  }
})

const ElTooltip = defineComponent({
  name: 'ElTooltip',
  props: { content: String },
  setup(props, { slots, attrs }) {
    return () =>
      h(
        'span',
        {
          class: ['el-tooltip', attrs.class],
          title: props.content || ''
        },
        slots.default ? slots.default() : []
      )
  }
})

const ElTag = defineComponent({
  name: 'ElTag',
  props: { type: String, size: String, effect: String },
  setup(props, { slots, attrs }) {
    return () =>
      h('span', { class: ['el-tag', props.type ? `el-tag--${props.type}` : '', attrs.class] }, slots.default ? slots.default() : [])
  }
})

const ElTableColumn = defineComponent({
  name: 'ElTableColumn',
  __isUiColumn: true,
  setup() {
    return () => null
  }
})

const ElTable = defineComponent({
  name: 'ElTable',
  props: {
    data: { type: Array, default: () => [] },
    stripe: Boolean,
    border: Boolean,
    headerCellStyle: { type: Object, default: () => ({}) }
  },
  setup(props, { slots, attrs }) {
    const columns = computed(() => {
      const nodes = slots.default ? slots.default() : []
      return collectVNodes(nodes, (node) => node?.type?.name === 'ElTableColumn')
    })

    const getCell = (row, colVNode, index) => {
      const prop = colVNode.props?.prop
      const slotDefault = colVNode.children && typeof colVNode.children.default === 'function' ? colVNode.children.default : null
      if (slotDefault) return slotDefault({ row, $index: index, column: colVNode.props || {} })
      if (!prop) return ''
      return row[prop]
    }

    return () =>
      h('div', { class: ['el-table', { 'el-table--border': props.border, 'el-table--stripe': props.stripe }, attrs.class] }, [
        h('table', { class: 'el-table__inner' }, [
          h(
            'thead',
            {},
            h(
              'tr',
              {},
              columns.value.map((col) =>
                h(
                  'th',
                  {
                    style: {
                      width: col.props?.width ? px(col.props.width) : undefined,
                      minWidth: col.props?.minWidth ? px(col.props.minWidth) : undefined,
                      textAlign: col.props?.align || 'left',
                      ...props.headerCellStyle
                    }
                  },
                  col.props?.label || ''
                )
              )
            )
          ),
          h(
            'tbody',
            {},
            (props.data || []).map((row, rowIndex) =>
              h(
                'tr',
                { class: rowIndex % 2 === 1 && props.stripe ? 'is-stripe' : '' },
                columns.value.map((col) =>
                  h(
                    'td',
                    {
                      style: {
                        textAlign: col.props?.align || 'left'
                      }
                    },
                    getCell(row, col, rowIndex)
                  )
                )
              )
            )
          )
        ])
      ])
  }
})

const ElPagination = defineComponent({
  name: 'ElPagination',
  props: {
    total: { type: Number, default: 0 },
    pageSize: { type: Number, default: 10 },
    currentPage: { type: Number, default: 1 }
  },
  emits: ['update:current-page', 'current-change'],
  setup(props, { emit, attrs }) {
    const pageCount = computed(() => Math.max(1, Math.ceil((props.total || 0) / (props.pageSize || 10))))
    const jump = (p) => {
      const page = Math.min(pageCount.value, Math.max(1, p))
      emit('update:current-page', page)
      emit('current-change', page)
    }
    return () =>
      h('div', { class: ['el-pagination', attrs.class] }, [
        h(
          'button',
          { class: 'el-pagination__btn', disabled: props.currentPage <= 1, onClick: () => jump(props.currentPage - 1) },
          '上一页'
        ),
        h('span', { class: 'el-pagination__state' }, `${props.currentPage} / ${pageCount.value}`),
        h(
          'button',
          { class: 'el-pagination__btn', disabled: props.currentPage >= pageCount.value, onClick: () => jump(props.currentPage + 1) },
          '下一页'
        )
      ])
  }
})

const ElRow = defineComponent({
  name: 'ElRow',
  props: { gutter: { type: Number, default: 0 } },
  setup(props, { slots, attrs }) {
    return () =>
      h(
        'div',
        {
          class: ['el-row', attrs.class],
          style: {
            gap: `${props.gutter || 0}px`
          }
        },
        slots.default ? slots.default() : []
      )
  }
})

const ElCol = defineComponent({
  name: 'ElCol',
  props: {
    span: Number,
    xs: Number,
    sm: Number,
    md: Number,
    lg: Number
  },
  setup(props, { slots, attrs }) {
    const basis = computed(() => {
      const value = props.span || props.lg || props.md || props.sm || props.xs || 24
      return `${(value / 24) * 100}%`
    })
    return () =>
      h(
        'div',
        {
          class: ['el-col', attrs.class],
          style: {
            flex: `0 0 ${basis.value}`,
            maxWidth: basis.value
          }
        },
        slots.default ? slots.default() : []
      )
  }
})

const menuContextKey = Symbol('menu')
const ElMenu = defineComponent({
  name: 'ElMenu',
  props: { defaultActive: String, router: Boolean },
  setup(props, { slots, attrs }) {
    const active = ref(props.defaultActive || '')
    const router = useRouter()
    const select = (index) => {
      active.value = index
      if (props.router && index) router.push(index)
    }
    provide(menuContextKey, { active, select })
    return () =>
      h('div', { class: ['el-menu', attrs.class] }, slots.default ? slots.default() : [])
  }
})

const ElMenuItem = defineComponent({
  name: 'ElMenuItem',
  props: { index: String },
  setup(props, { slots, attrs }) {
    const menu = inject(menuContextKey, null)
    return () =>
      h(
        'div',
        {
          class: ['el-menu-item', { 'is-active': menu?.active.value === props.index }, attrs.class],
          onClick: () => menu?.select(props.index)
        },
        slots.default ? slots.default() : []
      )
  }
})

const ElBreadcrumb = defineComponent({
  name: 'ElBreadcrumb',
  setup(_, { slots, attrs }) {
    return () => h('nav', { class: ['el-breadcrumb', attrs.class] }, slots.default ? slots.default() : [])
  }
})

const ElBreadcrumbItem = defineComponent({
  name: 'ElBreadcrumbItem',
  props: { to: [String, Object] },
  setup(props, { slots, attrs }) {
    const router = useRouter()
    return () =>
      h(
        'span',
        {
          class: ['el-breadcrumb__item', attrs.class],
          onClick: () => {
            if (!props.to) return
            const to = typeof props.to === 'string' ? props.to : props.to.path
            if (to) router.push(to)
          }
        },
        slots.default ? slots.default() : []
      )
  }
})

const ElInputNumber = defineComponent({
  name: 'ElInputNumber',
  props: { modelValue: Number, min: Number, max: Number },
  emits: ['update:modelValue', 'change'],
  setup(props, { emit, attrs }) {
    const clamp = (val) => {
      let v = Number(val || 0)
      if (Number.isNaN(v)) v = props.min || 0
      if (props.min !== undefined) v = Math.max(props.min, v)
      if (props.max !== undefined) v = Math.min(props.max, v)
      return v
    }
    const set = (val) => {
      const v = clamp(val)
      emit('update:modelValue', v)
      emit('change', v)
    }
    return () =>
      h('div', { class: ['el-input-number', attrs.class] }, [
        h('button', { type: 'button', class: 'el-input-number__decrease', onClick: () => set((props.modelValue || 0) - 1) }, '-'),
        h('input', {
          class: 'el-input-number__inner',
          type: 'number',
          value: props.modelValue ?? '',
          onInput: (e) => set(e.target.value)
        }),
        h('button', { type: 'button', class: 'el-input-number__increase', onClick: () => set((props.modelValue || 0) + 1) }, '+')
      ])
  }
})

const ElOption = defineComponent({
  name: 'ElOption',
  __isUiOption: true,
  props: { label: [String, Number], value: [String, Number] },
  setup() {
    return () => null
  }
})

const ElSelect = defineComponent({
  name: 'ElSelect',
  props: { modelValue: [String, Number], placeholder: String },
  emits: ['update:modelValue', 'change'],
  setup(props, { slots, emit, attrs }) {
    const options = computed(() => {
      const nodes = slots.default ? slots.default() : []
      const vnodes = collectVNodes(nodes, (node) => node?.type?.name === 'ElOption')
      return vnodes.map((n) => ({
        label: n.props?.label ?? n.props?.value,
        value: n.props?.value
      }))
    })
    return () =>
      h('div', { class: ['el-select', attrs.class] }, [
        h(
          'select',
          {
            class: 'el-select__inner',
            value: props.modelValue ?? '',
            onChange: (e) => {
              emit('update:modelValue', e.target.value)
              emit('change', e.target.value)
            }
          },
          [
            h('option', { value: '' }, props.placeholder || '请选择'),
            ...options.value.map((opt) => h('option', { value: opt.value }, String(opt.label)))
          ]
        )
      ])
  }
})

const radioContextKey = Symbol('radio-group')
const ElRadioGroup = defineComponent({
  name: 'ElRadioGroup',
  props: { modelValue: [String, Number] },
  emits: ['update:modelValue'],
  setup(props, { slots, emit, attrs }) {
    provide(radioContextKey, {
      value: computed(() => props.modelValue),
      set: (v) => emit('update:modelValue', v)
    })
    return () => h('div', { class: ['el-radio-group', attrs.class] }, slots.default ? slots.default() : [])
  }
})

const ElRadioButton = defineComponent({
  name: 'ElRadioButton',
  props: { label: [String, Number], disabled: Boolean },
  setup(props, { slots, attrs }) {
    const group = inject(radioContextKey)
    return () =>
      h(
        'button',
        {
          type: 'button',
          class: ['el-radio-button', { 'is-active': group?.value.value === props.label, 'is-disabled': props.disabled }, attrs.class],
          disabled: props.disabled,
          onClick: () => {
            if (!props.disabled) group?.set(props.label)
          }
        },
        slots.default ? slots.default() : String(props.label)
      )
  }
})

const ElTabPane = defineComponent({
  name: 'ElTabPane',
  __isUiTabPane: true,
  props: { label: String, name: [String, Number] },
  setup(_, { slots }) {
    return () => (slots.default ? slots.default() : [])
  }
})

const ElTabs = defineComponent({
  name: 'ElTabs',
  props: { modelValue: [String, Number] },
  emits: ['update:modelValue'],
  setup(props, { slots, emit, attrs }) {
    const panes = computed(() => {
      const nodes = slots.default ? slots.default() : []
      return collectVNodes(nodes, (node) => node?.type?.name === 'ElTabPane')
    })
    const current = computed(() => props.modelValue ?? panes.value[0]?.props?.name)
    return () =>
      h('div', { class: ['el-tabs', attrs.class] }, [
        h(
          'div',
          { class: 'el-tabs__header' },
          panes.value.map((pane) => {
            const labelSlot = pane.children?.label
            const labelNode = labelSlot ? labelSlot() : pane.props?.label || ''
            return h(
              'button',
              {
                class: ['el-tabs__item', { 'is-active': current.value === pane.props?.name }],
                type: 'button',
                onClick: () => emit('update:modelValue', pane.props?.name)
              },
              labelNode
            )
          })
        ),
        h(
          'div',
          { class: 'el-tabs__content' },
          panes.value
            .filter((pane) => pane.props?.name === current.value)
            .map((pane) => (pane.children?.default ? pane.children.default() : []))
        )
      ])
  }
})

const ElCollapseTransition = defineComponent({
  name: 'ElCollapseTransition',
  setup(_, { slots }) {
    return () =>
      h(
        Transition,
        { name: 'ui-collapse' },
        {
          default: () => (slots.default ? slots.default() : [])
        }
      )
  }
})

const loadingDirective = {
  mounted(el, binding) {
    if (getComputedStyle(el).position === 'static') {
      el.style.position = 'relative'
    }
    const overlay = document.createElement('div')
    overlay.className = 'ui-loading-mask'
    const bg = el.getAttribute('element-loading-background')
    if (bg) overlay.style.background = bg
    overlay.innerHTML = `<div class="ui-loading-spin"></div><div class="ui-loading-text"></div>`
    overlay.querySelector('.ui-loading-text').textContent = el.getAttribute('element-loading-text') || '加载中...'
    el.__uiLoadingMask = overlay
    if (binding.value) el.appendChild(overlay)
  },
  updated(el, binding) {
    const overlay = el.__uiLoadingMask
    if (!overlay) return
    if (binding.value && !overlay.parentNode) {
      el.appendChild(overlay)
    } else if (!binding.value && overlay.parentNode) {
      overlay.parentNode.removeChild(overlay)
    }
  },
  unmounted(el) {
    const overlay = el.__uiLoadingMask
    if (overlay && overlay.parentNode) overlay.parentNode.removeChild(overlay)
    delete el.__uiLoadingMask
  }
}

const components = {
  ElIcon,
  ElButton,
  ElAvatar,
  ElImage,
  ElEmpty,
  ElBadge,
  ElDialog,
  ElForm,
  ElFormItem,
  ElInput,
  ElUpload,
  ElSwitch,
  ElBacktop,
  ElDrawer,
  ElDropdown,
  ElDropdownMenu,
  ElDropdownItem,
  ElTooltip,
  ElTag,
  ElTable,
  ElTableColumn,
  ElPagination,
  ElRow,
  ElCol,
  ElMenu,
  ElMenuItem,
  ElBreadcrumb,
  ElBreadcrumbItem,
  ElInputNumber,
  ElSelect,
  ElOption,
  ElRadioGroup,
  ElRadioButton,
  ElTabs,
  ElTabPane,
  ElCollapseTransition
}

export default {
  install(app) {
    Object.entries(components).forEach(([name, component]) => {
      app.component(name, component)
    })
    app.directive('loading', loadingDirective)
  }
}
