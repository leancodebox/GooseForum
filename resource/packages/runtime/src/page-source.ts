import type { AnyPagePayload, GooseSiteApi, PagePayload } from "@gooseforum/client";

export interface PageSource<TPage extends PagePayload = AnyPagePayload> {
  api: GooseSiteApi;
  readInitial?(): TPage | undefined;
  load(url: URL, signal?: AbortSignal): Promise<TPage>;
}
