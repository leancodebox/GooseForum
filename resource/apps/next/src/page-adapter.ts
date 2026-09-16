import { cache } from "react";
import {
  createGooseClient,
  type AnyPagePayload,
} from "@gooseforum/client";
import { createDemoPage } from "./demo-payload";

export interface NextPageRequest {
  path: string;
  cookie?: string;
  acceptLanguage?: string;
}

export interface NextPageAdapter {
  readonly mode: "demo" | "upstream";
  load(request: NextPageRequest): Promise<AnyPagePayload>;
}

export function createNextPageAdapter(
  origin = process.env.GOOSEFORUM_ORIGIN,
): NextPageAdapter {
  if (!origin) {
    return {
      mode: "demo",
      load(request) {
        return Promise.resolve(createDemoPage(request.path));
      },
    };
  }

  const baseURL = normalizeOrigin(origin);
  return {
    mode: "upstream",
    async load(request) {
      if (!request.path.startsWith("/")) {
        throw new Error("Next page adapter paths must be root-relative");
      }
      const client = createGooseClient<AnyPagePayload>({ baseURL });
      const headers = new Headers();
      if (request.cookie) headers.set("Cookie", request.cookie);
      if (request.acceptLanguage)
        headers.set("Accept-Language", request.acceptLanguage);
      return client.pages.fetch(request.path, {
        cache: "no-store",
        headers,
      });
    },
  };
}

export const loadNextPage = cache(
  async (path: string, cookie = "", acceptLanguage = "") =>
    createNextPageAdapter().load({ path, cookie, acceptLanguage }),
);

function normalizeOrigin(value: string) {
  const url = new URL(value);
  if (url.protocol !== "http:" && url.protocol !== "https:") {
    throw new Error("GOOSEFORUM_ORIGIN must use http or https");
  }
  return url.origin;
}
