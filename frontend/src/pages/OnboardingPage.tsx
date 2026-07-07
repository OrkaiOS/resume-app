import { useState, useCallback } from "react"
import { Check, ChevronDown, Loader2 } from "lucide-react"

import { Card, CardTitle, CardContent } from "@/components/ui/card"
import { Progress } from "@/components/ui/progress"

import { useOnboardingStatus } from "@/api/onboarding"
import { LLMConfigSection } from "@/features/onboarding/LLMConfigSection"
import { ProfileSection } from "@/features/onboarding/ProfileSection"
import { OrkaiSetupSection } from "@/features/onboarding/OrkaiSetupSection"

interface OnboardingPageProps {}

const STEPS = [
  { id: 0, label: "Connect LLM", description: "Choose your AI provider" },
  { id: 1, label: "Your Profile", description: "Tell us about yourself" },
  { id: 2, label: "Orkai Setup", description: "Wire up your workspace" },
] as const

function OnboardingPage(_props: OnboardingPageProps) {
  const [openStep, setOpenStep] = useState(0)
  const { data: status, isLoading } = useOnboardingStatus()

  const handleStepComplete = useCallback((stepId: number) => {
    if (stepId < STEPS.length - 1) {
      setOpenStep(stepId + 1)
    }
  }, [])

  const stepDone = useCallback(
    (stepId: number) => {
      if (!status) return false
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

  const completedCount = STEPS.filter((s) => stepDone(s.id)).length
  const progress = (completedCount / STEPS.length) * 100

  if (isLoading) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center bg-background">
        <div className="flex flex-col items-center gap-4">
          <Loader2 className="size-8 animate-spin text-primary" />
          <p className="text-muted-foreground">Loading your workspace...</p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen flex-col items-center bg-background py-8">
      <div className="w-full max-w-2xl px-4">
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-semibold tracking-tight text-foreground">
            Set Up Your Workspace
          </h1>
          <p className="mt-2 text-muted-foreground">
            Three steps to connect your tools and build better resumes
          </p>
          <div className="mx-auto mt-6 flex max-w-xs items-center gap-3">
            <Progress value={progress} className="h-1.5 flex-1" />
            <span className="text-xs tabular-nums text-muted-foreground">
              {completedCount}/{STEPS.length}
            </span>
          </div>
        </div>

        <div className="space-y-3">
          {STEPS.map((step) => (
            <Card key={step.id} className={step.id === openStep ? "ring-2 ring-primary/20" : ""}>
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
                        ? "ring-2 ring-primary text-primary"
                        : "ring-1 ring-foreground/15 text-muted-foreground"
                  }`}
                >
                  {stepDone(step.id) ? <Check className="size-4" /> : step.id + 1}
                </div>
                <div className="flex-1">
                  <CardTitle className="text-base">{step.label}</CardTitle>
                  <p className="mt-0.5 text-xs text-muted-foreground">
                    {step.description}
                  </p>
                </div>
                <ChevronDown
                  className={`size-5 shrink-0 text-muted-foreground transition-transform ${
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
