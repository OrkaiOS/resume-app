import type { ReactNode } from "react"
import {
  AssistantRuntimeProvider,
  useLocalRuntime,
} from "@assistant-ui/react"
import { chatAdapter } from "@/components/chat/ChatAdapter"

interface ChatRuntimeProviderProps {
  children: ReactNode
  contextMessage?: string
}

export function ChatRuntimeProvider({
  children,
  contextMessage,
}: ChatRuntimeProviderProps) {
  const runtime = useLocalRuntime(chatAdapter, {
    initialMessages: contextMessage
      ? [
          {
            role: "system",
            content: [{ type: "text" as const, text: contextMessage }],
          },
        ]
      : undefined,
  })

  return (
    <AssistantRuntimeProvider runtime={runtime}>
      {children}
    </AssistantRuntimeProvider>
  )
}
