import { MessageSquare } from "lucide-react"
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
import type { OpportunityResponse } from "@/types/api"

interface OpportunityCardProps {
  opportunity: OpportunityResponse
}

// @orkai:ref(id=2cf97580-172f-410d-81b4-edb7e177a7b3)
// @orkai:ref(id=6e959cda)
// @orkai:ref(id=50cd15ac)
// @orkai:decision "Open Agent" is a placeholder until FR-030 ships; shows toast instead of navigating to a non-existent Chat page
function OpportunityCard({ opportunity }: OpportunityCardProps) {
  function handleOpenAgent() {
    toast.info(S.toast.chatComingSoon)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{opportunity.company}</CardTitle>
        <CardDescription>{opportunity.role}</CardDescription>
      </CardHeader>
      <CardContent>
        <p className="text-xs text-muted-foreground">
          {formatDate(opportunity.createdAt) || "—"}
        </p>
      </CardContent>
      <CardFooter>
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
      </CardFooter>
    </Card>
  )
}

export { OpportunityCard }
export type { OpportunityCardProps }