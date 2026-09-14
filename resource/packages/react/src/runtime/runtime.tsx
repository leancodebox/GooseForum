"use client";

import {
  createContext,
  useContext,
  type AnchorHTMLAttributes,
  type MouseEvent,
  type ReactNode,
} from "react";
import type { AnyPagePayload, GooseSiteApi } from "@gooseforum/client";
import type { ThemePayload } from "@gooseforum/client";
import type { AuthLocale } from "../i18n/auth";

export interface GooseNavigateOptions {
  replace?: boolean;
}

export interface GooseRuntime {
  api: GooseSiteApi;
  currentUrl: string;
  isNavigating: boolean;
  theme: ThemePayload["current"];
  locale: AuthLocale;
  navigate(href: string, options?: GooseNavigateOptions): Promise<void>;
  registerNavigationBlocker?(
    blocker: (href: string) => boolean | Promise<boolean>,
  ): () => void;
  fetchPage?(href: string, signal?: AbortSignal): Promise<AnyPagePayload>;
  queueFlash(message: string, type?: GooseFlashType): void;
  redirect(href: string): void;
  refresh(): Promise<void>;
  setLocale(locale: AuthLocale): void;
  toggleTheme(): void;
}

export type GooseFlashType = "success" | "info" | "warning" | "error";

const GooseRuntimeContext = createContext<GooseRuntime | null>(null);

export function GooseRuntimeProvider({
  runtime,
  children,
}: {
  runtime: GooseRuntime;
  children: ReactNode;
}) {
  return (
    <GooseRuntimeContext.Provider value={runtime}>
      {children}
    </GooseRuntimeContext.Provider>
  );
}

export function useGooseRuntime() {
  const runtime = useContext(GooseRuntimeContext);
  if (!runtime) {
    throw new Error("useGooseRuntime must be used inside GooseRuntimeProvider");
  }
  return runtime;
}

export interface GooseLinkProps
  extends AnchorHTMLAttributes<HTMLAnchorElement> {
  href: string;
  replace?: boolean;
}

export function GooseLink({
  href,
  replace,
  onClick,
  ...props
}: GooseLinkProps) {
  const runtime = useGooseRuntime();

  function handleClick(event: MouseEvent<HTMLAnchorElement>) {
    onClick?.(event);
    if (!shouldNavigateInApp(event, href)) return;

    event.preventDefault();
    void runtime.navigate(href, { replace });
  }

  return <a {...props} href={href} onClick={handleClick} />;
}

function shouldNavigateInApp(
  event: MouseEvent<HTMLAnchorElement>,
  href: string,
) {
  if (
    event.defaultPrevented ||
    event.button !== 0 ||
    event.metaKey ||
    event.ctrlKey ||
    event.shiftKey ||
    event.altKey ||
    event.currentTarget.target === "_blank" ||
    event.currentTarget.hasAttribute("download") ||
    href.startsWith("#")
  ) {
    return false;
  }

  const document = event.currentTarget.ownerDocument;
  const target = new URL(href, document.baseURI);
  return target.origin === document.location.origin;
}
