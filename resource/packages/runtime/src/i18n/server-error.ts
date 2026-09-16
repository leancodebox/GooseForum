"use client";

import { useCallback } from "react";
import { useTranslation } from "react-i18next";
import { GooseClientError } from "@gooseforum/client";

export function useServerErrorMessage() {
  const { t, i18n } = useTranslation("serverMessages");
  return useCallback(
    (reason: unknown, fallback: string) => {
      if (reason instanceof GooseClientError && reason.messageCode) {
        if (i18n.exists(reason.messageCode, { ns: "serverMessages" })) {
          return t(reason.messageCode, { ...(reason.params || {}) });
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
