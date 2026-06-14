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
  opacity: number
}

type ParticleColor = {
  r: number
  g: number
  b: number
  a: number
}

const props = withDefaults(defineProps<{
  leafCount?: number
  mobileLeafCount?: number
  fps?: number
  interactive?: boolean
  palette?: 'warm' | 'bright'
  mode?: 'container' | 'viewport'
}>(), {
  leafCount: 34,
  mobileLeafCount: 12,
  fps: 24,
  interactive: false,
  palette: 'warm',
  mode: 'container',
})

const { mode } = props

const containerRef = ref<HTMLDivElement | null>(null)
const canvasRef = ref<HTMLCanvasElement | null>(null)

let frameID = 0
let resizeObserver: ResizeObserver | null = null
let reduceMotionQuery: MediaQueryList | null = null
let mobileQuery: MediaQueryList | null = null
let isReducedMotion = false
let isPageVisible = true
let isSmallScreen = false
let time = 0
let lastFrameTime = 0
let lastUpdateTime = 0
let leaves: Leaf[] = []
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

const brightLeafColors = [
  { r: 238, g: 246, b: 250, a: 0.52 },
  { r: 218, g: 235, b: 242, a: 0.48 },
  { r: 239, g: 222, b: 173, a: 0.44 },
  { r: 184, g: 215, b: 206, a: 0.46 },
] satisfies readonly [ParticleColor, ...ParticleColor[]]

const brightVeinColors = [
  { r: 190, g: 213, b: 222, a: 0.54 },
  { r: 164, g: 194, b: 204, a: 0.50 },
  { r: 197, g: 174, b: 126, a: 0.48 },
  { r: 137, g: 177, b: 169, a: 0.46 },
] satisfies readonly [ParticleColor, ...ParticleColor[]]

function randomBetween(min: number, max: number) {
  return min + Math.random() * (max - min)
}

function createLeaf(initial = false): Leaf {
  const activeLeafColors = props.palette === 'bright' ? brightLeafColors : leafColors
  const activeVeinColors = props.palette === 'bright' ? brightVeinColors : veinColors
  const colorIndex = Math.floor(Math.random() * activeLeafColors.length)
  const leafColor = activeLeafColors[colorIndex] ?? activeLeafColors[0]
  const veinColor = activeVeinColors[colorIndex] ?? activeVeinColors[0]
  const z = randomBetween(0.08, 1)

  return {
    x: randomBetween(0, Math.max(width, 1)),
    y: initial ? randomBetween(0, Math.max(height, 1)) : randomBetween(-80, -24),
    z,
    vx: randomBetween(-0.18, 0.22),
    vy: randomBetween(0.18, 0.52),
    baseVx: randomBetween(0.04, 0.16),
    baseVy: randomBetween(0.28, 0.62),
    rotation: randomBetween(0, Math.PI * 2),
    rotationSpeed: randomBetween(-0.004, 0.004),
    scaleX: randomBetween(-1, 1),
    scaleXSpeed: randomBetween(0.004, 0.01),
    size: randomBetween(10, 18),
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
    opacity: 1,
  }
}

function targetLeafCount() {
  return Math.max(0, isSmallScreen ? props.mobileLeafCount : props.leafCount)
}

function resetLeaves() {
  const count = targetLeafCount()
  leaves = Array.from({ length: count }, () => createLeaf(true)).sort((a, b) => a.z - b.z)
}

function syncLeafCount() {
  const count = targetLeafCount()
  if (leaves.length > count) {
    leaves = leaves.slice(0, count)
    return
  }

  while (leaves.length < count) {
    leaves.push(createLeaf(true))
  }
  leaves.sort((a, b) => a.z - b.z)
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
  const dpr = 1
  canvas.width = Math.floor(width * dpr)
  canvas.height = Math.floor(height * dpr)
  canvas.style.width = `${width}px`
  canvas.style.height = `${height}px`
  ctx.setTransform(dpr, 0, 0, dpr, 0, 0)

  if (!leaves.length) {
    resetLeaves()
  } else {
    syncLeafCount()
  }
}

function updateLeaf(leaf: Leaf, deltaScale: number) {
  const renderScale = 0.3 + leaf.z * 0.7
  const adjustedDelta = Math.min(deltaScale, 2.4)

  leaf.vy += (leaf.baseVy - leaf.vy) * 0.04 * adjustedDelta
  leaf.vx += (leaf.baseVx - leaf.vx) * 0.04 * adjustedDelta

  const sway = Math.sin(time * leaf.flutterSpeed + leaf.flutterPhase) * 0.22
  leaf.vx += sway * 0.12 * adjustedDelta
  leaf.rotation += (leaf.rotationSpeed + sway * 0.003) * adjustedDelta
  leaf.scaleX = Math.sin(time * leaf.scaleXSpeed + leaf.flutterPhase)

  if (props.interactive) {
    const dx = leaf.x - mouse.x
    const dy = leaf.y - mouse.y
    const radius = 110 * renderScale
    const distSq = dx * dx + dy * dy

    if (distSq < radius * radius) {
      const dist = Math.sqrt(distSq)
      const force = (1 - dist / radius) * renderScale
      const dirX = dx / (dist || 1)
      const dirY = dy / (dist || 1)
      leaf.vx += (mouse.vx * force * 0.24 + dirX * force * 0.20) * adjustedDelta
      leaf.vy += (mouse.vy * force * 0.24 + dirY * force * 0.09) * adjustedDelta
    }
  }

  const damping = Math.pow(0.97, adjustedDelta)
  leaf.vx *= damping
  leaf.vy *= damping
  const speed = Math.hypot(leaf.vx, leaf.vy)
  const maxSpeed = 3.8
  if (speed > maxSpeed) {
    leaf.vx = (leaf.vx / speed) * maxSpeed
    leaf.vy = (leaf.vy / speed) * maxSpeed
  }

  leaf.x += leaf.vx * renderScale * adjustedDelta
  leaf.y += leaf.vy * renderScale * adjustedDelta

  if (leaf.x < -28) {
    leaf.x = width + 18
  } else if (leaf.x > width + 28) {
    leaf.x = -18
  }

  if (leaf.y > height + leaf.size * 2) {
    Object.assign(leaf, createLeaf(false))
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

  ctx.beginPath()
  ctx.moveTo(0, -leaf.size)
  ctx.quadraticCurveTo(-leaf.size * 0.72, -leaf.size * 0.18, 0, leaf.size)
  ctx.quadraticCurveTo(leaf.size * 0.72, -leaf.size * 0.18, 0, -leaf.size)
  ctx.fillStyle = `rgba(${leaf.r}, ${leaf.g}, ${leaf.b}, ${leaf.opacity * leaf.baseAlpha * depthAlpha})`
  ctx.fill()

  ctx.beginPath()
  ctx.moveTo(0, -leaf.size)
  ctx.lineTo(0, leaf.size * 0.82)
  ctx.strokeStyle = `rgba(${leaf.veinR}, ${leaf.veinG}, ${leaf.veinB}, ${leaf.opacity * leaf.veinAlpha * depthAlpha})`
  ctx.lineWidth = 0.75 * renderScale
  ctx.stroke()

  ctx.restore()
}

function animate(timestamp = 0) {
  const canvas = canvasRef.value
  const ctx = canvas?.getContext('2d')
  if (!canvas || !ctx) {
    return
  }

  if (!isPageVisible || isReducedMotion) {
    return
  }

  const frameInterval = 1000 / Math.max(1, props.fps)
  if (lastFrameTime && timestamp - lastFrameTime < frameInterval) {
    frameID = requestAnimationFrame(animate)
    return
  }
  const elapsed = lastUpdateTime ? timestamp - lastUpdateTime : 16.67
  lastFrameTime = timestamp
  lastUpdateTime = timestamp
  const deltaScale = elapsed / 16.67

  time += deltaScale
  ctx.clearRect(0, 0, width, height)

  if (props.interactive) {
    mouse.vx *= 0.94
    mouse.vy *= 0.94
  }

  leaves.forEach((leaf) => updateLeaf(leaf, deltaScale))
  leaves.forEach((leaf) => drawLeaf(ctx, leaf))

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
  if (isPageVisible) {
    startAnimation()
  } else {
    stopAnimation()
  }
}

function handleReduceMotionChange(event: MediaQueryListEvent) {
  isReducedMotion = event.matches
  if (isReducedMotion) {
    stopAnimation()
    clearCanvas()
  } else {
    startAnimation()
  }
}

function handleMobileChange(event: MediaQueryListEvent) {
  isSmallScreen = event.matches
  syncLeafCount()
}

function clearCanvas() {
  const canvas = canvasRef.value
  const ctx = canvas?.getContext('2d')
  if (!canvas || !ctx) {
    return
  }
  ctx.clearRect(0, 0, width, height)
}

function startAnimation() {
  if (frameID || isReducedMotion || !isPageVisible) {
    return
  }
  lastFrameTime = 0
  lastUpdateTime = 0
  frameID = requestAnimationFrame(animate)
}

function stopAnimation() {
  if (!frameID) {
    return
  }
  cancelAnimationFrame(frameID)
  frameID = 0
}

onMounted(() => {
  reduceMotionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
  isReducedMotion = reduceMotionQuery.matches
  reduceMotionQuery.addEventListener('change', handleReduceMotionChange)

  mobileQuery = window.matchMedia('(max-width: 640px)')
  isSmallScreen = mobileQuery.matches
  mobileQuery.addEventListener('change', handleMobileChange)

  resizeCanvas()
  resizeObserver = new ResizeObserver(resizeCanvas)
  if (containerRef.value) {
    resizeObserver.observe(containerRef.value)
  }

  if (props.interactive) {
    window.addEventListener('mousemove', handleMouseMove, { passive: true })
    window.addEventListener('mouseleave', handleMouseLeave)
  }
  document.addEventListener('visibilitychange', handleVisibilityChange)
  startAnimation()
})

onBeforeUnmount(() => {
  stopAnimation()
  resizeObserver?.disconnect()
  reduceMotionQuery?.removeEventListener('change', handleReduceMotionChange)
  mobileQuery?.removeEventListener('change', handleMobileChange)
  if (props.interactive) {
    window.removeEventListener('mousemove', handleMouseMove)
    window.removeEventListener('mouseleave', handleMouseLeave)
  }
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
