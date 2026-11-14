export function appendUniqueById<T extends { id: number }>(
  current: T[],
  incoming: T[],
): T[] {
  const seen = new Set(current.map((item) => item.id));
  return [
    ...current,
    ...incoming.filter((item) => {
      if (seen.has(item.id)) return false;
      seen.add(item.id);
      return true;
    }),
  ];
}

export function prependUniqueById<T extends { id: number }>(
  current: T[],
  incoming: T[],
): T[] {
  const seen = new Set(current.map((item) => item.id));
  return [
    ...incoming.filter((item) => {
      if (seen.has(item.id)) return false;
      seen.add(item.id);
      return true;
    }),
    ...current,
  ];
}
