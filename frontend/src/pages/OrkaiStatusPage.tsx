import { Loader2, ExternalLink, RotateCcw } from "lucide-react"
import { Button } from "@/components/ui/button"

interface OrkaiStatusPageProps {
  checking?: boolean
}

function OrkaiStatusPage({ checking = false }: OrkaiStatusPageProps) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="mx-4 w-full max-w-md rounded-xl bg-card p-8 text-card-foreground ring-1 ring-foreground/10">
        <div className="flex flex-col items-center text-center">
          {checking ? (
            <>
              <Loader2 className="mb-5 size-10 animate-spin text-primary" />
              <h2 className="text-xl font-semibold text-foreground">
                Checking Orkai status
              </h2>
              <p className="mt-2 text-sm text-muted-foreground">
                Verifying the Orkai daemon is reachable...
              </p>
            </>
          ) : (
            <>
              <div className="mb-5 flex size-10 items-center justify-center rounded-full bg-muted">
                <span className="text-lg font-bold text-muted-foreground">
                  !
                </span>
              </div>
              <h2 className="text-xl font-semibold text-foreground">
                Orkai is not running
              </h2>
              <p className="mt-2 text-sm text-muted-foreground">
                Resume App requires Orkai to function. Start the Orkai daemon,
                then retry.
              </p>
              <div className="mt-5 w-full rounded-lg bg-muted p-4 text-left">
                <p className="text-xs font-medium text-muted-foreground">
                  To start Orkai:
                </p>
                <code className="mt-2 block text-sm text-foreground">
                  orkai serve
                </code>
                <p className="mt-3 text-xs text-muted-foreground">
                  Not installed yet?{" "}
                  <a
                    href="https://github.com/cosmixyz/orkai"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center gap-1 text-primary underline underline-offset-2 hover:text-primary/80"
                  >
                    Install Orkai from GitHub
                    <ExternalLink className="size-3" />
                  </a>
                </p>
              </div>
              <Button
                className="mt-5 w-full"
                onClick={() => window.location.reload()}
              >
                <RotateCcw className="mr-2 size-4" />
                Retry Connection
              </Button>
            </>
          )}
        </div>
      </div>
    </div>
  )
}

export { type OrkaiStatusPageProps }
export default OrkaiStatusPage