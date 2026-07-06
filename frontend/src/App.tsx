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
        <div className="flex min-h-screen items-center justify-center">
          <p className="text-muted-foreground">Loading...</p>
        </div>
      )
    case "onboarding":
      return <OnboardingPage />
    case "app":
      return (
        <div className="flex min-h-screen items-center justify-center">
          <div className="text-center">
            <h1 className="text-4xl font-bold">Resume App</h1>
            <p className="mt-4 text-muted-foreground">Build beautiful resumes</p>
            <Button className="mt-6">Get Started</Button>
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