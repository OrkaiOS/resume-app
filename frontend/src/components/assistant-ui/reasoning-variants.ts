import { cva, type VariantProps } from "class-variance-authority";

const reasoningVariants = cva("aui-reasoning-root mb-4 w-full", {
  variants: {
    variant: {
      outline: "rounded-lg border px-3 py-2",
      ghost: "",
      muted: "bg-muted/50 rounded-lg px-3 py-2",
    },
  },
  defaultVariants: {
    variant: "outline",
  },
});

export { reasoningVariants };
export type ReasoningVariants = VariantProps<typeof reasoningVariants>;
