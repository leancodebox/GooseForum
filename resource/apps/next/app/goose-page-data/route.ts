import { NextResponse, type NextRequest } from "next/server";
import { loadNextPage } from "@/page-adapter";

export async function GET(request: NextRequest) {
  const path = request.nextUrl.searchParams.get("path") || "/";
  const page = await loadNextPage(
    path.startsWith("/") ? path : `/${path}`,
    request.headers.get("cookie") || "",
    request.headers.get("accept-language") || "",
  );
  return NextResponse.json(page, {
    headers: { "Cache-Control": "private, no-store" },
  });
}
