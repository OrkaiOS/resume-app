import { createContext, useContext } from "react"

interface BackendHealthContextValue {
  isBackendReachable: boolean | null
  isOrkaiRunning: boolean | null
  checkBackend: () => Promise<boolean>
  checkOrkai: () => Promise<boolean>
}

const BackendHealthContext = createContext<BackendHealthContextValue | null>(null)

function useBackendHealth(): BackendHealthContextValue {
  const ctx = useContext(BackendHealthContext)
  if (!ctx) {
    throw new Error("useBackendHealth must be used within BackendHealthProvider")
  }
  return ctx
}

export { BackendHealthContext, useBackendHealth }
export type { BackendHealthContextValue }
