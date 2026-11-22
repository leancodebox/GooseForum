const graphemes = typeof Intl.Segmenter === "function"
  ? new Intl.Segmenter(undefined, { granularity: "grapheme" })
  : undefined;

export function avatarInitial(nickname?: string, username?: string) {
  const name = nickname?.trim() || username?.trim() || "";
  const first = graphemes
    ? graphemes.segment(name)[Symbol.iterator]().next().value?.segment
    : name[Symbol.iterator]().next().value;
  return first?.toUpperCase() || "";
}
