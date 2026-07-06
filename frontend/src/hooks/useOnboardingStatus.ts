import { useOnboardingStatus as useApiOnboardingStatus } from "@/api/onboarding"

interface UseOnboardingStatusResult {
  isOnboarded: boolean | null
  isLoading: boolean
}

function useOnboardingStatus(): UseOnboardingStatusResult {
  const { data, isLoading } = useApiOnboardingStatus()

  return {
    isOnboarded: isLoading ? null : (data?.onboarded ?? false),
    isLoading,
  }
}

export default useOnboardingStatus
export type { UseOnboardingStatusResult }
