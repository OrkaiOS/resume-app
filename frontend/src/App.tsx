import { Button } from "@/components/ui/button"
import { OrkaiHealthProvider, useOrkaiHealth } from "@/hooks/useOrkaiHealth"
import OrkaiStatusPage from "@/pages/OrkaiStatusPage"

function AppContent() {
  const { isOrkaiRunning } = useOrkaiHealth()

  if (isOrkaiRunning === null) {
    return <OrkaiStatusPage checking />
  }

  if (!isOrkaiRunning) {
    return <OrkaiStatusPage />
  }

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

function App() {
  return (
    <OrkaiHealthProvider>
      <AppContent />
    </OrkaiHealthProvider>
  )
}

export default App