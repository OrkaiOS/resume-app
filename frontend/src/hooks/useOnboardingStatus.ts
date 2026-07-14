import { useOnboardingStatus as useApiOnboardingStatus } from "@/api/onboarding"

interface UseOnboardingStatusResult {
  isOnboarded: boolean | null
  isLoading: boolean
  isError: boolean
}

function useOnboardingStatus(): UseOnboardingStatusResult {
  const { data, isLoading, isError } = useApiOnboardingStatus()

  return {
    isOnboarded: isLoading || isError ? null : (data?.onboarded ?? false),
    isLoading,
    isError,
  }
}

export default useOnboardingStatus
export type { UseOnboardingStatusResult }
