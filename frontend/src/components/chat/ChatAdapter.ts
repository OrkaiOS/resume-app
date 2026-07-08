import type { ChatModelAdapter } from "@assistant-ui/react"

// @orkai:ref(id=2cf97580-172f-410d-81b4-edb7e177a7b3)
// @orkai:ref(id=61ce6f49-1307-4a1e-8ecf-9c49ce906520)
// @orkai:decision The system prompt is assembled server-side from orkai standards per FR-031. The adapter passes opportunityId so the backend can fetch the right context. Frontend no longer constructs system messages.

function getApiBase(): string {
  const base = import.meta.env.VITE_API_BASE
  if (!base) {
    throw new Error("VITE_API_BASE is required")
  }
  return base
}

const API_BASE = getApiBase()

interface ChatMessage {
  role: string
  content: string
}

export function createChatAdapter(opportunityId: string | null): ChatModelAdapter {
  return {
    async *run({ messages, abortSignal }) {
      const chatMessages: ChatMessage[] = messages.map((m) => ({
        role: m.role,
        content: m.content
          .filter((c): c is { type: "text"; text: string } => c.type === "text")
          .map((c) => c.text)
          .join("\n"),
      }))

      const body: Record<string, unknown> = { messages: chatMessages }
      if (opportunityId) {
        body.opportunityId = opportunityId
      }

      const response = await fetch(`${API_BASE}/chat`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
        signal: abortSignal,
      })

      if (!response.ok) {
        throw new Error(`Chat API error: ${response.status} ${response.statusText}`)
      }

      if (!response.body) {
        throw new Error("No response body")
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let fullText = ""
      let buffer = ""

      try {
        while (true) {
          const { done, value } = await reader.read()
          if (done) break

          buffer += decoder.decode(value, { stream: true })
          const lines = buffer.split("\n")
          buffer = lines.pop() || ""

          for (const line of lines) {
            if (!line.startsWith("data: ")) continue
            const json = line.slice(6)
            try {
              const event: {
                token?: string
                done?: boolean
                error?: string
                toolCall?: unknown
                toolResult?: unknown
              } = JSON.parse(json)
              if (event.error) {
                throw new Error(event.error)
              }
              // FR-032: toolCall and toolResult events are emitted by
              // the backend during the agentic tool-calling loop. They
              // don't produce visible text; the agent's final text
              // response describes what happened. A later slice will
              // render these as UI badges.
              if (event.toolCall || event.toolResult) {
                continue
              }
              if (event.token) {
                fullText += event.token
                yield { content: [{ type: "text" as const, text: fullText }] }
              }
              if (event.done) return
            } catch (e) {
              if (e instanceof SyntaxError) continue
              throw e
            }
          }
        }
      } catch (e) {
        if (e instanceof DOMException && e.name === "AbortError") {
          return
        }
        throw e
      }
    },
  }
}
