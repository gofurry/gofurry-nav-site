export const NAV_PAGE_REVEAL_EVENT = 'nav-page-reveal-change'

const revealLocks = new Set<string>()

export function dispatchNavPageReveal(visible: boolean) {
  window.dispatchEvent(
    new CustomEvent(NAV_PAGE_REVEAL_EVENT, {
      detail: { visible },
    })
  )
}

export function setNavPageRevealLock(key: string, locked: boolean) {
  if (locked) {
    revealLocks.add(key)
  } else {
    revealLocks.delete(key)
  }
}

export function isNavPageRevealLocked() {
  return revealLocks.size > 0
}
