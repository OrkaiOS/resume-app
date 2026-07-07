import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { apiGet, apiPost, apiUpload } from "@/api/client"
import { S } from "@/lib/strings"
import type {
  LLMConfigRequest,
  LLMConfigResponse,
  OrkaiSetupResponse,
  OrkaiSetupStatus,
  OnboardingStatus,
  ProfileResponse,
  ProfileUpsertRequest,
} from "@/types/api"

const onboardingKeys = {
  all: ["onboarding"] as const,
  status: () => [...onboardingKeys.all, "status"] as const,
  onboardingSetup: (setupId: string) => [...onboardingKeys.all, "onboarding-final-step", setupId] as const,
}

export function useOnboardingStatus() {
  return useQuery({
    queryKey: onboardingKeys.status(),
    queryFn: () => apiGet<OnboardingStatus>("/onboarding/status"),
    select: (res) => res.data,
  })
}

export function useSaveLLMConfig() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: LLMConfigRequest) =>
      apiPost<LLMConfigResponse>("/onboarding/llm-config", input),
    onSuccess: (res) => {
      if (res.error) {
        toast.error(S.toast.llmConfigSaveFailed, {
          description: res.error.message,
        })
        return
      }
      toast.success(S.toast.llmConfigSaved, {
        description: res.data?.validated ? S.toast.llmConnectionValidated : undefined,
      })
      queryClient.invalidateQueries({ queryKey: onboardingKeys.status() })
    },
    onError: (e: Error) => {
      toast.error("Failed to save LLM config", {
        description: e.message,
      })
    },
  })
}

export function useSaveProfile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: ProfileUpsertRequest) =>
      apiPost<ProfileResponse>("/onboarding/profile", input),
    onSuccess: (res) => {
      if (res.error) {
        toast.error(S.toast.profileSaveFailed, {
          description: res.error.message,
        })
        return
      }
      toast.success(S.toast.profileSaved)
      queryClient.invalidateQueries({ queryKey: onboardingKeys.status() })
    },
    onError: (e: Error) => {
      toast.error("Failed to save profile", {
        description: e.message,
      })
    },
  })
}

export function useUploadProfile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (file: File) =>
      apiUpload<ProfileResponse>("/profile/upload", file),
    onSuccess: (res) => {
      if (res.error) {
        toast.error(S.toast.profileUploadFailed, {
          description: res.error.message,
        })
        return
      }
      toast.success(S.toast.profileUploaded)
      queryClient.invalidateQueries({ queryKey: onboardingKeys.status() })
    },
    onError: (e: Error) => {
      toast.error("Failed to upload profile", {
        description: e.message,
      })
    },
  })
}

export function useTriggerSetup() {
  return useMutation({
    mutationFn: () =>
      apiPost<OrkaiSetupResponse>("/onboarding/orkai-setup", {}),
    onError: (e: Error) => {
      toast.error(S.toast.setupTriggerFailed, {
        description: e.message,
      })
    },
  })
}

export function useSetupStatus(setupId: string, enabled: boolean) {
  return useQuery({
    queryKey: onboardingKeys.onboardingSetup(setupId),
    queryFn: () =>
      apiGet<OrkaiSetupStatus>(`/onboarding/orkai-setup/status?sessionId=${setupId}`),
    select: (res) => res.data,
    enabled: !!setupId && enabled,
    refetchInterval: (query) => {
      const res = query.state.data
      if (!res || res.error) return 1000
      return res.data.completed ? false : 1000
    },
  })
}
