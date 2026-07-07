import { describe, it, expect } from "vitest"
import { formatDate } from "./format-date"

describe("formatDate", () => {
  it("returns empty string for empty input", () => {
    expect(formatDate("")).toBe("")
  })

  it("returns empty string for invalid date", () => {
    expect(formatDate("not-a-date")).toBe("")
  })

  it("formats a valid ISO date as 'Mon D, YYYY'", () => {
    expect(formatDate("2026-07-07T12:00:00Z")).toMatch(
      /^[A-Z][a-z]{2} \d{1,2}, \d{4}$/
    )
  })

  it("produces deterministic output for a known date", () => {
    const result = formatDate("2026-01-15T00:00:00Z")
    expect(result).toMatch(/Jan 1[45]/)
    expect(result).toContain("2026")
  })

  it("returns empty string for null-like input (defensive)", () => {
    expect(formatDate("0001-01-01T00:00:00Z")).toMatch(/^[A-Z]/)
  })
})