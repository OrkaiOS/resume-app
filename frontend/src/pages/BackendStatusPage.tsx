import { ServerOff, RotateCcw } from "lucide-react"
import { Button } from "@/components/ui/button"

function BackendStatusPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="mx-4 w-full max-w-md rounded-xl bg-card p-8 text-card-foreground ring-1 ring-foreground/10">
        <div className="flex flex-col items-center text-center">
          <div className="mb-5 flex size-10 items-center justify-center rounded-full bg-destructive/10">
            <ServerOff className="size-5 text-destructive" />
          </div>
          <h2 className="text-xl font-semibold text-foreground">
            Backend is unreachable
          </h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Resume App cannot reach its server. Start the backend, then retry.
          </p>
          <div className="mt-5 w-full rounded-lg bg-muted p-4 text-left">
            <p className="text-xs font-medium text-muted-foreground">
              To start the backend:
            </p>
            <code className="mt-2 block text-sm text-foreground">
              cd backend && go run ./cmd
            </code>
          </div>
          <Button
            className="mt-5 w-full"
            onClick={() => window.location.reload()}
          >
            <RotateCcw className="mr-2 size-4" />
            Retry Connection
          </Button>
        </div>
      </div>
    </div>
  )
}

export default BackendStatusPage
