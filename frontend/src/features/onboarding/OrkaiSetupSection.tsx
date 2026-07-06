import { useState } from "react"
import { Loader2, RotateCcw, Check, Circle } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"

import { useTriggerOrkaiSetup, useOrkaiSetupStatus } from "@/api/onboarding"
import type { OrkaiSetupStep } from "@/types/api"

interface OrkaiSetupSectionProps {
  disabled: boolean
  onComplete: () => void
}

const SETUP_STEPS = [
  "Create personal category",
  "Create profile standard",
  "Create cover letter principles",
  "Create PDF pipeline document",
  "Create PDF generation skill",
  "Link entities",
  "Verify MCP token",
]

function stepStatusIcon(status: OrkaiSetupStep["status"]) {
  switch (status) {
    case "success":
      return <Check className="size-4 text-primary" />
    case "in_progress":
      return <Loader2 className="size-4 animate-spin text-primary" />
    case "failed":
      return <Circle className="size-4 text-destructive" />
    default:
      return <Circle className="size-4 text-muted-foreground/40" />
  }
}

function OrkaiSetupSection({ disabled, onComplete }: OrkaiSetupSectionProps) {
  const [setupId, setSetupId] = useState<string | null>(null)
  const [setupStarted, setSetupStarted] = useState(false)

  const triggerSetup = useTriggerOrkaiSetup()
  const { data: setupStatus } = useOrkaiSetupStatus(setupId ?? "", setupStarted)

  function handleStart() {
    triggerSetup.mutate(undefined, {
      onSuccess: (res) => {
        if (res.data) {
          setSetupId(res.data.setupId)
          setSetupStarted(true)
        }
      },
    })
  }

  function handleRetry() {
    setSetupId(null)
    setSetupStarted(false)
  }

  const steps = setupStatus?.steps ?? []
  const completedCount = steps.filter((s) => s.status === "success").length
  const progress = steps.length > 0 ? (completedCount / steps.length) * 100 : 0
  const isComplete = setupStatus?.completed ?? false

  if (!setupStarted) {
    return (
      <div className="space-y-4 pt-2">
        <p className="text-sm text-muted-foreground">
          Connect orkai to your workspace with a one-click setup. This creates
          the standards, skills, and documents needed for resume generation.
        </p>
        <Button
          className="w-full"
          disabled={disabled || triggerSetup.isPending}
          onClick={handleStart}
        >
          {triggerSetup.isPending ? (
            <>
              <Loader2 className="mr-2 size-4 animate-spin" />
              Starting...
            </>
          ) : (
            "Start Setup"
          )}
        </Button>
        {disabled && (
          <p className="text-xs text-muted-foreground">
            Complete LLM Config and Profile steps first.
          </p>
        )}
      </div>
    )
  }

  return (
    <div className="space-y-4 pt-2">
      <div className="flex items-center gap-3">
        <Progress value={progress} className="flex-1" />
        <span className="text-sm text-muted-foreground tabular-nums">
          {completedCount}/{SETUP_STEPS.length}
        </span>
      </div>

      <div className="space-y-1.5">
        {SETUP_STEPS.map((stepName, index) => {
          const step = steps[index]
          const status = step?.status ?? "pending"

          return (
            <div
              key={stepName}
              className="flex items-center gap-3 rounded-md px-3 py-2"
            >
              {stepStatusIcon(status)}
              <span className="flex-1 text-sm text-foreground">{stepName}</span>
              <Badge
                variant={
                  status === "success"
                    ? "default"
                    : status === "failed"
                      ? "destructive"
                      : status === "in_progress"
                        ? "secondary"
                        : "outline"
                }
              >
                {status === "in_progress" ? "In Progress" : status}
              </Badge>
            </div>
          )
        })}
      </div>

      {isComplete && (
        <Button className="w-full" onClick={onComplete}>
          Finish
        </Button>
      )}

      {!isComplete && steps.some((s) => s.status === "failed") && (
        <Button
          variant="outline"
          className="w-full"
          onClick={handleRetry}
        >
          <RotateCcw className="mr-2 size-4" />
          Retry Setup
        </Button>
      )}
    </div>
  )
}

export { OrkaiSetupSection }
export type { OrkaiSetupSectionProps }
