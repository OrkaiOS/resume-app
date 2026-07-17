interface Props {}

// @orkai:decision glass-canvas-visible class is gated by html[data-style="glass"] in index.css
// because Tailwind v4 arbitrary variants like [data-style=glass]:block target the element
// itself, not ancestors. The global data-style lives on <html>, so a plain CSS rule in
// index.css handles the ancestor-based visibility gating per the GlassCanvas Component rule.
function GlassCanvas(_props: Props) {
  return (
    <div
      className="pointer-events-none fixed inset-0 -z-10 overflow-hidden hidden glass-canvas-visible"
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
