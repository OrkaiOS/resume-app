import { useState } from "react"
import { Loader2, Plus, AlertCircle } from "lucide-react"

import { Button } from "@/components/ui/button"
import { StyleSwitcher } from "@/components/ui/style-switcher"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { S } from "@/lib/strings"
import { useOpportunities } from "@/api/opportunities"
import { OpportunityCard } from "@/features/home/OpportunityCard"
import { EmptyState } from "@/features/home/EmptyState"
import { CreateOpportunityForm } from "@/features/home/CreateOpportunityForm"
import type { OpportunityResponse } from "@/types/api"

// @orkai:ref(id=2cf97580-172f-410d-81b4-edb7e177a7b3)
// @orkai:ref(id=96a4fb45-c107-4855-8f1a-cb701c958dac)
// @orkai:ref(id=6e959cda-9e4a-4c44-b87e-4c43deea936f)
// @orkai:ref(id=5759c69e-7fc5-4e0c-8e9e-554fd4388492)
// @orkai:decision Loading/error/empty/populated states per NFR-14; responsive grid 1 col mobile, 2 col md, 3 col lg per NFR-13 (1024-2560px). A single form instance serves create and edit: `editingOpportunity` is null → create mode, set → edit mode with that opportunity's values prefilled.
function HomePage() {
  const { data, isLoading, isError, refetch } = useOpportunities()
  const [showCreate, setShowCreate] = useState(false)
  const [editingOpportunity, setEditingOpportunity] = useState<OpportunityResponse | null>(null)

  if (isLoading) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-background">
        <div className="flex flex-col items-center gap-4">
          <Loader2 className="size-8 animate-spin text-primary" />
          <p className="text-muted-foreground">{S.home.loading}</p>
        </div>
      </div>
    )
  }

  if (isError) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-background px-4">
        <div className="mx-auto max-w-md text-center">
          <div className="mb-6 flex justify-center">
            <div className="flex size-14 items-center justify-center rounded-2xl bg-destructive/10">
              <AlertCircle className="size-7 text-destructive" />
            </div>
          </div>
          <h2 className="text-2xl font-semibold tracking-tight text-foreground">
            {S.home.errorTitle}
          </h2>
          <p className="mt-3 text-muted-foreground">{S.home.errorDescription}</p>
          <div className="mt-8">
            <Button onClick={() => refetch()} className="gap-2">
              <Loader2 className="size-4" />
              {S.home.errorRetry}
            </Button>
          </div>
        </div>
      </div>
    )
  }

  const items = data?.items ?? []
  const dialogOpen = showCreate || editingOpportunity !== null

  function closeDialog() {
    setShowCreate(false)
    setEditingOpportunity(null)
  }

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b bg-background">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-4 py-4">
          <div>
            <h1 className="text-xl font-semibold tracking-tight text-foreground">
              {S.home.title}
            </h1>
            <p className="text-sm text-muted-foreground">{S.home.subtitle}</p>
          </div>
          <div className="flex items-center gap-3">
            <StyleSwitcher />
            <Button
              type="button"
              onClick={() => {
                setEditingOpportunity(null)
                setShowCreate(true)
              }}
              className="gap-2"
            >
              <Plus className="size-4" />
              {S.home.newOpportunity}
            </Button>
          </div>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-4 py-8">
        <Dialog open={dialogOpen} onOpenChange={(open) => { if (!open) closeDialog() }}>
          <DialogContent className="sm:max-w-2xl">
            <DialogHeader>
              <DialogTitle>
                {editingOpportunity ? S.home.editOpportunity : S.home.newOpportunity}
              </DialogTitle>
            </DialogHeader>
            <CreateOpportunityForm
              key={editingOpportunity?.id ?? "new"}
              opportunity={editingOpportunity ?? undefined}
              onCancel={closeDialog}
              onSaved={closeDialog}
            />
          </DialogContent>
        </Dialog>

        {items.length === 0 ? (
          <EmptyState />
        ) : (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
            {items.map((opportunity) => (
              <OpportunityCard
                key={opportunity.id}
                opportunity={opportunity}
                onEdit={setEditingOpportunity}
              />
            ))}
          </div>
        )}
      </main>
    </div>
  )
}

export default HomePage