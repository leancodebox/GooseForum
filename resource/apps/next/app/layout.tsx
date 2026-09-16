import type { Metadata } from "next";
import { cookies } from "next/headers";
import type { ReactNode } from "react";
import "./globals.css";

export const metadata: Metadata = {
  title: "GooseForum Next",
  description: "Standalone Next.js host adapter for GooseForum.",
};

export default async function RootLayout({ children }: { children: ReactNode }) {
  const theme = (await cookies()).get("goose-site-theme")?.value;
  const activeTheme = theme === "gf-dark" ? "gf-dark" : "gf-light";
  return (
    <html lang="zh" data-theme={activeTheme} suppressHydrationWarning>
      <body>{children}</body>
    </html>
  );
}
