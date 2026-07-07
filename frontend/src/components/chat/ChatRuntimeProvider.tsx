import type { ReactNode } from "react"
import {
  AssistantRuntimeProvider,
  useLocalRuntime,
} from "@assistant-ui/react"
import { createChatAdapter } from "@/components/chat/ChatAdapter"

interface ChatRuntimeProviderProps {
  children: ReactNode
  opportunityId: string | null
}

export function ChatRuntimeProvider({
  children,
  opportunityId,
}: ChatRuntimeProviderProps) {
  const adapter = createChatAdapter(opportunityId)
  const runtime = useLocalRuntime(adapter)

  return (
    <AssistantRuntimeProvider runtime={runtime}>
      {children}
    </AssistantRuntimeProvider>
  )
}
