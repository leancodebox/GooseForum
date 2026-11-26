export function searchUrl(query: string, sort = "relevance", page = 1) {
  const params = new URLSearchParams();
  const keyword = query.trim();
  if (keyword) params.set("q", keyword);
  if (sort !== "relevance") params.set("sort", sort);
  if (page > 1) params.set("page", String(page));
  return `/search${params.size ? `?${params}` : ""}`;
}
