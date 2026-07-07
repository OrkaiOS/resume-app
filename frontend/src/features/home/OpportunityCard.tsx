import { useState } from "react"
import { FileText, Mail, MessageSquare, Trash2, Loader2 } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardHeader,
  CardTitle,
  CardDescription,
  CardContent,
  CardFooter,
} from "@/components/ui/card"
import { S } from "@/lib/strings"
import { formatDate } from "@/features/home/format-date"
import { useDeleteOpportunity } from "@/api/opportunities"
import {
  useResumeByOpportunity,
  useCoverLetterByOpportunity,
} from "@/api/documents"
import type { OpportunityResponse } from "@/types/api"

interface OpportunityCardProps {
  opportunity: OpportunityResponse
}

// @orkai:ref(id=2cf97580-172f-410d-81b4-edb7e177a7b3)
// @orkai:ref(id=6e959cda-9e4a-4c44-b87e-4c43deea936f)
// @orkai:ref(id=50cd15ac-f372-48ac-baa6-fdc20566c343)
// @orkai:decision "Open Agent" is a placeholder until FR-030 ships; shows toast instead of navigating to a non-existent Chat page
function OpportunityCard({ opportunity }: OpportunityCardProps) {
  const [confirmingDelete, setConfirmingDelete] = useState(false)
  const deleteOpportunity = useDeleteOpportunity()
  const resume = useResumeByOpportunity(opportunity.id)
  const coverLetter = useCoverLetterByOpportunity(opportunity.id)

  function handleOpenAgent() {
    toast.info(S.toast.chatComingSoon)
  }

  function handleDelete() {
    deleteOpportunity.mutate(opportunity.id, {
      onSuccess: (res) => {
        if (!res.error) {
          setConfirmingDelete(false)
        }
      },
    })
  }

  const hasResume = !resume.isPending && resume.data !== null
  const hasCoverLetter = !coverLetter.isPending && coverLetter.data !== null

  return (
    <Card>
      <CardHeader>
        <div className="flex items-start justify-between gap-2">
          <div className="min-w-0">
            <CardTitle className="text-base">{opportunity.company}</CardTitle>
            <CardDescription className="mt-1">{opportunity.role}</CardDescription>
          </div>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="size-8 shrink-0 text-muted-foreground hover:text-destructive"
            onClick={() => setConfirmingDelete(true)}
            aria-label={S.home.deleteOpportunity}
          >
            <Trash2 className="size-4" />
          </Button>
        </div>
      </CardHeader>
      <CardContent className="space-y-2">
        {opportunity.description && (
          <p className="whitespace-pre-wrap text-sm text-muted-foreground line-clamp-4">
            {opportunity.description}
          </p>
        )}
        <p className="text-xs text-muted-foreground">
          {formatDate(opportunity.createdAt) || "—"}
        </p>
      </CardContent>
      <CardFooter className="gap-1">
        {confirmingDelete ? (
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">
              {S.home.confirmDelete}
            </span>
            <Button
              type="button"
              size="sm"
              variant="destructive"
              disabled={deleteOpportunity.isPending}
              onClick={handleDelete}
            >
              {deleteOpportunity.isPending ? (
                <Loader2 className="size-3 animate-spin" />
              ) : (
                S.home.deleteConfirm
              )}
            </Button>
            <Button
              type="button"
              size="sm"
              variant="ghost"
              disabled={deleteOpportunity.isPending}
              onClick={() => setConfirmingDelete(false)}
            >
              {S.home.createCancel}
            </Button>
          </div>
        ) : (
          <>
            <div className="flex items-center gap-2 text-xs text-muted-foreground">
              <FileText
                className={hasResume ? "size-3.5" : "size-3.5 opacity-40"}
              />
              <span>{hasResume ? "1" : "0"}</span>
              <Mail
                className={
                  hasCoverLetter ? "size-3.5" : "size-3.5 opacity-40"
                }
              />
              <span>{hasCoverLetter ? "1" : "0"}</span>
            </div>
            <div className="ml-auto">
              <Button
                type="button"
                size="sm"
                variant="default"
                onClick={handleOpenAgent}
                className="gap-2"
              >
                <MessageSquare className="size-4" />
                {S.home.openAgent}
              </Button>
            </div>
          </>
        )}
      </CardFooter>
    </Card>
  )
}

export { OpportunityCard }
export type { OpportunityCardProps }
