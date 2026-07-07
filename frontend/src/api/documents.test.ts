import { describe, it, expect } from "vitest"
import { documentsKeys } from "./documents"

describe("documentsKeys", () => {
  it("all returns the base key", () => {
    expect(documentsKeys.all).toEqual(["documents"])
  })

  it("resume returns a scoped key by opportunity id", () => {
    expect(documentsKeys.resume("abc-123")).toEqual([
      "documents",
      "resume",
      "abc-123",
    ])
  })

  it("coverLetter returns a scoped key by opportunity id", () => {
    expect(documentsKeys.coverLetter("xyz-456")).toEqual([
      "documents",
      "cover-letter",
      "xyz-456",
    ])
  })

  it("resume and coverLetter keys are independent", () => {
    expect(documentsKeys.resume("abc")).not.toEqual(
      documentsKeys.coverLetter("abc")
    )
  })
})
