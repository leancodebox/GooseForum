const healthTokenHeader = "X-Goose-Health-Token";

export function GET() {
  const token = process.env.GOOSEFORUM_HEALTH_TOKEN;
  if (!token) return new Response(null, { status: 404 });

  return new Response(null, {
    status: 204,
    headers: {
      "Cache-Control": "private, no-store",
      [healthTokenHeader]: token,
    },
  });
}
