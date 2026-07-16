import { Sun, Droplets } from "lucide-react"
import { Button } from "@/components/ui/button"
import { useStyleStore } from "@/store/style-store"

interface Props {}

function StyleSwitcher(_props: Props) {
  const style = useStyleStore((s) => s.style)
  const setStyle = useStyleStore((s) => s.setStyle)

  const isDefault = style === "default"

  return (
    <div className="inline-flex items-center rounded-lg border p-0.5" role="radiogroup" aria-label="Style switcher">
      <Button
        variant={isDefault ? "default" : "ghost"}
        size="icon-sm"
        onClick={() => setStyle("default")}
        aria-label="Default style"
        aria-checked={isDefault}
        role="radio"
      >
        <Sun className="size-4" />
      </Button>
      <Button
        variant={isDefault ? "ghost" : "default"}
        size="icon-sm"
        onClick={() => setStyle("glass")}
        aria-label="Glass style"
        aria-checked={!isDefault}
        role="radio"
      >
        <Droplets className="size-4" />
      </Button>
    </div>
  )
}

export { StyleSwitcher }
export default StyleSwitcher
