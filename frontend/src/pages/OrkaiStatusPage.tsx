import { Loader2 } from "lucide-react"

interface OrkaiStatusPageProps {
  checking?: boolean
}

function OrkaiStatusPage({ checking = false }: OrkaiStatusPageProps) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="mx-4 w-full max-w-md rounded-lg border border-border bg-card p-8 text-card-foreground shadow-sm">
        <div className="flex flex-col items-center text-center">
          {checking ? (
            <>
              <Loader2 className="mb-4 size-10 animate-spin text-muted-foreground" />
              <h2 className="text-xl font-semibold text-foreground">
                Checking orkai status
              </h2>
              <p className="mt-2 text-sm text-muted-foreground">
                Verifying the orkai daemon is reachable...
              </p>
            </>
          ) : (
            <>
              <div className="mb-4 flex size-10 items-center justify-center rounded-full bg-muted">
                <span className="text-xl font-bold text-muted-foreground">
                  !
                </span>
              </div>
              <h2 className="text-xl font-semibold text-foreground">
                Orkai is not running
              </h2>
              <p className="mt-2 text-sm text-muted-foreground">
                Resume App requires orkai to function. Please start the orkai
                daemon and refresh this page.
              </p>
              <div className="mt-6 w-full rounded-md bg-muted p-4 text-left">
                <p className="text-xs font-medium text-muted-foreground">
                  To start orkai:
                </p>
                <code className="mt-2 block text-sm text-foreground">
                  orkai serve
                </code>
                <p className="mt-3 text-xs text-muted-foreground">
                  If you haven&apos;t installed orkai yet, visit{" "}
                  <a
                    href="https://github.com/cosmixyz/orkai"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="underline hover:text-foreground"
                  >
                    github.com/cosmixyz/orkai
                  </a>
                </p>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}

export { type OrkaiStatusPageProps }
export default OrkaiStatusPage