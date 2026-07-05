import { createContext, useContext } from "react"

interface OrkaiHealthContextValue {
  isOrkaiRunning: boolean | null
  checkHealth: () => Promise<boolean>
}

const OrkaiHealthContext = createContext<OrkaiHealthContextValue | null>(null)

function useOrkaiHealth(): OrkaiHealthContextValue {
  const ctx = useContext(OrkaiHealthContext)
  if (!ctx) {
    throw new Error("useOrkaiHealth must be used within OrkaiHealthProvider")
  }
  return ctx
}

export { OrkaiHealthContext, useOrkaiHealth }
export type { OrkaiHealthContextValue }