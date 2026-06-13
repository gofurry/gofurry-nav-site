<template>
  <div
    ref="containerRef"
    class="falling-leaves-canvas"
    :data-mode="mode"
    aria-hidden="true"
  >
    <canvas ref="canvasRef" />
  </div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

// Leaf particle motion adapted from taichuy/1flowbase HeroAnimation (Apache-2.0).
type Leaf = {
  x: number
  y: number
  z: number
  vx: number
  vy: number
  baseVx: number
  baseVy: number
  rotation: number
  rotationSpeed: number
  scaleX: number
  scaleXSpeed: number
  size: number
  r: number
  g: number
  b: number
  baseAlpha: number
  veinR: number
  veinG: number
  veinB: number
  veinAlpha: number
  flutterPhase: number
  flutterSpeed: number
  isSinking: boolean
  opacity: number
}

type Ripple = {
  x: number
  y: number
  z: number
  radius: number
  maxRadius: number
  alpha: number
  speed: number
}

type ParticleColor = {
  r: number
  g: number
  b: number
  a: number
}

const props = withDefaults(defineProps<{
  leafCount?: number
  mode?: 'container' | 'viewport'
}>(), {
  leafCount: 34,
  mode: 'container',
})

const { mode } = props

const containerRef = ref<HTMLDivElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)

let frameID = 0
let resizeObserver: ResizeObserver | null = null
let reduceMotionQuery: MediaQueryList | null = null
let isReducedMotion = false
let isPageVisible = true
let time = 0
let leaves: Leaf[] = []
let ripples: Ripple[] = []
let width = 0
let height = 0

const mouse = {
  x: -1000,
  y: -1000,
  vx: 0,
  vy: 0,
}

let lastMouse: { x: number; y: number } | null = null

const leafColors = [
  { r: 128, g: 92, b: 45, a: 0.42 },
  { r: 161, g: 111, b: 46, a: 0.36 },
  { r: 92, g: 118, b: 65, a: 0.32 },
  { r: 210, g: 150, b: 73, a: 0.30 },
] satisfies readonly [ParticleColor, ...ParticleColor[]]

const veinColors = [
  { r: 95, g: 69, b: 49, a: 0.46 },
  { r: 120, g: 78, b: 38, a: 0.42 },
  { r: 71, g: 91, b: 53, a: 0.38 },
  { r: 143, g: 93, b: 42, a: 0.36 },
] satisfies readonly [ParticleColor, ...ParticleColor[]]

function randomBetween(min: number, max: number) {
  return min + Math.random() * (max - min)
}

function createLeaf(initial = false): Leaf {
  const colorIndex = Math.floor(Math.random() * leafColors.length)
  const leafColor = leafColors[colorIndex] ?? leafColors[0]
  const veinColor = veinColors[colorIndex] ?? veinColors[0]
  const z = randomBetween(0.08, 1)
  const waterY = waterLineFor(z)

  return {
    x: randomBetween(0, Math.max(width, 1)),
    y: initial ? randomBetween(0, waterY - 24) : randomBetween(-80, -24),
    z,
    vx: randomBetween(-0.18, 0.22),
    vy: randomBetween(0.18, 0.52),
    baseVx: randomBetween(0.03, 0.14),
    baseVy: randomBetween(0.22, 0.48),
    rotation: randomBetween(0, Math.PI * 2),
    rotationSpeed: randomBetween(-0.004, 0.004),
    scaleX: randomBetween(-1, 1),
    scaleXSpeed: randomBetween(0.004, 0.01),
    size: randomBetween(7, 16),
    r: leafColor.r,
    g: leafColor.g,
    b: leafColor.b,
    baseAlpha: leafColor.a,
    veinR: veinColor.r,
    veinG: veinColor.g,
    veinB: veinColor.b,
    veinAlpha: veinColor.a,
    flutterPhase: randomBetween(0, Math.PI * 2),
    flutterSpeed: randomBetween(0.004, 0.014),
    isSinking: false,
    opacity: 1,
  }
}

function waterLineFor(z: number) {
  return height * 0.74 + z * height * 0.22
}

function resetLeaves() {
  const count = Math.max(0, props.leafCount)
  leaves = Array.from({ length: count }, () => createLeaf(true))
  ripples = []
}

function resizeCanvas() {
  const container = containerRef.value
  const canvas = canvasRef.value
  const ctx = canvas?.getContext('2d')
  if (!container || !canvas || !ctx) {
    return
  }

  const rect = container.getBoundingClientRect()
  width = Math.max(1, rect.width)
  height = Math.max(1, rect.height)
  const dpr = Math.min(window.devicePixelRatio || 1, 1.5)
  canvas.width = Math.floor(width * dpr)
  canvas.height = Math.floor(height * dpr)
  canvas.style.width = `${width}px`
  canvas.style.height = `${height}px`
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)

  if (!leaves.length) {
    resetLeaves()
  }
}

function drawRipple(ctx: CanvasRenderingContext2D, ripple: Ripple) {
  const scale = 0.28 + ripple.z * 0.72
  const perspectiveY = 0.12 + ripple.z * 0.18

  ctx.save()
  for (let i = 0; i < 3; i += 1) {
    const radius = ripple.radius - i * 8
    if (radius <= 0) {
      continue
    }
    const localAlpha = (1 - radius / ripple.maxRadius) * ripple.alpha
    if (localAlpha <= 0) {
      continue
    }

    ctx.beginPath()
    ctx.ellipse(ripple.x, ripple.y, radius * scale, radius * scale * perspectiveY, 0, 0, Math.PI * 2)
    ctx.strokeStyle = `rgba(143, 93, 42, ${localAlpha * (i === 0 ? 0.20 : 0.10)})`
    ctx.lineWidth = i === 0 ? 0.9 * scale : 0.55 * scale
    ctx.stroke()
  }
  ctx.restore()
}

function updateRipple(ripple: Ripple) {
  ripple.radius += ripple.speed
  ripple.alpha = 1 - ripple.radius / ripple.maxRadius
  return ripple.alpha > 0
}

function updateLeaf(leaf: Leaf) {
  const renderScale = 0.3 + leaf.z * 0.7
  const waterY = waterLineFor(leaf.z)

  if (!leaf.isSinking && leaf.y >= waterY) {
    leaf.isSinking = true
    leaf.vy = 0.055
    leaf.vx *= 0.5
    ripples.push({
      x: leaf.x,
      y: waterY,
      z: leaf.z,
      radius: 1,
      maxRadius: leaf.size * 3.5,
      alpha: 0.42,
      speed: 0.55,
    })
  }

  if (leaf.isSinking) {
    leaf.vx *= 0.93
    leaf.vy = 0.055
    leaf.rotationSpeed *= 0.92
    leaf.opacity -= 0.012

    if (leaf.opacity <= 0) {
      Object.assign(leaf, createLeaf(false))
      return
    }
  } else {
    leaf.vy += (leaf.baseVy - leaf.vy) * 0.04
    leaf.vx += (leaf.baseVx - leaf.vx) * 0.04

    const sway = Math.sin(time * leaf.flutterSpeed + leaf.flutterPhase) * 0.22
    leaf.vx += sway * 0.12
    leaf.rotation += leaf.rotationSpeed + sway * 0.003
    leaf.scaleX = Math.sin(time * leaf.scaleXSpeed + leaf.flutterPhase)

    const dx = leaf.x - mouse.x
    const dy = leaf.y - mouse.y
    const radius = 130 * renderScale
    const distSq = dx * dx + dy * dy

    if (distSq < radius * radius) {
      const dist = Math.sqrt(distSq)
      const force = (1 - dist / radius) * renderScale
      const dirX = dx / (dist || 1)
      const dirY = dy / (dist || 1)
      leaf.vx += mouse.vx * force * 0.35 + dirX * force * 0.28
      leaf.vy += mouse.vy * force * 0.35 + dirY * force * 0.12
    }

    leaf.vx *= 0.97
    leaf.vy *= 0.97
    const speed = Math.hypot(leaf.vx, leaf.vy)
    const maxSpeed = 4.8
    if (speed > maxSpeed) {
      leaf.vx = (leaf.vx / speed) * maxSpeed
      leaf.vy = (leaf.vy / speed) * maxSpeed
    }
  }

  leaf.x += leaf.vx * renderScale
  leaf.y += leaf.vy * renderScale

  if (leaf.x < -28) {
    leaf.x = width + 18
  } else if (leaf.x > width + 28) {
    leaf.x = -18
  }
}

function drawLeaf(ctx: CanvasRenderingContext2D, leaf: Leaf) {
  const renderScale = 0.3 + leaf.z * 0.7
  const depthAlpha = 0.42 + leaf.z * 0.58

  ctx.save()
  ctx.translate(leaf.x, leaf.y)
  ctx.rotate(leaf.rotation)
  const scaleDirection = leaf.scaleX < 0 ? -1 : 1
  const scaleX = scaleDirection * Math.max(Math.abs(leaf.scaleX), 0.16) * renderScale
  ctx.scale(scaleX, renderScale)

  ctx.shadowColor = `rgba(${leaf.r}, ${leaf.g}, ${leaf.b}, ${0.20 * leaf.opacity})`
  ctx.shadowBlur = 7 * renderScale

  ctx.beginPath()
  ctx.moveTo(0, -leaf.size)
  ctx.quadraticCurveTo(-leaf.size * 0.72, -leaf.size * 0.18, 0, leaf.size)
  ctx.quadraticCurveTo(leaf.size * 0.72, -leaf.size * 0.18, 0, -leaf.size)
  ctx.fillStyle = `rgba(${leaf.r}, ${leaf.g}, ${leaf.b}, ${leaf.opacity * leaf.baseAlpha * depthAlpha})`
  ctx.fill()

  ctx.shadowBlur = 0
  ctx.beginPath()
  ctx.moveTo(0, -leaf.size)
  ctx.lineTo(0, leaf.size * 0.82)
  ctx.strokeStyle = `rgba(${leaf.veinR}, ${leaf.veinG}, ${leaf.veinB}, ${leaf.opacity * leaf.veinAlpha * depthAlpha})`
  ctx.lineWidth = 0.75 * renderScale
  ctx.stroke()

  ctx.restore()
}

function animate() {
  const canvas = canvasRef.value
  const ctx = canvas?.getContext('2d')
  if (!canvas || !ctx) {
    return
  }

  if (!isPageVisible || isReducedMotion) {
    frameID = requestAnimationFrame(animate)
    return
  }

  time += 1
  ctx.clearRect(0, 0, width, height)

  mouse.vx *= 0.94
  mouse.vy *= 0.94

  ripples = ripples.filter((ripple) => {
    const keep = updateRipple(ripple)
    if (keep) {
      drawRipple(ctx, ripple)
    }
    return keep
  })

  leaves.forEach(updateLeaf)
  leaves
    .slice()
    .sort((a, b) => a.z - b.z)
    .forEach((leaf) => drawLeaf(ctx, leaf))

  frameID = requestAnimationFrame(animate)
}

function handleMouseMove(event: MouseEvent) {
  const rect = containerRef.value?.getBoundingClientRect()
  if (!rect) {
    return
  }
  const x = event.clientX - rect.left
  const y = event.clientY - rect.top

  if (lastMouse) {
    mouse.vx = x - lastMouse.x
    mouse.vy = y - lastMouse.y
  }

  mouse.x = x
  mouse.y = y
  lastMouse = { x, y }
}

function handleMouseLeave() {
  mouse.x = -1000
  mouse.y = -1000
  mouse.vx = 0
  mouse.vy = 0
  lastMouse = null
}

function handleVisibilityChange() {
  isPageVisible = document.visibilityState === 'visible'
}

function handleReduceMotionChange(event: MediaQueryListEvent) {
  isReducedMotion = event.matches
}

onMounted(() => {
  reduceMotionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
  isReducedMotion = reduceMotionQuery.matches
  reduceMotionQuery.addEventListener('change', handleReduceMotionChange)

  resizeCanvas()
  resizeObserver = new ResizeObserver(resizeCanvas)
  if (containerRef.value) {
    resizeObserver.observe(containerRef.value)
  }

  window.addEventListener('mousemove', handleMouseMove, { passive: true })
  window.addEventListener('mouseleave', handleMouseLeave)
  document.addEventListener('visibilitychange', handleVisibilityChange)
  frameID = requestAnimationFrame(animate)
})

onBeforeUnmount(() => {
  if (frameID) {
    cancelAnimationFrame(frameID)
  }
  resizeObserver?.disconnect()
  reduceMotionQuery?.removeEventListener('change', handleReduceMotionChange)
  window.removeEventListener('mousemove', handleMouseMove)
  window.removeEventListener('mouseleave', handleMouseLeave)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<style scoped>
.falling-leaves-canvas {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
}

.falling-leaves-canvas[data-mode="viewport"] {
  position: fixed;
}

.falling-leaves-canvas canvas {
  display: block;
  width: 100%;
  height: 100%;
}
</style>
