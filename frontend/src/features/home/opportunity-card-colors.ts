// @orkai:decision Colored accent bar — 3px border-t with deterministic color from opportunity ID hash, stable across renders.

export const CARD_ACCENT_COLORS = [
  "border-t-blue-500",
  "border-t-emerald-500",
  "border-t-violet-500",
  "border-t-amber-500",
  "border-t-rose-500",
  "border-t-cyan-500",
  "border-t-teal-500",
  "border-t-indigo-500",
] as const

export function getAccentIndex(id: string): number {
  let hash = 0
  for (let i = 0; i < id.length; i++) {
    hash = ((hash << 5) - hash) + id.charCodeAt(i)
    hash |= 0
  }
  return Math.abs(hash) % CARD_ACCENT_COLORS.length
}
