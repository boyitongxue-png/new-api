/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useEffect, useState, type RefObject } from 'react'

const SCRAMBLE_GLYPHS = 'abcdefghijklmnopqrstuvwxyz0123456789'
const SCRAMBLE_KEEP = '/:*.​'
const SCRAMBLE_DURATION_MS = 520
const INTRO_FALLBACK_MS = 1200

type Cleanup = () => void

function prefersReducedMotion(): boolean {
  return window.matchMedia('(prefers-reduced-motion: reduce)').matches
}

function hasFinePointer(): boolean {
  return window.matchMedia('(hover: hover) and (pointer: fine)').matches
}

/**
 * Hold the CSS load choreography until the landing is on screen and the
 * brand fonts are in, so headline wipes never run against fallback glyphs.
 * Returns `true` while the intro is still held.
 */
export function useLandingIntroHold(): boolean {
  const [holding, setHolding] = useState(true)

  useEffect(() => {
    let released = false
    let frame = 0
    const release = () => {
      if (released) return
      released = true
      frame = window.requestAnimationFrame(() => {
        frame = window.requestAnimationFrame(() => setHolding(false))
      })
    }
    const fallbackId = window.setTimeout(release, INTRO_FALLBACK_MS)
    const start = () => {
      if (document.visibilityState !== 'visible') return
      document.removeEventListener('visibilitychange', start)
      if (document.fonts && document.fonts.ready) {
        document.fonts.ready.then(release, release)
      } else {
        release()
      }
    }
    if (document.visibilityState === 'visible') {
      start()
    } else {
      document.addEventListener('visibilitychange', start)
    }
    return () => {
      released = true
      window.clearTimeout(fallbackId)
      window.cancelAnimationFrame(frame)
      document.removeEventListener('visibilitychange', start)
    }
  }, [])

  return holding
}

function bindHeroParallax(root: HTMLElement): Cleanup | undefined {
  const hero = root.querySelector<HTMLElement>('.amx-hero')
  const mix = root.querySelector<HTMLElement>('.amx-hero-mix')
  const glow = root.querySelector<HTMLElement>('.amx-hero-glow')
  if (!hero || !mix || !glow) return undefined

  const target = { x: 0, y: 0, gx: 0, gy: 0 }
  const current = { x: 0, y: 0, gx: 0, gy: 0 }
  let frame = 0
  let running = false

  const tick = () => {
    current.x += (target.x - current.x) * 0.06
    current.y += (target.y - current.y) * 0.06
    current.gx += (target.gx - current.gx) * 0.12
    current.gy += (target.gy - current.gy) * 0.12
    mix.style.transform = `translate3d(${current.x.toFixed(2)}px,${current.y.toFixed(2)}px,0)`
    glow.style.transform = `translate3d(${current.gx.toFixed(1)}px,${current.gy.toFixed(1)}px,0)`
    const delta =
      Math.abs(target.x - current.x) +
      Math.abs(target.y - current.y) +
      Math.abs(target.gx - current.gx) +
      Math.abs(target.gy - current.gy)
    if (delta > 0.1) {
      frame = window.requestAnimationFrame(tick)
    } else {
      running = false
    }
  }
  const kick = () => {
    if (running) return
    running = true
    frame = window.requestAnimationFrame(tick)
  }
  const onMove = (event: PointerEvent) => {
    const rect = hero.getBoundingClientRect()
    target.x = ((event.clientX - rect.left) / rect.width - 0.5) * -26
    target.y = ((event.clientY - rect.top) / rect.height - 0.5) * -16
    target.gx = event.clientX - rect.left
    target.gy = event.clientY - rect.top
    if (!hero.classList.contains('is-lit')) {
      current.gx = target.gx
      current.gy = target.gy
      hero.classList.add('is-lit')
    }
    kick()
  }
  const onLeave = () => {
    hero.classList.remove('is-lit')
    target.x = 0
    target.y = 0
    kick()
  }

  hero.addEventListener('pointermove', onMove)
  hero.addEventListener('pointerleave', onLeave)
  return () => {
    hero.removeEventListener('pointermove', onMove)
    hero.removeEventListener('pointerleave', onLeave)
    window.cancelAnimationFrame(frame)
    mix.style.transform = ''
    glow.style.transform = ''
  }
}

function bindMagneticButtons(root: HTMLElement): Cleanup {
  const cleanups: Cleanup[] = []
  root.querySelectorAll<HTMLElement>('.amx-magnetic').forEach((button) => {
    const label = button.querySelector<HTMLElement>('.amx-btn-label')
    const onMove = (event: PointerEvent) => {
      const rect = button.getBoundingClientRect()
      const x = event.clientX - (rect.left + rect.width / 2)
      const y = event.clientY - (rect.top + rect.height / 2)
      button.style.transitionDuration = '.25s'
      button.style.transform = `translate3d(${(x * 0.2).toFixed(1)}px,${(y * 0.32).toFixed(1)}px,0)`
      if (label) {
        label.style.transitionDuration = '.25s'
        label.style.transform = `translate3d(${(x * 0.08).toFixed(1)}px,${(y * 0.12).toFixed(1)}px,0)`
      }
    }
    const onLeave = () => {
      button.style.transitionDuration = ''
      button.style.transform = ''
      if (label) {
        label.style.transitionDuration = ''
        label.style.transform = ''
      }
    }
    button.addEventListener('pointermove', onMove)
    button.addEventListener('pointerleave', onLeave)
    cleanups.push(() => {
      button.removeEventListener('pointermove', onMove)
      button.removeEventListener('pointerleave', onLeave)
      onLeave()
    })
  })
  return () => cleanups.forEach((cleanup) => cleanup())
}

/**
 * Decode effect on endpoint paths. Mutates the existing text node (instead of
 * `textContent`) so React keeps ownership of the node it rendered.
 */
function bindEndpointScramble(root: HTMLElement): Cleanup {
  const cleanups: Cleanup[] = []
  root.querySelectorAll<HTMLElement>('.amx-ep').forEach((row) => {
    const pathEl = row.querySelector<HTMLElement>('[data-scramble]')
    const textNode = pathEl?.firstChild
    const finalText = pathEl?.dataset.scramble
    if (!textNode || textNode.nodeType !== Node.TEXT_NODE || !finalText) {
      return
    }
    let frame = 0
    const run = () => {
      window.cancelAnimationFrame(frame)
      const startedAt = performance.now()
      const length = finalText.length
      const cache: { key: number; glyph: string }[] = []
      const step = (now: number) => {
        const progress = (now - startedAt) / SCRAMBLE_DURATION_MS
        const refresh = Math.floor((now - startedAt) / 45)
        let out = ''
        for (let i = 0; i < length; i++) {
          const char = finalText.charAt(i)
          const settleAt = 0.18 + (i / length) * 0.82
          if (progress >= settleAt || SCRAMBLE_KEEP.includes(char)) {
            out += char
            continue
          }
          if (!cache[i] || cache[i].key !== refresh) {
            cache[i] = {
              key: refresh,
              glyph: SCRAMBLE_GLYPHS.charAt(
                Math.floor(Math.random() * SCRAMBLE_GLYPHS.length)
              ),
            }
          }
          out += cache[i].glyph
        }
        textNode.nodeValue = out
        if (progress < 1) {
          frame = window.requestAnimationFrame(step)
        } else {
          textNode.nodeValue = finalText
        }
      }
      frame = window.requestAnimationFrame(step)
    }
    row.addEventListener('mouseenter', run)
    row.addEventListener('focus', run)
    cleanups.push(() => {
      row.removeEventListener('mouseenter', run)
      row.removeEventListener('focus', run)
      window.cancelAnimationFrame(frame)
      textNode.nodeValue = finalText
    })
  })
  return () => cleanups.forEach((cleanup) => cleanup())
}

/**
 * Scroll reveals are armed only after the first scroll and only for blocks
 * still below the fold, so a page that is never scrolled (thumbnails,
 * full-page captures) always shows its content.
 */
function bindScrollReveals(root: HTMLElement): Cleanup | undefined {
  const blocks = Array.from(root.querySelectorAll<HTMLElement>('[data-reveal]'))
  if (blocks.length === 0) return undefined

  let armed = false
  let pending: HTMLElement[] = []
  let checking = false
  let frame = 0

  const check = () => {
    checking = false
    const line = window.innerHeight * 0.9
    pending = pending.filter((block) => {
      if (block.getBoundingClientRect().top < line) {
        block.classList.remove('is-armed')
        return false
      }
      return true
    })
  }
  const onScroll = () => {
    if (!armed) {
      armed = true
      const viewport = window.innerHeight
      blocks.forEach((block) => {
        if (block.getBoundingClientRect().top > viewport) {
          block.classList.add('is-armed')
          pending.push(block)
        }
      })
    }
    if (pending.length > 0 && !checking) {
      checking = true
      frame = window.requestAnimationFrame(check)
    }
  }

  window.addEventListener('scroll', onScroll, { passive: true })
  return () => {
    window.removeEventListener('scroll', onScroll)
    window.cancelAnimationFrame(frame)
    blocks.forEach((block) => block.classList.remove('is-armed'))
  }
}

/** Footer wordmark letters settle as the page bottoms out. */
function bindWordmarkLift(root: HTMLElement): Cleanup | undefined {
  const letters = Array.from(root.querySelectorAll<HTMLElement>('.amx-wm-l'))
  if (letters.length === 0) return undefined

  let ticking = false
  let frame = 0
  const update = () => {
    ticking = false
    const doc = document.documentElement
    const remaining = doc.scrollHeight - (window.scrollY + window.innerHeight)
    const range = Math.max(240, window.innerHeight * 0.5)
    const progress = 1 - Math.min(Math.max(remaining / range, 0), 1)
    const eased = 1 - Math.pow(1 - progress, 3)
    letters.forEach((letter, index) => {
      const offset = (1 - eased) * (0.32 + index * 0.07)
      letter.style.transform =
        offset > 0.001 ? `translate3d(0,${offset.toFixed(3)}em,0)` : ''
    })
  }
  const onScroll = () => {
    if (ticking) return
    ticking = true
    frame = window.requestAnimationFrame(update)
  }

  update()
  window.addEventListener('scroll', onScroll, { passive: true })
  window.addEventListener('resize', onScroll)
  return () => {
    window.removeEventListener('scroll', onScroll)
    window.removeEventListener('resize', onScroll)
    window.cancelAnimationFrame(frame)
    letters.forEach((letter) => {
      letter.style.transform = ''
    })
  }
}

/**
 * Wires the landing page's pointer and scroll motion (hero parallax + glow,
 * magnetic CTAs, endpoint scramble, scroll reveals, wordmark lift).
 * Everything is skipped under `prefers-reduced-motion`; pointer effects only
 * run on fine pointers. All listeners and inline styles are cleaned up.
 */
export function useLandingMotion(rootRef: RefObject<HTMLElement | null>) {
  useEffect(() => {
    const root = rootRef.current
    if (!root || prefersReducedMotion()) return

    const cleanups: (Cleanup | undefined)[] = [
      bindEndpointScramble(root),
      bindScrollReveals(root),
      bindWordmarkLift(root),
    ]
    if (hasFinePointer()) {
      cleanups.push(bindHeroParallax(root), bindMagneticButtons(root))
    }
    return () => cleanups.forEach((cleanup) => cleanup?.())
  }, [rootRef])
}
