import { describe, it, expect } from "vitest"
import { resolveAppGate } from "./gate"

describe("resolveAppGate", () => {
  it("returns backend-loading when backend reachability is unknown", () => {
    expect(resolveAppGate(null, null, null)).toBe("backend-loading")
    expect(resolveAppGate(null, null, true)).toBe("backend-loading")
    expect(resolveAppGate(null, null, false)).toBe("backend-loading")
  })

  it("returns backend-down when backend is not reachable", () => {
    expect(resolveAppGate(false, null, null)).toBe("backend-down")
    expect(resolveAppGate(false, null, true)).toBe("backend-down")
    expect(resolveAppGate(false, null, false)).toBe("backend-down")
    expect(resolveAppGate(false, true, true)).toBe("backend-down")
  })

  it("returns orkai-loading when backend is up but orkai status is unknown", () => {
    expect(resolveAppGate(true, null, null)).toBe("orkai-loading")
    expect(resolveAppGate(true, null, true)).toBe("orkai-loading")
    expect(resolveAppGate(true, null, false)).toBe("orkai-loading")
  })

  it("returns orkai-down when orkai is not running", () => {
    expect(resolveAppGate(true, false, null)).toBe("orkai-down")
    expect(resolveAppGate(true, false, true)).toBe("orkai-down")
    expect(resolveAppGate(true, false, false)).toBe("orkai-down")
  })

  it("returns onboarding-loading when orkai is up but onboarding status is unknown", () => {
    expect(resolveAppGate(true, true, null)).toBe("onboarding-loading")
  })

  it("returns onboarding when orkai is up and user is not onboarded", () => {
    expect(resolveAppGate(true, true, false)).toBe("onboarding")
  })

  it("returns app when backend and orkai are up and user is onboarded", () => {
    expect(resolveAppGate(true, true, true)).toBe("app")
  })
})
