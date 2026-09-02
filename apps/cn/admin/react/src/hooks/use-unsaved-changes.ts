import { useEffect } from 'react'
import { useBlocker } from 'react-router-dom'

export function useUnsavedChanges(dirty: boolean) {
  const blocker = useBlocker(dirty)
  useEffect(() => {
    const beforeUnload = (event: BeforeUnloadEvent) => { if (dirty) event.preventDefault() }
    window.addEventListener('beforeunload', beforeUnload)
    return () => window.removeEventListener('beforeunload', beforeUnload)
  }, [dirty])
  useEffect(() => {
    if (blocker.state !== 'blocked') return
    if (window.confirm('存在未保存的修改，确定离开当前页面吗？')) blocker.proceed()
    else blocker.reset()
  }, [blocker])
}
