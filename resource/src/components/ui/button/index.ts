import type { VariantProps } from "class-variance-authority"
import { cva } from "class-variance-authority"

export { default as Button } from "./Button.vue"

export const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg:not([class*='size-'])]:size-4 shrink-0 [&_svg]:shrink-0 outline-none focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px] aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground hover:bg-primary/90",
        brand:
          "rounded-[var(--gf-radius-field)] bg-primary font-semibold text-primary-content hover:bg-primary",
        neutral:
          "rounded-[var(--gf-radius-field)] bg-neutral font-semibold text-neutral-content hover:bg-neutral/90",
        surface:
          "rounded-[var(--gf-radius-field)] border border-line bg-base-100 font-semibold text-base-content/75 shadow-none hover:bg-base-200 hover:text-base-content",
        "brand-ghost":
          "rounded-[var(--gf-radius-field)] font-semibold text-primary shadow-none hover:bg-info/10 hover:text-primary",
        muted:
          "rounded-[var(--gf-radius-field)] font-semibold text-base-content/55 shadow-none hover:bg-base-300 hover:text-base-content",
        danger:
          "rounded-[var(--gf-radius-field)] bg-error font-semibold text-error-content hover:bg-error/90 focus-visible:ring-error/20",
        destructive:
          "bg-destructive text-white hover:bg-destructive/90 focus-visible:ring-destructive/20 dark:focus-visible:ring-destructive/40 dark:bg-destructive/60",
        outline:
          "border bg-background shadow-xs hover:bg-accent hover:text-accent-foreground dark:bg-input/30 dark:border-input dark:hover:bg-input/50",
        secondary:
          "bg-secondary text-secondary-foreground hover:bg-secondary/80",
        ghost:
          "hover:bg-accent hover:text-accent-foreground dark:hover:bg-accent/50",
        link: "text-primary underline-offset-4 hover:underline",
      },
      size: {
        "default": "h-9 px-4 py-2 has-[>svg]:px-3",
        "sm": "h-8 rounded-md gap-1.5 px-3 has-[>svg]:px-2.5",
        "lg": "h-10 rounded-md px-6 has-[>svg]:px-4",
        "xl": "h-11 rounded-md px-4 has-[>svg]:px-3",
        "icon": "size-9",
        "icon-sm": "size-8",
        "icon-lg": "size-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
)
export type ButtonVariants = VariantProps<typeof buttonVariants>
