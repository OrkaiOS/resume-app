import { FileText, Sparkles } from "lucide-react"
import { Button } from "@/components/ui/button"
import OrkaiHealthProvider from "@/hooks/OrkaiHealthProvider"
import { useOrkaiHealth } from "@/hooks/useOrkaiHealth"
import useOnboardingStatus from "@/hooks/useOnboardingStatus"
import { resolveAppGate } from "@/lib/gate"
import OnboardingPage from "@/pages/OnboardingPage"
import OrkaiStatusPage from "@/pages/OrkaiStatusPage"

function AppContent() {
  const { isOrkaiRunning } = useOrkaiHealth()
  const { isOnboarded } = useOnboardingStatus()

  const gate = resolveAppGate(isOrkaiRunning, isOnboarded)

  switch (gate) {
    case "orkai-loading":
      return <OrkaiStatusPage checking />
    case "orkai-down":
      return <OrkaiStatusPage />
    case "onboarding-loading":
      return (
        <div className="flex min-h-screen items-center justify-center bg-background">
          <div className="flex flex-col items-center gap-4">
            <Sparkles className="size-8 animate-pulse text-primary" />
            <p className="text-sm text-muted-foreground">Checking your workspace...</p>
          </div>
        </div>
      )
    case "onboarding":
      return <OnboardingPage />
    case "app":
      return (
        <div className="flex min-h-screen items-center justify-center bg-background">
          <div className="mx-auto max-w-md px-4 text-center">
            <div className="mb-6 flex justify-center">
              <div className="flex size-14 items-center justify-center rounded-2xl bg-primary/10">
                <FileText className="size-7 text-primary" />
              </div>
            </div>
            <h1 className="text-3xl font-semibold tracking-tight text-foreground">
              Resume App
            </h1>
            <p className="mt-3 text-muted-foreground">
              Build, tailor, and export professional resumes — powered by AI and
              your own standards stored in orkai.
            </p>
            <div className="mt-8">
              <Button className="w-full" size="lg">
                Build Your First Resume
              </Button>
            </div>
          </div>
        </div>
      )
  }
}

function App() {
  return (
    <OrkaiHealthProvider>
      <AppContent />
    </OrkaiHealthProvider>
  )
}

export default App