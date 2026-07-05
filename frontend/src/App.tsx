import { Button } from '@/components/ui/button'

function App() {
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

export default App