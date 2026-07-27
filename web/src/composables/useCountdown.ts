import { onUnmounted } from 'vue'

export function useCountdown(seconds: number, onExpire: () => void) {
  let timeout: number | null = null

  function start() {
    timeout = window.setTimeout(onExpire, seconds * 1000)
  }

  function stop() {
    if (timeout) {
      window.clearTimeout(timeout)
      timeout = null
    }
  }

  onUnmounted(stop)

  return { start, stop }
}
