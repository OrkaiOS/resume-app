import { describe, it, expect } from "vitest"
import { useStyleStore } from "@/store/style-store"
import GlassCanvas, { GlassCanvas as NamedExport } from "@/components/ui/glass-canvas"

describe("GlassCanvas", () => {
  it("default export is a function", () => {
    expect(typeof GlassCanvas).toBe("function")
  })

  it("named export matches default export", () => {
    expect(NamedExport).toBe(GlassCanvas)
  })

  describe("visibility logic", () => {
    it("store reflects glass style after setStyle", () => {
      useStyleStore.getState().setStyle("glass")
      const { style } = useStyleStore.getState()
      expect(style).toBe("glass")
    })

    it("store reflects default style after setStyle", () => {
      useStyleStore.getState().setStyle("default")
      const { style } = useStyleStore.getState()
      expect(style).toBe("default")
    })

    it("style values are only 'default' or 'glass'", () => {
      const { style } = useStyleStore.getState()
      expect(["default", "glass"]).toContain(style)
    })
  })
})
