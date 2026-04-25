import gsap from 'gsap'
import { ScrollTrigger } from 'gsap/ScrollTrigger'

// Register plugins once at app level
gsap.registerPlugin(ScrollTrigger)

export function shouldReduceMotion() {
  return typeof window !== 'undefined' &&
    typeof window.matchMedia === 'function' &&
    window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

// ===== Global Animation Configuration =====
export const animConfig = {
  // Easing presets
  ease: {
    smooth: 'power3.out',
    bouncy: 'back.out(1.12)',
    snappy: 'expo.out',
    gentle: 'power2.out',
    elastic: 'elastic.out(1, 0.35)',
    expo: 'expo.out'
  },

  // Duration presets (seconds)
  duration: {
    fast: 0.18,
    normal: 0.26,
    slow: 0.4,
    dramatic: 0.6
  },

  // Stagger presets
  stagger: {
    tight: 0.03,
    normal: 0.05,
    loose: 0.1,
    cascade: 0.08
  },

  // ScrollTrigger defaults
  scroll: {
    start: 'top 85%',
    end: 'bottom 20%',
    once: true
  }
}

// ===== Reusable Animation Helpers =====

/**
 * Batch animate elements with ScrollTrigger
 * @param {string} selector - CSS selector for target elements
 * @param {object} fromVars - GSAP from properties
 * @param {object} options - ScrollTrigger options
 */
export function scrollBatchAnimate(selector, fromVars = {}, options = {}) {
  if (shouldReduceMotion()) return null

  const elements = document.querySelectorAll(selector)
  if (!elements.length) return null

  const defaults = {
    y: 30,
    opacity: 0,
    duration: animConfig.duration.normal,
    ease: animConfig.ease.smooth,
    stagger: animConfig.stagger.normal,
    scrollTrigger: {
      trigger: elements[0].parentElement || elements[0],
      start: animConfig.scroll.start,
      once: animConfig.scroll.once
    }
  }

  // Merge custom scrollTrigger options
  if (options.scrollTrigger) {
    defaults.scrollTrigger = { ...defaults.scrollTrigger, ...options.scrollTrigger }
    delete options.scrollTrigger
  }

  return gsap.from(elements, { ...defaults, ...fromVars, ...options })
}

/**
 * Create a scoped context for component animations
 * @param {HTMLElement} scope - Root element for selector scoping
 * @param {Function} callback - Animation setup function
 * @returns {object} GSAP context
 */
export function createAnimContext(scope, callback) {
  if (!scope) return null
  return gsap.context(callback, scope)
}

/**
 * Page entrance animation sequence
 * @param {HTMLElement} container - Page container element
 * @param {object} options - Animation options
 */
export function pageEntrance(container, options = {}) {
  if (!container || shouldReduceMotion()) return null

  const {
    headerSelector = '.nav-header, .navbar, .page-header, .fixed-header',
    contentSelector = '.main-content, .hero-section, .result-container',
    cardSelector = '.product-card, .order-card, .cart-item-card, .grid-item',
    delay = 0
  } = options

  const ctx = gsap.context(() => {
    const tl = gsap.timeline({ delay })

    // Header slides down
    const header = container.querySelector(headerSelector)
    if (header) {
      tl.from(header, {
        y: -18,
        opacity: 0,
        duration: animConfig.duration.fast,
        ease: animConfig.ease.snappy
      }, 0)
    }

    // Content fades up
    const content = container.querySelector(contentSelector)
    if (content) {
      tl.from(content, {
        y: 18,
        opacity: 0,
        duration: animConfig.duration.normal,
        ease: animConfig.ease.smooth
      }, 0.08)
    }

    // Cards stagger in
    const cards = container.querySelectorAll(cardSelector)
    if (cards.length) {
      tl.from(cards, {
        y: 24,
        opacity: 0,
        scale: 0.985,
        duration: animConfig.duration.normal,
        stagger: animConfig.stagger.tight,
        ease: animConfig.ease.smooth
      }, 0.14)
    }
  }, container)

  return ctx
}

/**
 * Interactive feedback animation
 * @param {HTMLElement} element - Target element
 * @param {string} type - Feedback type: 'success' | 'error' | 'pulse' | 'bounce'
 */
export function feedbackAnim(element, type = 'pulse') {
  if (!element || shouldReduceMotion()) return

  const presets = {
    success: {
      scale: 1.03,
      duration: 0.16,
      ease: animConfig.ease.bouncy,
      yoyo: true,
      repeat: 1
    },
    error: {
      x: [-8, 8, -6, 6, -3, 3, 0],
      duration: 0.4,
      ease: animConfig.ease.smooth
    },
    pulse: {
      scale: 1.03,
      duration: 0.12,
      ease: animConfig.ease.gentle,
      yoyo: true,
      repeat: 1
    },
    bounce: {
      y: -6,
      duration: 0.2,
      ease: animConfig.ease.bouncy,
      yoyo: true,
      repeat: 1
    }
  }

  gsap.to(element, presets[type] || presets.pulse)
}

/**
 * Smooth number counter animation
 * @param {object} target - Vue ref or object with value property
 * @param {number} endValue - Target number
 * @param {number} duration - Animation duration
 */
export function countUp(target, endValue, duration = 1) {
  if (shouldReduceMotion()) {
    if (target && typeof target.value !== 'undefined') {
      target.value = endValue
    }
    return
  }

  const obj = { value: 0 }
  gsap.to(obj, {
    value: endValue,
    duration,
    ease: animConfig.ease.smooth,
    onUpdate: () => {
      if (target && typeof target.value !== 'undefined') {
        target.value = Math.round(obj.value)
      }
    }
  })
}

/**
 * Refresh ScrollTrigger after dynamic content loads
 * Use after API calls or DOM updates
 */
export function refreshScrollTriggers() {
  requestAnimationFrame(() => {
    ScrollTrigger.refresh()
  })
}

/**
 * Kill all animations in a context safely
 * @param {object} ctx - GSAP context
 */
export function cleanupContext(ctx) {
  if (ctx && typeof ctx.revert === 'function') {
    ctx.revert()
  }
}

export default {
  animConfig,
  shouldReduceMotion,
  scrollBatchAnimate,
  createAnimContext,
  pageEntrance,
  feedbackAnim,
  countUp,
  refreshScrollTriggers,
  cleanupContext
}
