import { describe, it, expect } from "vitest"
import { resolveAppGate } from "./gate"

describe("resolveAppGate", () => {
  it("returns orkai-loading when orkai health is unknown", () => {
    expect(resolveAppGate(null, null)).toBe("orkai-loading")
    expect(resolveAppGate(null, true)).toBe("orkai-loading")
    expect(resolveAppGate(null, false)).toBe("orkai-loading")
  })

  it("returns orkai-down when orkai is not running", () => {
    expect(resolveAppGate(false, null)).toBe("orkai-down")
    expect(resolveAppGate(false, true)).toBe("orkai-down")
    expect(resolveAppGate(false, false)).toBe("orkai-down")
  })

  it("returns onboarding-loading when orkai is up but onboarding status is unknown", () => {
    expect(resolveAppGate(true, null)).toBe("onboarding-loading")
  })

  it("returns onboarding when orkai is up and user is not onboarded", () => {
    expect(resolveAppGate(true, false)).toBe("onboarding")
  })

  it("returns app when orkai is up and user is onboarded", () => {
    expect(resolveAppGate(true, true)).toBe("app")
  })
})
