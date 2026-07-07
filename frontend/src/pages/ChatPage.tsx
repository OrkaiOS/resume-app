import { TooltipProvider } from "@/components/ui/tooltip"
import { Thread } from "@/components/assistant-ui/thread"
import { ChatRuntimeProvider } from "@/components/chat/ChatRuntimeProvider"
import { useOpportunity } from "@/api/opportunities"
import { Building2 } from "lucide-react"

interface ChatPageProps {}

function ChatPage(_props: ChatPageProps) {
  const params = new URLSearchParams(window.location.search)
  const opportunityId = params.get("chat")
  const { data: opportunity } = useOpportunity(opportunityId)

  if (!opportunityId) {
    window.location.search = ""
    return null
  }

  const contextMessage = opportunity
    ? `You are helping the user create a tailored resume and cover letter for a job opportunity at ${opportunity.company} for the role of ${opportunity.role}. The opportunity was created on ${new Date(opportunity.createdAt).toLocaleDateString()}.`
    : undefined

  return (
    <div className="flex h-screen flex-col bg-background">
      <header className="flex items-center gap-3 border-b px-6 py-3">
        <Building2 className="size-5 text-muted-foreground" />
        <div>
          <h1 className="text-sm font-semibold">
            {opportunity ? `${opportunity.company} — ${opportunity.role}` : "Loading..."}
          </h1>
          <p className="text-xs text-muted-foreground">Chat Agent</p>
        </div>
      </header>
      <div className="flex-1 overflow-hidden">
        <TooltipProvider>
          <ChatRuntimeProvider contextMessage={contextMessage}>
            <Thread />
          </ChatRuntimeProvider>
        </TooltipProvider>
      </div>
    </div>
  )
}

export default ChatPage
