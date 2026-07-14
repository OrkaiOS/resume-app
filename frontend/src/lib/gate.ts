export type AppGate =
  | "backend-loading"
  | "backend-down"
  | "orkai-loading"
  | "orkai-down"
  | "onboarding-loading"
  | "onboarding"
  | "app"

export function resolveAppGate(
  isBackendReachable: boolean | null,
  isOrkaiRunning: boolean | null,
  isOnboarded: boolean | null
): AppGate {
  if (isBackendReachable === null) return "backend-loading"
  if (!isBackendReachable) return "backend-down"
  if (isOrkaiRunning === null) return "orkai-loading"
  if (!isOrkaiRunning) return "orkai-down"
  if (isOnboarded === null) return "onboarding-loading"
  if (!isOnboarded) return "onboarding"
  return "app"
}
