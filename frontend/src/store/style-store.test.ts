import { describe, it, expect } from "vitest"
import { useStyleStore } from "@/store/style-store"

describe("useStyleStore", () => {
  it("initializes with default style", () => {
    const state = useStyleStore.getState()
    expect(state.style).toBe("default")
  })

  it("setStyle updates the style to glass", () => {
    useStyleStore.getState().setStyle("glass")
    const state = useStyleStore.getState()
    expect(state.style).toBe("glass")
  })

  it("setStyle updates the style back to default", () => {
    useStyleStore.getState().setStyle("default")
    const state = useStyleStore.getState()
    expect(state.style).toBe("default")
  })

  it("selector returns only the style field", () => {
    const style = useStyleStore.getState().style
    expect(typeof style).toBe("string")
    expect(["default", "glass"]).toContain(style)
  })
})
