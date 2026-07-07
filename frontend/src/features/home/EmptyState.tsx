import { FileText, Sparkles } from "lucide-react"
import { toast } from "sonner"

import { Button } from "@/components/ui/button"
import { S } from "@/lib/strings"

// @orkai:ref(id=2cf97580-172f-410d-81b4-edb7e177a7b3)
// @orkai:ref(id=6e959cda)
// @orkai:ref(id=5759c69e)
// @orkai:decision "Start Chat" shows a toast until FR-030 (Chat) ships; visual AC of FR-021 (prominent button) is met, navigation AC is deferred to the chat slice
function EmptyState() {
  function handleStartChat() {
    toast.info(S.toast.chatComingSoon)
  }

  return (
    <div className="flex min-h-[60vh] flex-col items-center justify-center px-4 text-center">
      <div className="mb-6 flex justify-center">
        <div className="flex size-14 items-center justify-center rounded-2xl bg-primary/10">
          <FileText className="size-7 text-primary" />
        </div>
      </div>
      <h2 className="text-2xl font-semibold tracking-tight text-foreground">
        {S.home.emptyTitle}
      </h2>
      <p className="mt-3 max-w-md text-muted-foreground">
        {S.home.emptyDescription}
      </p>
      <div className="mt-8">
        <Button size="lg" onClick={handleStartChat} className="gap-2">
          <Sparkles className="size-4" />
          {S.home.emptyAction}
        </Button>
      </div>
    </div>
  )
}

export { EmptyState }