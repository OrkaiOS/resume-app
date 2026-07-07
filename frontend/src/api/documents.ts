import { useQuery } from "@tanstack/react-query"
import { apiGet } from "@/api/client"
import type { ResumeResponse, CoverLetterResponse } from "@/types/api"

export const documentsKeys = {
  all: ["documents"] as const,
  resume: (opportunityId: string) =>
    [...documentsKeys.all, "resume", opportunityId] as const,
  coverLetter: (opportunityId: string) =>
    [...documentsKeys.all, "cover-letter", opportunityId] as const,
}

export function useResumeByOpportunity(opportunityId: string | null) {
  return useQuery({
    queryKey: opportunityId
      ? documentsKeys.resume(opportunityId)
      : ["documents", "resume", "__none__"],
    queryFn: () =>
      apiGet<ResumeResponse>(`/opportunities/${opportunityId}/resume`),
    select: (res) => (res.error ? null : res.data),
    enabled: !!opportunityId,
  })
}

export function useCoverLetterByOpportunity(opportunityId: string | null) {
  return useQuery({
    queryKey: opportunityId
      ? documentsKeys.coverLetter(opportunityId)
      : ["documents", "cover-letter", "__none__"],
    queryFn: () =>
      apiGet<CoverLetterResponse>(
        `/opportunities/${opportunityId}/cover-letter`
      ),
    select: (res) => (res.error ? null : res.data),
    enabled: !!opportunityId,
  })
}
