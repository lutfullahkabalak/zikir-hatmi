export const installNoZoom = () => {
  if (typeof document === 'undefined') return

  let lastTouchEnd = 0
  document.addEventListener(
    'touchend',
    (event) => {
      const now = Date.now()
      if (now - lastTouchEnd <= 350) {
        event.preventDefault()
      }
      lastTouchEnd = now
    },
    { passive: false },
  )

  document.addEventListener(
    'gesturestart',
    (event) => event.preventDefault(),
    { passive: false },
  )

  document.addEventListener(
    'dblclick',
    (event) => event.preventDefault(),
    { passive: false },
  )
}
