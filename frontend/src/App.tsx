import { lazy, Suspense } from "react"
import { Sparkles } from "lucide-react"
import BackendHealthProvider from "@/hooks/BackendHealthProvider"
import { useBackendHealth } from "@/hooks/useBackendHealth"
import useOnboardingStatus from "@/hooks/useOnboardingStatus"
import { resolveAppGate } from "@/lib/gate"
import BackendStatusPage from "@/pages/BackendStatusPage"
import OnboardingPage from "@/pages/OnboardingPage"
import OrkaiStatusPage from "@/pages/OrkaiStatusPage"
import HomePage from "@/pages/HomePage"
import GlassCanvas from "@/components/ui/glass-canvas"

const ChatPage = lazy(() => import("@/pages/ChatPage"))

const isChatRoute = () => window.location.search.includes("chat")

function AppContent() {
  const { isBackendReachable, isOrkaiRunning } = useBackendHealth()
  const { isOnboarded } = useOnboardingStatus()

  if (isChatRoute()) {
    return (
      <Suspense fallback={
        <div className="flex min-h-screen items-center justify-center bg-background">
          <Sparkles className="size-8 animate-pulse text-primary" />
        </div>
      }
      >
        <ChatPage />
      </Suspense>
    )
  }

  const gate = resolveAppGate(isBackendReachable, isOrkaiRunning, isOnboarded)

  switch (gate) {
    case "backend-loading":
      return (
        <div className="flex min-h-screen items-center justify-center bg-background">
          <div className="flex flex-col items-center gap-4">
            <Sparkles className="size-8 animate-pulse text-primary" />
            <p className="text-sm text-muted-foreground">Checking backend connection...</p>
          </div>
        </div>
      )
    case "backend-down":
      return <BackendStatusPage />
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
      return <HomePage />
  }
}

function App() {
  return (
    <BackendHealthProvider>
      <GlassCanvas />
      <AppContent />
    </BackendHealthProvider>
  )
}

export default App
