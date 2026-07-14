import {
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react"
import { apiHealth } from "@/api/client"
import { BackendHealthContext } from "./useBackendHealth"

const POLL_INTERVAL_MS = 5000
const ORKAI_HEALTH_URL = "/v1/api/health/orkai"

function BackendHealthProvider({ children }: { children: React.ReactNode }) {
  const [isBackendReachable, setIsBackendReachable] = useState<boolean | null>(null)
  const [isOrkaiRunning, setIsOrkaiRunning] = useState<boolean | null>(null)
  const backendIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const orkaiIntervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const checkBackend = useCallback(async () => {
    const reachable = await apiHealth()
    setIsBackendReachable(reachable)
    return reachable
  }, [])

  const checkOrkai = useCallback(async () => {
    try {
      const res = await fetch(ORKAI_HEALTH_URL, { cache: "no-store" })
      const body = await res.json()
      const running = body?.data?.running === true
      setIsOrkaiRunning(running)
      return running
    } catch {
      setIsOrkaiRunning(false)
      return false
    }
  }, [])

  useEffect(() => {
    async function init() {
      const reachable = await checkBackend()
      if (reachable) {
        await checkOrkai()
      } else {
        backendIntervalRef.current = setInterval(async () => {
          const nowReachable = await checkBackend()
          if (nowReachable && backendIntervalRef.current) {
            clearInterval(backendIntervalRef.current)
            backendIntervalRef.current = null
            await checkOrkai()
            orkaiIntervalRef.current = setInterval(async () => {
              const running = await checkOrkai()
              if (running && orkaiIntervalRef.current) {
                clearInterval(orkaiIntervalRef.current)
                orkaiIntervalRef.current = null
              }
            }, POLL_INTERVAL_MS)
          }
        }, POLL_INTERVAL_MS)
      }
    }

    init()

    return () => {
      if (backendIntervalRef.current) {
        clearInterval(backendIntervalRef.current)
        backendIntervalRef.current = null
      }
      if (orkaiIntervalRef.current) {
        clearInterval(orkaiIntervalRef.current)
        orkaiIntervalRef.current = null
      }
    }
  }, [checkBackend, checkOrkai])

  useEffect(() => {
    const onVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        checkBackend().then((reachable) => {
          if (reachable) checkOrkai()
        })
      }
    }

    const onPopState = () => {
      checkBackend().then((reachable) => {
        if (reachable) checkOrkai()
      })
    }

    document.addEventListener("visibilitychange", onVisibilityChange)
    window.addEventListener("popstate", onPopState)

    return () => {
      document.removeEventListener("visibilitychange", onVisibilityChange)
      window.removeEventListener("popstate", onPopState)
    }
  }, [checkBackend, checkOrkai])

  return (
    <BackendHealthContext.Provider
      value={{ isBackendReachable, isOrkaiRunning, checkBackend, checkOrkai }}
    >
      {children}
    </BackendHealthContext.Provider>
  )
}

export default BackendHealthProvider
