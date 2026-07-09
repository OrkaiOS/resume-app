import { describe, it, expect } from "vitest"
import { processChatEvent, toChatMessages, type ChatAdapterState } from "./ChatAdapter"

function freshState(): ChatAdapterState {
  return { fullText: "", fullReasoning: "", toolCallParts: [] }
}

describe("processChatEvent", () => {
  it("yields a reasoning part when a reasoning event arrives", () => {
    const state = freshState()
    const content = processChatEvent(state, { reasoning: "Let me think" })
    expect(content).not.toBeNull()
    expect(content!).toHaveLength(1)
    expect(content![0]).toEqual({ type: "reasoning", text: "Let me think" })
  })

  it("accumulates reasoning across multiple events", () => {
    const state = freshState()
    processChatEvent(state, { reasoning: "Step 1. " })
    const content = processChatEvent(state, { reasoning: "Step 2." })
    expect(content).not.toBeNull()
    const reasoningPart = content!.find((p) => p.type === "reasoning")
    expect(reasoningPart).toEqual({ type: "reasoning", text: "Step 1. Step 2." })
  })

  it("yields a text part when a token event arrives", () => {
    const state = freshState()
    const content = processChatEvent(state, { token: "Hello" })
    expect(content).not.toBeNull()
    expect(content!).toHaveLength(1)
    expect(content![0]).toEqual({ type: "text", text: "Hello" })
  })

  it("accumulates text across multiple token events", () => {
    const state = freshState()
    processChatEvent(state, { token: "Hel" })
    const content = processChatEvent(state, { token: "lo" })
    expect(content).not.toBeNull()
    const textPart = content!.find((p) => p.type === "text")
    expect(textPart).toEqual({ type: "text", text: "Hello" })
  })

  it("yields reasoning and text parts together in correct order (reasoning first)", () => {
    const state = freshState()
    processChatEvent(state, { reasoning: "Let me think" })
    const content = processChatEvent(state, { token: "Here is the answer" })
    expect(content).not.toBeNull()
    expect(content!).toHaveLength(2)
    expect(content![0]).toEqual({ type: "reasoning", text: "Let me think" })
    expect(content![1]).toEqual({ type: "text", text: "Here is the answer" })
  })

  it("does not include an empty reasoning part when no reasoning was emitted", () => {
    const state = freshState()
    const content = processChatEvent(state, { token: "hello" })
    expect(content).not.toBeNull()
    expect(content!).toHaveLength(1)
    expect(content![0].type).toBe("text")
  })

  it("does not include an empty text part when no text was emitted", () => {
    const state = freshState()
    const content = processChatEvent(state, { reasoning: "thinking" })
    expect(content).not.toBeNull()
    expect(content!).toHaveLength(1)
    expect(content![0].type).toBe("reasoning")
  })

  it("yields a tool-call part with correct fields", () => {
    const state = freshState()
    const content = processChatEvent(state, {
      toolCall: { id: "call_1", name: "shell", args: '{"cmd":"ls"}' },
    })
    expect(content).not.toBeNull()
    const tcPart = content!.find((p) => p.type === "tool-call")
    expect(tcPart).toBeDefined()
    expect(tcPart).toMatchObject({
      type: "tool-call",
      toolCallId: "call_1",
      toolName: "shell",
      argsText: '{"cmd":"ls"}',
      args: { cmd: "ls" },
    })
    expect(state.toolCallParts).toHaveLength(1)
  })

  it("handles tool-call args parse failure gracefully", () => {
    const state = freshState()
    const content = processChatEvent(state, {
      toolCall: { id: "call_1", name: "bad", args: "not json" },
    })
    expect(content).not.toBeNull()
    const tcPart = content!.find((p) => p.type === "tool-call")
    expect(tcPart?.args).toEqual({})
  })

  it("updates matching tool-call part result on tool-result event", () => {
    const state = freshState()
    processChatEvent(state, {
      toolCall: { id: "call_1", name: "shell", args: "{}" },
    })
    const content = processChatEvent(state, {
      toolResult: { id: "call_1", output: "file1.txt\nfile2.txt" },
    })
    expect(content).not.toBeNull()
    const tcPart = content!.find((p) => p.type === "tool-call") as { result?: string } | undefined
    expect(tcPart?.result).toBe("file1.txt\nfile2.txt")
  })

  it("sets error result on tool-result event with error", () => {
    const state = freshState()
    processChatEvent(state, {
      toolCall: { id: "call_1", name: "shell", args: "{}" },
    })
    const content = processChatEvent(state, {
      toolResult: { id: "call_1", output: "", error: "command not found" },
    })
    expect(content).not.toBeNull()
    const tcPart = content!.find((p) => p.type === "tool-call") as { result?: string } | undefined
    expect(tcPart?.result).toBe("Error: command not found")
  })

  it("does not crash when tool-result references unknown call ID", () => {
    const state = freshState()
    const content = processChatEvent(state, {
      toolResult: { id: "call_unknown", output: "ok" },
    })
    expect(content).not.toBeNull()
    expect(content!).toHaveLength(0)
  })

  it("includes tool-call parts alongside reasoning and text", () => {
    const state = freshState()
    processChatEvent(state, { reasoning: "thinking" })
    processChatEvent(state, { token: "answer" })
    const content = processChatEvent(state, {
      toolCall: { id: "c1", name: "search", args: "{}" },
    })
    expect(content).not.toBeNull()
    expect(content!).toHaveLength(3)
    expect(content![0].type).toBe("reasoning")
    expect(content![1].type).toBe("text")
    expect(content![2].type).toBe("tool-call")
  })
})

describe("toChatMessages", () => {
  it("extracts text from text parts and joins with newline", () => {
    const out = toChatMessages([
      {
        role: "user",
        content: [
          { type: "text", text: "Hello" },
          { type: "text", text: "World" },
        ],
      },
    ])
    expect(out).toEqual([{ role: "user", content: "Hello\nWorld" }])
  })

  it("drops reasoning-only parts — display-only, never sent to model", () => {
    const out = toChatMessages([
      {
        role: "assistant",
        content: [{ type: "reasoning", text: "Let me think about this" }],
      },
    ])
    expect(out).toEqual([])
  })

  it("drops an interrupted assistant turn that produced only reasoning (no text)", () => {
    const out = toChatMessages([
      { role: "user", content: [{ type: "text", text: "What about this opportunity?" }] },
      {
        role: "assistant",
        content: [{ type: "reasoning", text: "I should look up..." }],
      },
      { role: "user", content: [{ type: "text", text: "Sorry, a quick detail please." }] },
    ])
    expect(out).toEqual([
      { role: "user", content: "What about this opportunity?" },
      { role: "user", content: "Sorry, a quick detail please." },
    ])
  })

  it("keeps an assistant turn that produced text alongside reasoning", () => {
    const out = toChatMessages([
      {
        role: "assistant",
        content: [
          { type: "reasoning", text: "thinking" },
          { type: "text", text: "answer" },
        ],
      },
    ])
    expect(out).toEqual([{ role: "assistant", content: "answer" }])
  })

  it("drops tool-call-only assistant turns (no text produced)", () => {
    const out = toChatMessages([
      {
        role: "assistant",
        content: [
          { type: "tool-call", toolCallId: "c1", toolName: "shell", argsText: "{}" },
        ],
      },
    ])
    expect(out).toEqual([])
  })
})
