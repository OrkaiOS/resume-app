import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query"
import { toast } from "sonner"
import { apiGet, apiPost, apiPut, apiDelete } from "@/api/client"
import { S } from "@/lib/strings"
import type {
  PaginatedResponse,
  OpportunityResponse,
  OpportunityCreateRequest,
  OpportunityUpdateRequest,
  OpportunityArchiveRequest,
} from "@/types/api"

// @orkai:ref(id=2cf97580-172f-410d-81b4-edb7e177a7b3)
// @orkai:ref(id=61ce6f49-1307-4a1e-8ecf-9c49ce906520)
// @orkai:decision Array query-key convention with list/detail split; invalidate list on every mutate to avoid stale cards (NFR-04)
export const opportunitiesKeys = {
  all: ["opportunities"] as const,
  list: () => [...opportunitiesKeys.all, "list"] as const,
  detail: (id: string) => [...opportunitiesKeys.all, "detail", id] as const,
}

export function useOpportunities() {
  return useQuery({
    queryKey: opportunitiesKeys.list(),
    queryFn: () =>
      apiGet<PaginatedResponse<OpportunityResponse>>("/opportunities"),
    select: (res) => res.data,
  })
}

export function useOpportunity(id: string | null) {
  return useQuery({
    queryKey: id ? opportunitiesKeys.detail(id) : ["opportunities", "detail", "__none__"],
    queryFn: () => apiGet<OpportunityResponse>(`/opportunities/${id}`),
    select: (res) => res.data,
    enabled: !!id,
  })
}

export function useCreateOpportunity() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: OpportunityCreateRequest) =>
      apiPost<OpportunityResponse>("/opportunities", input),
    onSuccess: (res) => {
      if (res.error) {
        toast.error(S.toast.opportunityCreateFailed, {
          description: res.error.message,
        })
        return
      }
      toast.success(S.toast.opportunityCreated)
      queryClient.invalidateQueries({ queryKey: opportunitiesKeys.list() })
    },
    onError: (e: Error) => {
      toast.error(S.toast.opportunityCreateFailed, { description: e.message })
    },
  })
}

export function useUpdateOpportunity() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: { id: string; body: OpportunityUpdateRequest }) =>
      apiPut<OpportunityResponse>(`/opportunities/${input.id}`, input.body),
    onSuccess: (res) => {
      if (res.error) {
        toast.error("Failed to update opportunity", {
          description: res.error.message,
        })
        return
      }
      queryClient.invalidateQueries({ queryKey: opportunitiesKeys.list() })
    },
    onError: (e: Error) => {
      toast.error("Failed to update opportunity", { description: e.message })
    },
  })
}

export function useDeleteOpportunity() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: string) =>
      apiDelete<null>(`/opportunities/${id}`),
    onSuccess: (res) => {
      if (res.error) {
        toast.error("Failed to delete opportunity", {
          description: res.error.message,
        })
        return
      }
      toast.success(S.toast.opportunityDeleted)
      queryClient.invalidateQueries({ queryKey: opportunitiesKeys.list() })
    },
    onError: (e: Error) => {
      toast.error("Failed to delete opportunity", { description: e.message })
    },
  })
}

export function useArchiveOpportunity() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (input: { id: string; body: OpportunityArchiveRequest }) =>
      apiPut<OpportunityResponse>(`/opportunities/${input.id}/archive`, input.body),
    onSuccess: (res) => {
      if (res.error) {
        toast.error("Failed to archive opportunity", {
          description: res.error.message,
        })
        return
      }
      queryClient.invalidateQueries({ queryKey: opportunitiesKeys.list() })
    },
    onError: (e: Error) => {
      toast.error("Failed to archive opportunity", { description: e.message })
    },
  })
}