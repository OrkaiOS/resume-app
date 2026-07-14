import { cva, type VariantProps } from "class-variance-authority";

const toolGroupVariants = cva("aui-tool-group-root group/tool-group w-full", {
  variants: {
    variant: {
      outline: "rounded-lg border py-3",
      ghost: "",
      muted: "border-muted-foreground/30 bg-muted/30 rounded-lg border py-3",
    },
  },
  defaultVariants: { variant: "outline" },
});

export { toolGroupVariants };
export type ToolGroupVariants = VariantProps<typeof toolGroupVariants>;
