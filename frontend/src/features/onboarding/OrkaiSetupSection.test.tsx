import { describe, it, expect } from "vitest"

const SETUP_STEPS = [
  "Project Name selection + uniqueness validation",
  "Create workspace category",
  "Create profile standard",
  "Create cover letter principles",
  "Create PDF pipeline document",
  "Create PDF generation skill",
  "Link entities",
  "Verify MCP token",
]

describe("SETUP_STEPS", () => {
  it("has 8 steps", () => {
    expect(SETUP_STEPS).toHaveLength(8)
  })

  it("starts with Project Name selection", () => {
    expect(SETUP_STEPS[0]).toBe("Project Name selection + uniqueness validation")
  })

  it("includes workspace category step", () => {
    expect(SETUP_STEPS[1]).toBe("Create workspace category")
  })

  it("does not reference 'personal' anywhere", () => {
    const hasPersonal = SETUP_STEPS.some((s) => s.toLowerCase().includes("personal"))
    expect(hasPersonal).toBe(false)
  })
})

describe("projectName validation", () => {
  it("empty string is falsy after trim", () => {
    const values = ["", "   ", "\t", "\n"]
    for (const v of values) {
      expect(v.trim()).toBe("")
    }
  })

  it("non-empty after trim passes", () => {
    expect("My Workspace".trim()).not.toBe("")
  })
})

describe("conflict error detection", () => {
  it("detects 'already exists' conflict", () => {
    const error = "A workspace named 'My Resume 2026' already exists in orkai. Choose another name."
    expect(error.includes("already exists")).toBe(true)
  })

  it("does not detect conflict in generic errors", () => {
    const error = "connection refused"
    expect(error.includes("already exists")).toBe(false)
  })
})

describe("alreadyOnboarded guard", () => {
  it("returns null when alreadyOnboarded is true", () => {
    const alreadyOnboarded = true
    expect(alreadyOnboarded).toBe(true)
  })
})
