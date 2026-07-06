import { useState, useCallback } from "react"
import { Check, ChevronDown } from "lucide-react"

import { Card, CardTitle, CardContent } from "@/components/ui/card"

import { useOnboardingStatus } from "@/api/onboarding"
import { LLMConfigSection } from "@/features/onboarding/LLMConfigSection"
import { ProfileSection } from "@/features/onboarding/ProfileSection"
import { OrkaiSetupSection } from "@/features/onboarding/OrkaiSetupSection"

interface OnboardingPageProps {}

const STEPS = [
  { id: 0, label: "LLM Config" },
  { id: 1, label: "Profile" },
  { id: 2, label: "Orkai Setup" },
] as const

function OnboardingPage(_props: OnboardingPageProps) {
  const [openStep, setOpenStep] = useState(0)
  const { data: status } = useOnboardingStatus()

  const handleStepComplete = useCallback((stepId: number) => {
    if (stepId < STEPS.length - 1) {
      setOpenStep(stepId + 1)
    }
  }, [])

  const stepDone = useCallback(
    (stepId: number) => {
      if (!status) return false
      if (!status.onboarded) return false
      switch (stepId) {
        case 0:
          return status.steps.llmConfig
        case 1:
          return status.steps.profile
        case 2:
          return status.steps.orkaiSetup
        default:
          return false
      }
    },
    [status]
  )

  return (
    <div className="flex min-h-screen flex-col items-center bg-background py-8">
      <div className="w-full max-w-3xl px-4">
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-bold text-foreground">Welcome to Resume App</h1>
          <p className="mt-2 text-muted-foreground">
            Let&apos;s get you set up in three steps
          </p>
        </div>

        <div className="space-y-3">
          {STEPS.map((step) => (
            <Card key={step.id}>
              <button
                type="button"
                className="flex w-full items-center gap-4 p-4 text-left"
                onClick={() => setOpenStep(step.id)}
              >
                <div
                  className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-sm font-medium ${
                    stepDone(step.id)
                      ? "bg-primary text-primary-foreground"
                      : step.id === openStep
                        ? "border-2 border-primary text-primary"
                        : "border-2 border-muted-foreground/30 text-muted-foreground"
                  }`}
                >
                  {stepDone(step.id) ? <Check className="size-4" /> : step.id + 1}
                </div>
                <div className="flex-1">
                  <CardTitle className="text-base">{step.label}</CardTitle>
                </div>
                <ChevronDown
                  className={`size-5 text-muted-foreground transition-transform ${
                    openStep === step.id ? "rotate-180" : ""
                  }`}
                />
              </button>

              {openStep === step.id && (
                <CardContent className="border-t px-4 pb-4">
                  {step.id === 0 && (
                    <LLMConfigSection onComplete={() => handleStepComplete(0)} />
                  )}
                  {step.id === 1 && (
                    <ProfileSection onComplete={() => handleStepComplete(1)} />
                  )}
                  {step.id === 2 && (
                    <OrkaiSetupSection
                      disabled={!status?.steps.llmConfig || !status?.steps.profile}
                      onComplete={() => handleStepComplete(2)}
                    />
                  )}
                </CardContent>
              )}
            </Card>
          ))}
        </div>
      </div>
    </div>
  )
}

export default OnboardingPage
