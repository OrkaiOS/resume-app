import { describe, it, expect } from "vitest"
import { getAccentIndex } from "./opportunity-card-colors"

const PALETTE_SIZE = 8

describe("getAccentIndex", () => {
  it("returns a value within the palette range for any string", () => {
    const ids = ["a", "abc", "hello-world", "uuid-1234-5678", ""]
    for (const id of ids) {
      const index = getAccentIndex(id)
      expect(index).toBeGreaterThanOrEqual(0)
      expect(index).toBeLessThan(PALETTE_SIZE)
    }
  })

  it("returns the same index for the same input (deterministic)", () => {
    const id = "test-opportunity-id"
    const first = getAccentIndex(id)
    const second = getAccentIndex(id)
    expect(first).toBe(second)
  })

  it("returns a stable index across repeated calls", () => {
    const id = "550e8400-e29b-41d4-a716-446655440000"
    const results = Array.from({ length: 10 }, () => getAccentIndex(id))
    expect(new Set(results).size).toBe(1)
  })

  it("produces varying indices for different inputs", () => {
    const indices = new Set(
      ["id-1", "id-2", "id-3", "id-4", "id-5"].map(getAccentIndex)
    )
    expect(indices.size).toBeGreaterThanOrEqual(2)
  })

  it("handles empty string", () => {
    const index = getAccentIndex("")
    expect(index).toBeGreaterThanOrEqual(0)
    expect(index).toBeLessThan(PALETTE_SIZE)
  })

  it("handles long strings without error", () => {
    const long = "x".repeat(1000)
    const index = getAccentIndex(long)
    expect(index).toBeGreaterThanOrEqual(0)
    expect(index).toBeLessThan(PALETTE_SIZE)
  })
})
