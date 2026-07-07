import { useState } from "react"
import { Loader2, Plus, AlertCircle } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { S } from "@/lib/strings"
import { useOpportunities } from "@/api/opportunities"
import { OpportunityCard } from "@/features/home/OpportunityCard"
import { EmptyState } from "@/features/home/EmptyState"
import { CreateOpportunityForm } from "@/features/home/CreateOpportunityForm"

// @orkai:ref(id=2cf97580-172f-410d-81b4-edb7e177a7b3)
// @orkai:ref(id=96a4fb45)
// @orkai:ref(id=6e959cda)
// @orkai:ref(id=5759c69e)
// @orkai:decision Loading/error/empty/populated states per NFR-14; responsive grid 1 col mobile, 2 col md, 3 col lg per NFR-13 (1024-2560px)
function HomePage() {
  const { data, isLoading, isError, refetch } = useOpportunities()
  const [showCreate, setShowCreate] = useState(false)

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
          <Button
            type="button"
            onClick={() => setShowCreate((v) => !v)}
            className="gap-2"
          >
            <Plus className="size-4" />
            {S.home.newOpportunity}
          </Button>
        </div>
      </header>

      <main className="mx-auto max-w-6xl px-4 py-8">
        {showCreate && (
          <Card className="mb-6">
            <CardContent className="pt-4">
              <CreateOpportunityForm
                onCancel={() => setShowCreate(false)}
                onCreated={() => setShowCreate(false)}
              />
            </CardContent>
          </Card>
        )}

        {items.length === 0 && !showCreate ? (
          <EmptyState />
        ) : (
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
            {items.map((opportunity) => (
              <OpportunityCard
                key={opportunity.id}
                opportunity={opportunity}
              />
            ))}
          </div>
        )}
      </main>
    </div>
  )
}

export default HomePage