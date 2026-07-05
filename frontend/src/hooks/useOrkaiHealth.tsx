import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react"

const POLL_INTERVAL_MS = 5000
const HEALTH_URL = "/v1/api/health/orkai"

interface OrkaiHealthContextValue {
  isOrkaiRunning: boolean | null
  checkHealth: () => Promise<boolean>
}

const OrkaiHealthContext = createContext<OrkaiHealthContextValue | null>(null)

function OrkaiHealthProvider({ children }: { children: React.ReactNode }) {
  const [isOrkaiRunning, setIsOrkaiRunning] = useState<boolean | null>(null)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const checkHealth = useCallback(async () => {
    try {
      const res = await fetch(HEALTH_URL, { cache: "no-store" })
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
    const perform = async () => {
      const running = await checkHealth()
      return running
    }

    perform().then((running) => {
      if (!running) {
        intervalRef.current = setInterval(async () => {
          const stillRunning = await checkHealth()
          if (stillRunning) {
            if (intervalRef.current) {
              clearInterval(intervalRef.current)
              intervalRef.current = null
            }
          }
        }, POLL_INTERVAL_MS)
      }
    })

    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
        intervalRef.current = null
      }
    }
  }, [checkHealth])

  useEffect(() => {
    const onVisibilityChange = () => {
      if (document.visibilityState === "visible") {
        checkHealth()
      }
    }

    const onPopState = () => {
      checkHealth()
    }

    document.addEventListener("visibilitychange", onVisibilityChange)
    window.addEventListener("popstate", onPopState)

    return () => {
      document.removeEventListener("visibilitychange", onVisibilityChange)
      window.removeEventListener("popstate", onPopState)
    }
  }, [checkHealth])

  return (
    <OrkaiHealthContext.Provider value={{ isOrkaiRunning, checkHealth }}>
      {children}
    </OrkaiHealthContext.Provider>
  )
}

function useOrkaiHealth(): OrkaiHealthContextValue {
  const ctx = useContext(OrkaiHealthContext)
  if (!ctx) {
    throw new Error("useOrkaiHealth must be used within OrkaiHealthProvider")
  }
  return ctx
}

export { OrkaiHealthProvider, useOrkaiHealth }
export type { OrkaiHealthContextValue }