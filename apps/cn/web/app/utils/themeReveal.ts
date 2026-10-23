// The Store owns the theme. This controller owns only one optional snapshot.
export function createThemeReveal() {
  type Job = {
    apply: () => Promise<void>
    transition?: ViewTransition
    animation?: Animation
    ending: boolean
    release: () => void
  }
  let active: Job | undefined
  let disposed = false

  async function finish(job: Job, skip = false) {
    if (active !== job || job.ending) return
    job.ending = true
    // A skipped snapshot may never call its update callback. Apply exactly once.
    const applied = job.apply()
    if (skip) job.transition?.skipTransition()
    job.animation?.cancel()
    try { await applied } catch { /* Snapshot failure must not strand the UI. */ }
    finally {
      job.release()
      document.documentElement.removeAttribute('data-gf-theme-reveal')
      if (active === job) active = undefined
    }
  }
  const cancel = () => { if (active) void finish(active, true) }

  function toggle(button: HTMLElement | null, update: () => Promise<void>) {
    if (active || disposed) return
    const root = document.documentElement
    const media = window.matchMedia?.('(prefers-reduced-motion: reduce)')
    const rect = button?.getBoundingClientRect()
    const x = rect ? rect.left + rect.width / 2 : NaN
    const y = rect ? rect.top + rect.height / 2 : NaN
    if (!document.startViewTransition || !root.animate || !media || media.matches || document.hidden
      || !window.CSS?.supports('clip-path', 'circle(1px at 0px 0px)')
      || !button?.isConnected || button.matches(':disabled') || !rect || rect.width <= 0 || rect.height <= 0
      || !Number.isFinite(x) || !Number.isFinite(y) || x < 0 || y < 0 || x > innerWidth || y > innerHeight) {
      void update().then(() => {
        if (media?.matches) settleThemeColors()
      }).catch(() => {})
      return
    }
    const radius = Math.hypot(Math.max(x, innerWidth - x), Math.max(y, innerHeight - y))
    let applied: Promise<void> | undefined
    const job: Job = {
      ending: false,
      apply() {
        // Cache before invoking user code, including a synchronous Store error.
        if (!applied) {
          let resolve!: () => void, reject!: (reason: unknown) => void
          applied = new Promise<void>((done, fail) => { resolve = done; reject = fail })
          try {
            void update().then(() => {
              if (active === job) settleThemeColors()
            }).then(resolve, reject)
          } catch (error) { reject(error) }
        }
        return applied
      },
      release() {
        media.removeEventListener('change', reduced)
        document.removeEventListener('visibilitychange', hidden)
        window.removeEventListener('resize', cancel)
        window.removeEventListener('pagehide', cancel)
        window.removeEventListener('pageswap', cancel)
      },
    }
    const reduced = () => { if (media.matches) cancel() }
    const hidden = () => { if (document.hidden) cancel() }
    active = job
    root.setAttribute('data-gf-theme-reveal', '')
    media.addEventListener('change', reduced)
    document.addEventListener('visibilitychange', hidden)
    window.addEventListener('resize', cancel)
    window.addEventListener('pagehide', cancel)
    window.addEventListener('pageswap', cancel)
    try {
      const transition = document.startViewTransition(job.apply)
      job.transition = transition
      void transition.ready.then(() => {
        if (active !== job || job.ending) return
        if (document.hidden || media.matches) { cancel(); return }
        job.animation = root.animate({ clipPath: [
          `circle(0px at ${x}px ${y}px)`, `circle(${radius}px at ${x}px ${y}px)`,
        ] }, {
          duration: 420, easing: 'cubic-bezier(0.22, 1, 0.36, 1)',
          pseudoElement: '::view-transition-new(root)', fill: 'both',
        })
        void job.animation.finished.catch(() => { void finish(job, true) })
      }).catch(() => { void finish(job, true) })
      void transition.updateCallbackDone.catch(() => { void finish(job, true) })
      void transition.finished.then(() => finish(job), () => finish(job, true))
    } catch { void finish(job, true) }
  }

  return { toggle, cancel, dispose() { disposed = true; cancel() } }
}

function settleThemeColors() {
  // getAnimations flushes pending styles. Finish only color/surface transitions,
  // including teleported panels; leave opacity, transforms and CSS animations alone.
  const properties = new Set(['color', 'background-color', 'border-top-color', 'border-right-color',
    'border-bottom-color', 'border-left-color', 'outline-color', 'text-decoration-color',
    'fill', 'stroke', 'box-shadow', 'text-shadow', '--tw-gradient-from', '--tw-gradient-via', '--tw-gradient-to'])
  for (const animation of document.getAnimations()) {
    if (properties.has((animation as CSSTransition).transitionProperty)) {
      try { animation.finish() } catch { /* A concurrently removed node is harmless. */ }
    }
  }
}
