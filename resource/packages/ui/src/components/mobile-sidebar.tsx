import type * as React from "react"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "./sheet"

export default function MobileSidebar({ side, dir, children, open, onOpenChange, ...props }: React.ComponentProps<"div"> & {
  side: "left" | "right"
  open: boolean
  onOpenChange(open: boolean): void
}) {
  return (
    <Sheet open={open} onOpenChange={onOpenChange} {...props}>
      <SheetContent
        dir={dir}
        data-sidebar="sidebar"
        data-slot="sidebar"
        data-mobile="true"
        className="w-(--sidebar-width) bg-sidebar p-0 text-sidebar-foreground [&>button]:hidden"
        style={
          {
            "--sidebar-width": "18rem",
          } as React.CSSProperties
        }
        side={side}
      >
        <SheetHeader className="sr-only">
          <SheetTitle>Sidebar</SheetTitle>
          <SheetDescription>Displays the mobile sidebar.</SheetDescription>
        </SheetHeader>
        <div className="flex h-full w-full flex-col">{children}</div>
      </SheetContent>
    </Sheet>
  )
}
