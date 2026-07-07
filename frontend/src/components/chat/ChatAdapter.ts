import type { ChatModelAdapter } from "@assistant-ui/react"

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

export const chatAdapter: ChatModelAdapter = {
  async *run({ messages, abortSignal }) {
    const chatMessages: ChatMessage[] = messages.map((m) => ({
      role: m.role,
      content: m.content
        .filter((c): c is { type: "text"; text: string } => c.type === "text")
        .map((c) => c.text)
        .join("\n"),
    }))

    const response = await fetch(`${API_BASE}/v1/api/chat`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ messages: chatMessages }),
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
            const event: { token?: string; done?: boolean; error?: string } =
              JSON.parse(json)
            if (event.error) {
              throw new Error(event.error)
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
