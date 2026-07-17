import { describe, it, expect } from "vitest"
import GlassCanvas, { GlassCanvas as NamedExport } from "@/components/ui/glass-canvas"

describe("GlassCanvas", () => {
  it("default export is a function", () => {
    expect(typeof GlassCanvas).toBe("function")
  })

  it("named export matches default export", () => {
    expect(NamedExport).toBe(GlassCanvas)
  })

  it("has a display name", () => {
    expect(GlassCanvas.name).toBe("GlassCanvas")
  })

  it("always renders a container div", () => {
    const result = GlassCanvas({})
    expect(result).toBeDefined()
    expect(result.type).toBe("div")
  })

  it("container has aria-hidden for accessibility", () => {
    const result = GlassCanvas({})
    expect(result.props).toBeDefined()
    expect(result.props["aria-hidden"]).toBe("true")
  })

  it("container has hidden class and glass-canvas-visible visibility class", () => {
    const result = GlassCanvas({})
    const className = result.props.className
    expect(className).toContain("hidden")
    expect(className).toContain("glass-canvas-visible")
  })

  it("contains three gradient blob children", () => {
    const result = GlassCanvas({})
    expect(result.props.children).toBeDefined()
    expect(result.props.children).toHaveLength(3)
  })

  it("gradient blobs include motion-safe animation classes", () => {
    const result = GlassCanvas({})
    const blobs = result.props.children
    if (Array.isArray(blobs)) {
      for (const blob of blobs) {
        expect(blob.props.className).toContain("motion-safe:animate-pulse")
      }
    }
  })
})
