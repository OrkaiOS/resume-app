import { describe, it, expect } from "vitest"
import { opportunitiesKeys } from "./opportunities"

describe("opportunitiesKeys", () => {
  it("all returns the base key array", () => {
    expect(opportunitiesKeys.all).toEqual(["opportunities"])
  })

  it("list returns [all, 'list']", () => {
    expect(opportunitiesKeys.list()).toEqual(["opportunities", "list"])
  })

  it("detail returns [all, 'detail', id]", () => {
    expect(opportunitiesKeys.detail("abc")).toEqual([
      "opportunities",
      "detail",
      "abc",
    ])
  })

  it("detail produces distinct keys for different ids", () => {
    expect(opportunitiesKeys.detail("a")).not.toEqual(
      opportunitiesKeys.detail("b")
    )
  })
})