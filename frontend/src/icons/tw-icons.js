import { defineComponent, h } from 'vue'

const iconNode = (tag, attrs) => ({ tag, attrs })

const createIcon = (name, nodes, options = {}) =>
  defineComponent({
    name,
    inheritAttrs: false,
    props: {
      size: {
        type: [Number, String],
        default: '1em'
      },
      strokeWidth: {
        type: [Number, String],
        default: 2
      }
    },
    setup(props, { attrs }) {
      return () => {
        const filled = !!options.filled
        const svgAttrs = {
          ...attrs,
          xmlns: 'http://www.w3.org/2000/svg',
          viewBox: options.viewBox || '0 0 24 24',
          width: props.size,
          height: props.size,
          fill: filled ? 'currentColor' : 'none',
          stroke: filled ? 'none' : 'currentColor',
          'stroke-width': filled ? undefined : props.strokeWidth,
          'stroke-linecap': 'round',
          'stroke-linejoin': 'round',
          class: ['tw-icon', attrs.class],
          'aria-hidden': 'true'
        }

        return h(
          'svg',
          svgAttrs,
          nodes.map((n, idx) => h(n.tag, { key: `${name}-${idx}`, ...n.attrs }))
        )
      }
    }
  })

export const Aim = createIcon('Aim', [
  iconNode('circle', { cx: '12', cy: '12', r: '7' }),
  iconNode('path', { d: 'M12 2v3M12 19v3M2 12h3M19 12h3' }),
  iconNode('circle', { cx: '12', cy: '12', r: '2.2' })
])

export const ArrowLeft = createIcon('ArrowLeft', [
  iconNode('path', { d: 'M19 12H5' }),
  iconNode('path', { d: 'M12 19l-7-7 7-7' })
])

export const ArrowRight = createIcon('ArrowRight', [
  iconNode('path', { d: 'M5 12h14' }),
  iconNode('path', { d: 'M12 5l7 7-7 7' })
])

export const Bell = createIcon('Bell', [
  iconNode('path', { d: 'M15 17h5l-1.4-1.4A2 2 0 0 1 18 14.2V11a6 6 0 1 0-12 0v3.2a2 2 0 0 1-.6 1.4L4 17h5' }),
  iconNode('path', { d: 'M9.5 20a2.5 2.5 0 0 0 5 0' })
])

export const Box = createIcon('Box', [
  iconNode('path', { d: 'M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16Z' }),
  iconNode('path', { d: 'm3.3 7 8.7 5 8.7-5' }),
  iconNode('path', { d: 'M12 22V12' })
], { strokeWidth: 2 })

export const Calendar = createIcon('Calendar', [
  iconNode('rect', { x: '3', y: '4', width: '18', height: '18', rx: '2' }),
  iconNode('path', { d: 'M16 2v4M8 2v4M3 10h18' })
])

export const Camera = createIcon('Camera', [
  iconNode('path', { d: 'M4 7h3l2-2h6l2 2h3a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2z' }),
  iconNode('circle', { cx: '12', cy: '13', r: '3.5' })
])

export const CaretBottom = createIcon('CaretBottom', [iconNode('path', { d: 'M6 9l6 6 6-6' })])
export const CaretTop = createIcon('CaretTop', [iconNode('path', { d: 'M18 15l-6-6-6 6' })])

export const ChatDotRound = createIcon('ChatDotRound', [
  iconNode('path', { d: 'M21 12a8.5 8.5 0 0 1-8.5 8.5H6l-3 2V12A8.5 8.5 0 1 1 21 12z' }),
  iconNode('circle', { cx: '12', cy: '12', r: '1.2' })
])

export const Check = createIcon('Check', [iconNode('path', { d: 'M20 6 9 17l-5-5' })])

export const CircleCheck = createIcon('CircleCheck', [
  iconNode('circle', { cx: '12', cy: '12', r: '9' }),
  iconNode('path', { d: 'm8.5 12.5 2.4 2.4 4.6-5.2' })
])

export const CircleCloseFilled = createIcon(
  'CircleCloseFilled',
  [
    iconNode('circle', { cx: '12', cy: '12', r: '10' }),
    iconNode('path', { d: 'M9 9l6 6M15 9l-6 6', stroke: '#fff', 'stroke-width': '2', fill: 'none' })
  ],
  { filled: true }
)

export const Close = createIcon('Close', [iconNode('path', { d: 'M18 6 6 18M6 6l12 12' })])
export const CloseBold = createIcon('CloseBold', [iconNode('path', { d: 'M18 6 6 18M6 6l12 12' })], { strokeWidth: 3 })

export const Delete = createIcon('Delete', [
  iconNode('path', { d: 'M3 6h18' }),
  iconNode('path', { d: 'M8 6V4h8v2' }),
  iconNode('path', { d: 'M6 6l1 14h10l1-14' }),
  iconNode('path', { d: 'M10 11v6M14 11v6' })
])

export const Edit = createIcon('Edit', [
  iconNode('path', { d: 'M12 20h9' }),
  iconNode('path', { d: 'M16.5 3.5a2.1 2.1 0 0 1 3 3L8 18l-4 1 1-4Z' })
])

export const Filter = createIcon('Filter', [iconNode('path', { d: 'M3 5h18l-7 8v5l-4 2v-7Z' })])

export const Goods = createIcon('Goods', [
  iconNode('path', { d: 'M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4Z' }),
  iconNode('path', { d: 'M3 6h18' }),
  iconNode('path', { d: 'M16 10a4 4 0 0 1-8 0' })
], { strokeWidth: 2 })

export const HomeFilled = createIcon('HomeFilled', [
  iconNode('path', { d: 'm3 9 9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z' }),
  iconNode('path', { d: 'M9 22V12h6v10' })
], { strokeWidth: 2 })

export const Iphone = createIcon('Iphone', [
  iconNode('rect', { x: '7', y: '2.5', width: '10', height: '19', rx: '2' }),
  iconNode('circle', { cx: '12', cy: '18.5', r: '0.7' })
])

export const List = createIcon('List', [
  iconNode('path', { d: 'M8 6h11M8 12h11M8 18h11' }),
  iconNode('path', { d: 'M3 6h.01M3 12h.01M3 18h.01' })
], { strokeWidth: 2.5 })

export const Loading = createIcon('Loading', [iconNode('path', { d: 'M21 12a9 9 0 1 1-4.2-7.6' })])

export const Location = createIcon('Location', [
  iconNode('path', { d: 'M12 21s7-6.4 7-11a7 7 0 1 0-14 0c0 4.6 7 11 7 11Z' }),
  iconNode('circle', { cx: '12', cy: '10', r: '2.2' })
])

export const LocationInformation = createIcon('LocationInformation', [
  iconNode('path', { d: 'M12 21s7-6.4 7-11a7 7 0 1 0-14 0c0 4.6 7 11 7 11Z' }),
  iconNode('path', { d: 'M12 7.8v4.4' }),
  iconNode('circle', { cx: '12', cy: '14.5', r: '0.7' })
])

export const Lock = createIcon('Lock', [
  iconNode('rect', { x: '4', y: '10', width: '16', height: '11', rx: '2' }),
  iconNode('path', { d: 'M8 10V7a4 4 0 0 1 8 0v3' })
])

export const Money = createIcon('Money', [
  iconNode('rect', { x: '2.5', y: '6', width: '19', height: '12', rx: '2' }),
  iconNode('circle', { cx: '12', cy: '12', r: '2.5' }),
  iconNode('path', { d: 'M6 12h.01M18 12h.01' })
])

export const Monitor = createIcon('Monitor', [
  iconNode('rect', { x: '2', y: '3', width: '20', height: '14', rx: '2' }),
  iconNode('path', { d: 'M12 17v4' }),
  iconNode('path', { d: 'M8 21h8' })
])

export const MoreFilled = createIcon(
  'MoreFilled',
  [
    iconNode('circle', { cx: '6', cy: '12', r: '2' }),
    iconNode('circle', { cx: '12', cy: '12', r: '2' }),
    iconNode('circle', { cx: '18', cy: '12', r: '2' })
  ],
  { filled: true }
)

export const Odometer = createIcon('Odometer', [
  iconNode('path', { d: 'M4 14a8 8 0 1 1 16 0' }),
  iconNode('path', { d: 'm12 12 4-3' }),
  iconNode('circle', { cx: '12', cy: '14', r: '1.3' })
])

export const Operation = createIcon('Operation', [
  iconNode('path', { d: 'M4 6h8M4 12h16M4 18h10' }),
  iconNode('circle', { cx: '16', cy: '6', r: '1.5' }),
  iconNode('circle', { cx: '10', cy: '18', r: '1.5' })
])

export const Picture = createIcon('Picture', [
  iconNode('rect', { x: '3', y: '4', width: '18', height: '16', rx: '2' }),
  iconNode('path', { d: 'm7 15 3-3 2 2 4-4 3 3' }),
  iconNode('circle', { cx: '8.5', cy: '9', r: '1.2' })
])

export const Plus = createIcon('Plus', [iconNode('path', { d: 'M12 5v14M5 12h14' })])

export const Promotion = createIcon('Promotion', [
  iconNode('path', { d: 'M3 11 21 3l-8 18-2-7-8-3z' }),
  iconNode('path', { d: 'm11 14 10-11' })
])

export const Refresh = createIcon('Refresh', [
  iconNode('path', { d: 'M21 12a9 9 0 0 1-15.5 6.3M3 12A9 9 0 0 1 18.5 5.7' }),
  iconNode('path', { d: 'M3 3v5h5M16 16h5v5' })
])

export const Right = createIcon('Right', [iconNode('path', { d: 'M9 18l6-6-6-6' })])

export const Scissor = createIcon('Scissor', [
  iconNode('circle', { cx: '6', cy: '6', r: '2' }),
  iconNode('circle', { cx: '6', cy: '18', r: '2' }),
  iconNode('path', { d: 'm20 4-8.5 8M20 20l-8.5-8' })
])

export const Search = createIcon('Search', [
  iconNode('circle', { cx: '11', cy: '11', r: '7' }),
  iconNode('path', { d: 'm20 20-3.5-3.5' })
])

export const Select = createIcon('Select', [iconNode('path', { d: 'M20 6 9 17l-5-5' })])

export const ShoppingCart = createIcon('ShoppingCart', [
  iconNode('path', { d: 'M3 4h2l2.2 10.2a2 2 0 0 0 2 1.6H18a2 2 0 0 0 2-1.6L21 7H7' }),
  iconNode('circle', { cx: '10', cy: '20', r: '1.4' }),
  iconNode('circle', { cx: '17', cy: '20', r: '1.4' })
])

export const Star = createIcon('Star', [iconNode('path', { d: 'm12 3 2.8 5.7 6.2.9-4.5 4.4 1 6.2L12 17l-5.5 3.2 1-6.2L3 9.6l6.2-.9Z' })])

export const StarFilled = createIcon('StarFilled', [iconNode('path', { d: 'm12 3 2.8 5.7 6.2.9-4.5 4.4 1 6.2L12 17l-5.5 3.2 1-6.2L3 9.6l6.2-.9Z' })], { filled: true })

export const Switch = createIcon('Switch', [
  iconNode('path', { d: 'M4 7h12l-3-3M20 17H8l3 3' }),
  iconNode('path', { d: 'M16 4v6M8 14v6' })
])

export const SwitchButton = createIcon('SwitchButton', [
  iconNode('path', { d: 'M12 2v10' }),
  iconNode('path', { d: 'M6.4 5.8A8 8 0 1 0 17.6 5.8' })
])

export const Top = createIcon('Top', [
  iconNode('path', { d: 'M12 19V8' }),
  iconNode('path', { d: 'm7 13 5-5 5 5' }),
  iconNode('path', { d: 'M5 4h14' })
])

export const User = createIcon('User', [
  iconNode('circle', { cx: '12', cy: '8', r: '5' }),
  iconNode('path', { d: 'M20 21a8 8 0 0 0-16 0' })
], { strokeWidth: 2 })

export const Van = createIcon('Van', [
  iconNode('path', { d: 'M2 7h10v8H2zM12 10h5l3 3v2h-8z' }),
  iconNode('circle', { cx: '6', cy: '18', r: '2' }),
  iconNode('circle', { cx: '16', cy: '18', r: '2' })
])

export const Wallet = createIcon('Wallet', [
  iconNode('path', { d: 'M19 7V4a1 1 0 0 0-1-1H5a2 2 0 0 0 0 4h15a1 1 0 0 1 1 1v10a1 1 0 0 1-1 1H5a2 2 0 0 1-2-2V5' }),
  iconNode('path', { d: 'M16 12h2' })
], { strokeWidth: 2 })

export const View = createIcon('View', [
  iconNode('path', { d: 'M2 12s3.5-7 10-7 10 7 10 7-3.5 7-10 7-10-7-10-7z' }),
  iconNode('circle', { cx: '12', cy: '12', r: '3' })
])

export const Hide = createIcon('Hide', [
  iconNode('path', { d: 'M3 3l18 18' }),
  iconNode('path', { d: 'M10.6 10.7a2.9 2.9 0 0 0 3.9 3.9' }),
  iconNode('path', { d: 'M9.9 5.3A11.8 11.8 0 0 1 12 5c6.5 0 10 7 10 7a15.5 15.5 0 0 1-4.4 4.9' }),
  iconNode('path', { d: 'M6.3 6.3A15.7 15.7 0 0 0 2 12s3.5 7 10 7c1.7 0 3.2-.4 4.6-1.1' })
])

export const Warning = createIcon('Warning', [
  iconNode('path', { d: 'm10.3 3.9-8.5 14.1c-.4.7-.4 1.5 0 2.2.4.7 1 1.1 1.8 1.1h16.9c.8 0 1.4-.4 1.8-1.1.4-.7.4-1.5 0-2.2L13.7 3.9c-.8-1.2-2.6-1.2-3.4 0z' }),
  iconNode('path', { d: 'M12 9v4' }),
  iconNode('path', { d: 'M12 17h.01' })
])

export const ZoomIn = createIcon('ZoomIn', [
  iconNode('circle', { cx: '11', cy: '11', r: '7' }),
  iconNode('path', { d: 'M11 8v6M8 11h6' }),
  iconNode('path', { d: 'm20 20-3.5-3.5' })
])

export const Message = createIcon('Message', [
  iconNode('rect', { x: '3', y: '5', width: '18', height: '14', rx: '2' }),
  iconNode('path', { d: 'M3 7l9 6 9-6' })
])

export const Cpu = createIcon('Cpu', [
  iconNode('rect', { x: '4', y: '4', width: '16', height: '16', rx: '2' }),
  iconNode('path', { d: 'M9 9h6v6H9zM9 1v3M15 1v3M9 20v3M15 20v3M20 9h3M20 14h3M1 9h3M1 14h3' })
])

export const DataLine = createIcon('DataLine', [
  iconNode('path', { d: 'M3 3v18h18' }),
  iconNode('path', { d: 'M19 9l-5 5-4-4-3 3' })
])

export const Coin = createIcon('Coin', [
  iconNode('circle', { cx: '12', cy: '12', r: '9' }),
  iconNode('path', { d: 'M12 7v10M9 10h6M9 14h6' })
])

export const Stopwatch = createIcon('Stopwatch', [
  iconNode('circle', { cx: '12', cy: '13', r: '8' }),
  iconNode('path', { d: 'M12 9v4l2 2' }),
  iconNode('path', { d: 'M10 2h4M12 2v3' })
])
