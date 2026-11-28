"use client";

import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { runtimeText } from "@gooseforum/i18n";
import { GooseClientError } from "@gooseforum/client";

export function useServerErrorMessage() {
  const { t, i18n } = useTranslation("shared.serverMessages");
  return useCallback(
    (reason: unknown, fallback: string) => {
      if (reason instanceof GooseClientError && reason.messageCode) {
        if (i18n.exists(reason.messageCode, { ns: "shared.serverMessages" })) {
          return runtimeText(t, reason.messageCode, { replace: reason.params ?? {} });
        }
        if (reason.message === reason.messageCode) return fallback;
      }
      return reason instanceof Error && reason.message
        ? reason.message
        : fallback;
    },
    [i18n, t],
  );
}
