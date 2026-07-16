import { useStyleStore } from "@/store/style-store"

interface Props {}

function GlassCanvas(_props: Props) {
  const style = useStyleStore((s) => s.style)

  if (style !== "glass") return null

  return (
    <div
      className="pointer-events-none fixed inset-0 -z-10 overflow-hidden"
      aria-hidden="true"
    >
      <div className="absolute -left-32 top-1/4 size-96 rounded-full bg-gradient-to-br from-primary/10 via-transparent to-chart-2/10 blur-3xl motion-safe:animate-pulse" />
      <div className="absolute -right-32 bottom-1/4 size-80 rounded-full bg-gradient-to-br from-chart-2/10 via-transparent to-primary/10 blur-3xl motion-safe:animate-pulse motion-safe:[animation-delay:1s]" />
      <div className="absolute left-1/3 bottom-1/3 size-72 rounded-full bg-gradient-to-br from-primary/10 via-transparent to-chart-1/10 blur-3xl motion-safe:animate-pulse motion-safe:[animation-delay:2s]" />
    </div>
  )
}

export { GlassCanvas }
export default GlassCanvas
