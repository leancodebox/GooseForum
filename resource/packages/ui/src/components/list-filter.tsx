import type { ComponentProps } from "react"
import { Slot } from "radix-ui"
import { cn } from "cn"
import { buttonVariants } from "./button"

function ListFilter({ asChild = false, className, ...props }: ComponentProps<"div"> & { asChild?: boolean }) {
  const Comp = asChild ? Slot.Root : "div"
  return <Comp data-filter="group" className={cn("inline-flex h-8 w-fit min-w-0 max-w-full items-center justify-start gap-2 overflow-x-auto rounded-none bg-transparent p-0.5", className)} {...props} />
}

function ListFilterItem({ asChild = false, className, ...props }: ComponentProps<"button"> & { asChild?: boolean }) {
  const Comp = asChild ? Slot.Root : "button"
  return <Comp data-filter="item" className={cn(
    buttonVariants({ variant: "secondary", size: "sm" }),
    "flex-none px-3 shadow-none dark:text-secondary-foreground",
    "aria-[current=page]:bg-primary aria-[current=page]:text-primary-foreground aria-[current=page]:hover:bg-primary/80 data-[state=active]:border-transparent data-[state=active]:bg-primary data-[state=active]:text-primary-foreground data-[state=active]:shadow-none data-[state=active]:hover:bg-primary/80 dark:data-[state=active]:border-transparent dark:data-[state=active]:bg-primary dark:data-[state=active]:text-primary-foreground",
    className,
  )} {...props} />
}

export { ListFilter, ListFilterItem }
