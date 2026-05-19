import { useEffect, useRef } from 'react'
import { useRouterState } from '@tanstack/react-router'
import LoadingBar, { type LoadingBarRef } from 'react-top-loading-bar'

const progressDelayMs = 100
const minVisibleMs = 220

export function NavigationProgress() {
  const ref = useRef<LoadingBarRef>(null)
  const startTimerRef = useRef<number | null>(null)
  const finishTimerRef = useRef<number | null>(null)
  const startedAtRef = useRef<number | null>(null)
  const state = useRouterState()

  useEffect(() => {
    const clearStartTimer = () => {
      if (startTimerRef.current === null) return
      window.clearTimeout(startTimerRef.current)
      startTimerRef.current = null
    }

    const clearFinishTimer = () => {
      if (finishTimerRef.current === null) return
      window.clearTimeout(finishTimerRef.current)
      finishTimerRef.current = null
    }

    if (state.status === 'pending') {
      clearFinishTimer()
      if (startedAtRef.current !== null || startTimerRef.current !== null) return

      startTimerRef.current = window.setTimeout(() => {
        startTimerRef.current = null
        startedAtRef.current = Date.now()
        ref.current?.continuousStart()
      }, progressDelayMs)
      return
    }

    clearStartTimer()
    if (startedAtRef.current === null) {
      return
    }

    const elapsed = Date.now() - startedAtRef.current
    const remaining = Math.max(minVisibleMs - elapsed, 0)
    clearFinishTimer()
    finishTimerRef.current = window.setTimeout(() => {
      finishTimerRef.current = null
      startedAtRef.current = null
      ref.current?.complete()
    }, remaining)
  }, [state.status])

  useEffect(() => {
    return () => {
      if (startTimerRef.current !== null) {
        window.clearTimeout(startTimerRef.current)
      }
      if (finishTimerRef.current !== null) {
        window.clearTimeout(finishTimerRef.current)
      }
    }
  }, [])

  return (
    <LoadingBar
      color='var(--muted-foreground)'
      ref={ref}
      shadow={true}
      height={2}
    />
  )
}
