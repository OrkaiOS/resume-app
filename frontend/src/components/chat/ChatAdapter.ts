import type { ChatModelAdapter } from "@assistant-ui/react"
import type { ReadonlyJSONObject } from "assistant-stream/utils"

// @orkai:ref(id=2cf97580-172f-410d-81b4-edb7e177a7b3)
// @orkai:ref(id=61ce6f49-1307-4a1e-8ecf-9c49ce906520)
// @orkai:decision The system prompt is assembled server-side from orkai standards per FR-031. The adapter passes opportunityId so the backend can fetch the right context. Frontend no longer constructs system messages.
// @orkai:decision Reasoning tokens are display-only and yielded as reasoning parts alongside text parts. Tool-call and tool-result events are surfaced as inline badges. The assistant-ui thread.tsx already renders ReasoningGroup and ToolGroup from GroupedParts — no frontend component changes needed beyond this adapter.

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

export interface ToolCallPart {
  type: "tool-call"
  toolCallId: string
  toolName: string
  argsText: string
  args: ReadonlyJSONObject
  result?: string
}

export type ContentPart =
  | { type: "text"; text: string }
  | { type: "reasoning"; text: string }
  | ToolCallPart

export interface ChatAdapterState {
  fullText: string
  fullReasoning: string
  toolCallParts: ToolCallPart[]
}

export interface ChatStreamEvent {
  token?: string
  reasoning?: string
  done?: boolean
  error?: string
  toolCall?: { id: string; name: string; args: string }
  toolResult?: { id: string; output: string; error?: string }
}

function buildContent(state: ChatAdapterState): ContentPart[] {
  const parts: ContentPart[] = []
  if (state.fullReasoning) {
    parts.push({ type: "reasoning", text: state.fullReasoning })
  }
  if (state.fullText) {
    parts.push({ type: "text", text: state.fullText })
  }
  parts.push(...state.toolCallParts)
  return parts
}

export function processChatEvent(state: ChatAdapterState, event: ChatStreamEvent): ContentPart[] | null {
  if (event.reasoning) {
    state.fullReasoning += event.reasoning
    return buildContent(state)
  }
  if (event.token) {
    state.fullText += event.token
    return buildContent(state)
  }
  if (event.toolCall) {
    const tc = event.toolCall
    let args: ReadonlyJSONObject
    try {
      args = JSON.parse(tc.args)
    } catch {
      args = {}
    }
    state.toolCallParts.push({
      type: "tool-call",
      toolCallId: tc.id,
      toolName: tc.name,
      argsText: tc.args,
      args,
    })
    return buildContent(state)
  }
  if (event.toolResult) {
    const tr = event.toolResult
    for (const part of state.toolCallParts) {
      if (part.toolCallId === tr.id) {
        if (tr.error) {
          part.result = `Error: ${tr.error}`
        } else {
          part.result = tr.output
        }
      }
    }
    return buildContent(state)
  }
  return null
}

export function createChatAdapter(opportunityId: string | null): ChatModelAdapter {
  return {
    run: async function* ({ messages, abortSignal }) {
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
      const state: ChatAdapterState = {
        fullText: "",
        fullReasoning: "",
        toolCallParts: [],
      }
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
              const event: ChatStreamEvent = JSON.parse(json)
              if (event.error) {
                throw new Error(event.error)
              }
              if (event.done) return
              const content = processChatEvent(state, event)
              if (content) {
                yield { content }
              }
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
