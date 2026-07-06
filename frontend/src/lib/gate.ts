export type AppGate = "orkai-loading" | "orkai-down" | "onboarding-loading" | "onboarding" | "app"

export function resolveAppGate(isOrkaiRunning: boolean | null, isOnboarded: boolean | null): AppGate {
  if (isOrkaiRunning === null) return "orkai-loading"
  if (!isOrkaiRunning) return "orkai-down"
  if (isOnboarded === null) return "onboarding-loading"
  if (!isOnboarded) return "onboarding"
  return "app"
}
