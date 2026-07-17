import { useState } from "react"
import { Loader2, RotateCcw, Check, Circle } from "lucide-react"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Progress } from "@/components/ui/progress"

import { useTriggerSetup, useSetupStatus } from "@/api/onboarding"
import type { OrkaiSetupStep } from "@/types/api"

interface OrkaiSetupSectionProps {
  disabled: boolean
  onComplete: () => void
  alreadyOnboarded?: boolean
}

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

function OrkaiSetupSection({ disabled, onComplete, alreadyOnboarded }: OrkaiSetupSectionProps) {
  const [setupId, setSetupId] = useState<string | null>(null)
  const [setupStarted, setSetupStarted] = useState(false)
  const [projectName, setProjectName] = useState("")
  const [projectNameTouched, setProjectNameTouched] = useState(false)

  const triggerSetup = useTriggerSetup()
  const { data: setupStatus, isLoading } = useSetupStatus(setupId ?? "", setupStarted)

  function handleStart() {
    const name = projectName.trim()
    if (!name) {
      setProjectNameTouched(true)
      return
    }
    triggerSetup.mutate({ projectName: name }, {
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
    setProjectNameTouched(false)
  }

  const steps = setupStatus?.steps ?? []
  const completedCount = steps.filter((s) => s.status === "success").length
  const progress = steps.length > 0 ? (completedCount / steps.length) * 100 : 0
  const isComplete = setupStatus?.completed ?? false

  const step0Failed = steps[0]?.status === "failed"
  const step0Error = steps[0]?.error ?? ""
  const isConflictError = step0Failed && step0Error.includes("already exists")
  const showEmptyError = projectNameTouched && !projectName.trim()

  if (alreadyOnboarded) return null

  if (!setupStarted) {
    return (
      <div className="space-y-4 pt-2">
        <p className="text-sm text-muted-foreground">
          Connect Orkai to your workspace with a one-click setup. This creates
          the standards, skills, and documents needed for resume generation.
        </p>
        <div className="space-y-2">
          <label className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70" htmlFor="project-name">
            Project Name
          </label>
          <p className="text-xs text-muted-foreground">
            Choose a name for your workspace. This will organize your resume
            standards, skills, and documents in Orkai.
          </p>
          <Input
            id="project-name"
            placeholder="e.g. My Resume 2026"
            maxLength={64}
            required
            value={projectName}
            onChange={(e) => {
              setProjectName(e.target.value)
              if (e.target.value.trim()) setProjectNameTouched(false)
            }}
            onBlur={() => setProjectNameTouched(true)}
            disabled={disabled || triggerSetup.isPending}
            onKeyDown={(e) => {
              if (e.key === "Enter" && projectName.trim()) handleStart()
            }}
          />
          {showEmptyError && (
            <p className="text-sm text-destructive">Project name is required.</p>
          )}
        </div>
        <Button
          className="w-full"
          disabled={disabled || triggerSetup.isPending || !projectName.trim()}
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

      {isConflictError && (
        <div className="space-y-3 rounded-md border border-destructive/50 p-3">
          <p className="text-sm font-medium text-destructive">Workspace name already taken</p>
          <div className="space-y-2">
            <label className="text-sm font-medium" htmlFor="project-name-retry">
              Project Name
            </label>
            <Input
              id="project-name-retry"
              placeholder="e.g. My Resume 2026"
              maxLength={64}
              required
              value={projectName}
              onChange={(e) => setProjectName(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter" && projectName.trim()) handleRetry()
              }}
            />
            <p className="text-sm text-destructive">
              A workspace named &apos;{projectName}&apos; already exists in Orkai. Choose a different name to continue.
            </p>
          </div>
          <Button
            className="w-full"
            disabled={!projectName.trim()}
            onClick={handleRetry}
          >
            Retry Setup
          </Button>
        </div>
      )}

      {isComplete && (
        <Button className="w-full" onClick={onComplete}>
          Finish
        </Button>
      )}

      {!isComplete && steps.some((s) => s.status === "failed") && !isConflictError && (
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
