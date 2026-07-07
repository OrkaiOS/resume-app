import { TooltipProvider } from "@/components/ui/tooltip"
import { Thread } from "@/components/assistant-ui/thread"
import { ChatRuntimeProvider } from "@/components/chat/ChatRuntimeProvider"
import { useOpportunity } from "@/api/opportunities"
import { Building2, Loader2, Sparkles, ArrowLeft } from "lucide-react"

interface ChatPageProps {}

function ChatPage(_props: ChatPageProps) {
  const params = new URLSearchParams(window.location.search)
  const hasChat = params.has("chat")
  const opportunityId = params.get("chat") || null
  const { data: opportunity, isLoading } = useOpportunity(opportunityId)

  if (!hasChat) {
    return (
      <div className="flex min-h-[60vh] flex-col items-center justify-center px-4 text-center">
        <div className="mb-6 flex justify-center">
          <div className="flex size-14 items-center justify-center rounded-2xl bg-muted">
            <Sparkles className="size-7 text-muted-foreground" />
          </div>
        </div>
        <h2 className="text-2xl font-semibold tracking-tight text-foreground">
          Chat Agent
        </h2>
        <p className="mt-3 max-w-md text-muted-foreground">
          Your AI assistant lives here. Open a chat from any opportunity card to get tailored help with your resume and cover letter.
        </p>
        <div className="mt-8">
          <a href="/" className="inline-flex items-center gap-2 rounded-lg border border-border bg-background px-6 py-3 text-sm font-medium text-foreground hover:bg-muted transition-colors">
            <ArrowLeft className="size-4" />
            Back to Opportunities
          </a>
        </div>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="flex h-screen flex-col items-center justify-center gap-4">
        <Loader2 className="size-8 animate-spin text-muted-foreground" />
        <p className="text-sm text-muted-foreground">Loading opportunity details...</p>
      </div>
    )
  }

  return (
    <div className="flex h-screen flex-col bg-background">
      <header className="flex items-center gap-3 border-b px-6 py-3">
        <Building2 className="size-5 text-muted-foreground" />
        <div>
          <h1 className="text-sm font-semibold">
            {opportunity ? `${opportunity.company} — ${opportunity.role}` : "New Conversation"}
          </h1>
          <p className="text-xs text-muted-foreground">Chat Agent</p>
        </div>
      </header>
      <div className="flex-1 overflow-hidden">
        <TooltipProvider>
          <ChatRuntimeProvider opportunityId={opportunityId}>
            <Thread />
          </ChatRuntimeProvider>
        </TooltipProvider>
      </div>
    </div>
  )
}

export default ChatPage
