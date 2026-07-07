import { useState } from "react"
import { Loader2, RotateCcw, Check, Circle } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"

import { useTriggerSetup, useSetupStatus } from "@/api/onboarding"
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

  const triggerSetup = useTriggerSetup()
  const { data: setupStatus, isLoading } = useSetupStatus(setupId ?? "", setupStarted)

  function handleStart() {
    triggerSetup.mutate(undefined, {
      onSuccess: (res) => {
        if (res.data) {
          setSetupId(res.data.sessionId)
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
          Connect Orkai to your workspace with a one-click setup. This creates
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
              Connecting to Orkai...
            </>
          ) : (
            "Start Setup"
          )}
        </Button>
        {triggerSetup.isError && (
          <p className="text-sm text-destructive">
            Could not connect to Orkai. Check that the daemon is running and try
            again.
          </p>
        )}
        {disabled && (
          <p className="text-xs text-muted-foreground">
            Complete the previous steps before setting up Orkai.
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

      <div className="space-y-2">
        {isLoading ? (
          <div className="flex items-center gap-3 py-4">
            <Loader2 className="size-4 animate-spin text-muted-foreground" />
            <span className="text-sm text-muted-foreground">
              Setting up your workspace — creating standards, skills, and
              documents for resume generation. This may take a moment.
            </span>
          </div>
        ) : steps.length === 0 ? (
          <p className="py-4 text-sm text-muted-foreground">
            Setup steps are being prepared. You'll see a checklist of items
            created automatically — standards, skills, and documents for your
            workspace. No action needed — this runs on its own.
          </p>
        ) : (
          SETUP_STEPS.map((stepName, index) => {
          const step = steps[index]
          const status = step?.status ?? "pending"
          const error = step?.error

          return (
            <div key={stepName}>
              <div className="flex items-center gap-3 rounded-md px-3 py-2">
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
              {status === "failed" && error && (
                <p className="ml-8 text-xs text-destructive">{error}</p>
              )}
            </div>
          )
          })
        )}
      </div>

      {isComplete && (
        <Button className="w-full" onClick={onComplete}>
          Finish
        </Button>
      )}

      {!isComplete && steps.some((s) => s.status === "failed") && (
        <div className="space-y-3">
          <Button
            className="w-full"
            onClick={handleRetry}
          >
            <RotateCcw className="mr-2 size-4" />
            Retry Setup
          </Button>
          <button
            type="button"
            className="mx-auto block text-xs text-muted-foreground underline underline-offset-2 hover:text-foreground"
            onClick={onComplete}
          >
            Continue without setup
          </button>
        </div>
      )}
    </div>
  )
}

export { OrkaiSetupSection }
export type { OrkaiSetupSectionProps }
