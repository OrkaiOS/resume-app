import { describe, it, expect } from "vitest"
import { createOpportunitySchema } from "./create-opportunity-schema"

describe("createOpportunitySchema", () => {
  it("accepts a valid opportunity with description", () => {
    const result = createOpportunitySchema.safeParse({
      company: "Acme",
      role: "Backend Developer",
      description: "Test role",
    })
    expect(result.success).toBe(true)
  })

  it("rejects when description is empty", () => {
    const result = createOpportunitySchema.safeParse({
      company: "Acme",
      role: "Backend Developer",
      description: "",
    })
    expect(result.success).toBe(false)
  })

  it("rejects when description is missing", () => {
    const result = createOpportunitySchema.safeParse({
      company: "Acme",
      role: "Backend Developer",
    })
    expect(result.success).toBe(false)
  })

  it("rejects when company is missing", () => {
    const result = createOpportunitySchema.safeParse({
      role: "Backend Developer",
    })
    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues.some((i) => i.path[0] === "company")).toBe(true)
    }
  })

  it("rejects when company is empty string", () => {
    const result = createOpportunitySchema.safeParse({
      company: "",
      role: "Backend Developer",
    })
    expect(result.success).toBe(false)
  })

  it("rejects when role is missing", () => {
    const result = createOpportunitySchema.safeParse({
      company: "Acme",
    })
    expect(result.success).toBe(false)
    if (!result.success) {
      expect(result.error.issues.some((i) => i.path[0] === "role")).toBe(true)
    }
  })

  it("rejects when role is empty string", () => {
    const result = createOpportunitySchema.safeParse({
      company: "Acme",
      role: "",
    })
    expect(result.success).toBe(false)
  })
})