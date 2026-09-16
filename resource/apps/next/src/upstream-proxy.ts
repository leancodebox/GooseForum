import type { NextRequest } from "next/server";

export const proxyToGooseForum = async (request: NextRequest) => {
  const origin = process.env.GOOSEFORUM_ORIGIN;
  if (!origin) {
    return Response.json(
      { error: "GOOSEFORUM_ORIGIN is not configured" },
      { status: 503 },
    );
  }
  const upstream = new URL(origin);
  if (upstream.protocol !== "http:" && upstream.protocol !== "https:") {
    return Response.json({ error: "Invalid GOOSEFORUM_ORIGIN" }, { status: 500 });
  }
  const target = new URL(
    `${request.nextUrl.pathname}${request.nextUrl.search}`,
    upstream.origin,
  );
  const headers = new Headers(request.headers);
  headers.delete("host");
  headers.delete("content-length");
  stripHopByHopHeaders(headers);
  headers.set("X-Forwarded-Host", request.nextUrl.host);
  headers.set("X-Forwarded-Proto", request.nextUrl.protocol.replace(":", ""));

  const response = await fetch(target, {
    method: request.method,
    headers,
    body:
      request.method === "GET" || request.method === "HEAD"
        ? undefined
        : await request.arrayBuffer(),
    cache: "no-store",
    redirect: "manual",
  });
  const responseHeaders = new Headers(response.headers);
  const setCookies = response.headers.getSetCookie();
  responseHeaders.delete("content-length");
  responseHeaders.delete("content-encoding");
  stripHopByHopHeaders(responseHeaders);
  if (setCookies.length) {
    responseHeaders.delete("set-cookie");
    for (const cookie of setCookies) responseHeaders.append("set-cookie", cookie);
  }
  rewriteLocation(responseHeaders, upstream.origin);
  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers: responseHeaders,
  });
};

function rewriteLocation(headers: Headers, upstreamOrigin: string) {
  const location = headers.get("location");
  if (!location) return;
  const url = new URL(location, upstreamOrigin);
  if (url.origin === upstreamOrigin)
    headers.set("location", `${url.pathname}${url.search}${url.hash}`);
}

function stripHopByHopHeaders(headers: Headers) {
  for (const name of [
    "connection",
    "keep-alive",
    "proxy-authenticate",
    "proxy-authorization",
    "te",
    "trailer",
    "transfer-encoding",
    "upgrade",
  ])
    headers.delete(name);
}
