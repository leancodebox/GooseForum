import MarkdownIt from "markdown-it";
import type StateInline from "markdown-it/lib/rules_inline/state_inline.mjs";

export function parseExtensionHeader(input: string) {
  const source = input.slice(0, 1024);
  const name = /^\[([a-z][a-z0-9-]*)/.exec(source);
  if (!name) return null;
  const attributes: Record<string, string> = {};
  let position = name[0].length;
  while (position < source.length && position < 1024) {
    if (source[position] === "]")
      return { name: name[1], attributes, end: position + 1 };
    if (source[position] !== " ") return null;
    while (source[position] === " ") position++;
    if (source[position] === "]") continue;
    const attr = /^([a-z][a-z0-9-]*)="((?:[^"\\]|\\["\\\]])*)"/.exec(source.slice(position));
    if (!attr || Object.hasOwn(attributes, attr[1]) || [...attr[2]].some((character) => character.charCodeAt(0) < 32 || character.charCodeAt(0) === 127)) return null;
    attributes[attr[1]] = attr[2].replace(/\\(["\\\]])/g, "$1");
    position += attr[0].length;
  }
  return null;
}

export const mentionMarkdown = (id: string, username: string) =>
  `[mention user="${id}"]@${username}[/mention]`;

const inlineContexts = new WeakMap<object, { pairs: Map<number, number>; html: Array<[number, number]> }>();
const literalMarkdown = new MarkdownIt({ html: false });
const voidTags = new Set("area base br col embed hr img input link meta param source track wbr".split(" "));

// Use the existing tokenizer to skip code, links and escapes, once per block.
function inlineContext(state: StateInline) {
  const cached = inlineContexts.get(state);
  if (cached) return cached;
  const result = { pairs: new Map<number, number>(), html: [] as Array<[number, number]> };
  inlineContexts.set(state, result);
  const saved = state.pos;
  const stack: Array<{ name: string; start: number }> = [];
  let htmlDepth = 0;
  let htmlStart = 0;
  state.pos = 0;
  while (state.pos < state.posMax) {
    const start = state.pos;
    if (state.src[start] === "[") {
      const close = /^\[\/([a-z][a-z0-9-]*)\]/.exec(state.src.slice(start, start + 1024));
      const header = parseExtensionHeader(state.src.slice(start, start + 1024));
      if (close && stack.at(-1)?.name === close[1]) {
        const open = stack.pop()!;
        result.pairs.set(open.start, start + close[0].length);
      } else if (header && header.name !== "mention" && stack.length < 8 && state.src[start + header.end] !== "(") {
        stack.push({ name: header.name, start });
      }
    }
    if (state.src[start] === "<") {
      const tag = /^<\/?([A-Za-z][A-Za-z0-9-]*)(?:\s[^<>]*|)>/.exec(state.src.slice(start));
      if (tag) {
        if (tag[0].startsWith("</") && htmlDepth > 0) {
          if (--htmlDepth === 0) result.html.push([htmlStart, start + tag[0].length]);
        } else if (!tag[0].startsWith("</") && !tag[0].endsWith("/>") && !voidTags.has(tag[1].toLowerCase())) {
          if (htmlDepth++ === 0) htmlStart = start;
        }
      }
    }
    state.md.inline.skipToken(state);
    if (state.pos <= start) state.pos++;
  }
  if (htmlDepth > 0) result.html.push([htmlStart, state.posMax]);
  state.pos = saved;
  return result;
}

export function forumExtensions(md: MarkdownIt) {
  const blockContexts = new WeakMap<object, Map<number, number>>();
  md.block.ruler.before("paragraph", "forum_opaque", (state, start, end, silent) => {
    const environment = state.env as { sourceVersion?: number };
    if (environment?.sourceVersion === 0 || state.sCount[start] - state.blkIndent >= 4) return false;
    const line = state.src.slice(state.bMarks[start] + state.tShift[start], state.eMarks[start]);
    const header = parseExtensionHeader(line);
    if (!header || header.name === "mention" || line.slice(header.end).trim()) return false;
    let pairs = blockContexts.get(state);
    if (!pairs) {
      const codeLines = new Set<number>();
      for (const token of literalMarkdown.parse(state.src, {})) {
        if ((token.type === "fence" || token.type === "code_block") && token.map)
          for (let line = token.map[0]; line < token.map[1]; line++) codeLines.add(line);
      }
      pairs = new Map<number, number>();
      const stack: Array<{ name: string; line: number }> = [];
      for (let line = 0; line < state.lineMax; line++) {
        if (codeLines.has(line)) continue;
        const text = state.src.slice(state.bMarks[line] + state.tShift[line], state.eMarks[line]).trim();
        const opening = parseExtensionHeader(text);
        const closing = /^\[\/([a-z][a-z0-9-]*)\]$/.exec(text);
        if (closing && closing[1] === stack.at(-1)?.name) {
          pairs.set(stack.pop()!.line, line + 1);
        } else if (opening && opening.name !== "mention" && !text.slice(opening.end).trim() && stack.length < 8) {
          stack.push({ name: opening.name, line });
        }
      }
      blockContexts.set(state, pairs);
    }
    const next = pairs.get(start);
    if (!next || next > end) return false;
    if (silent) return true;
    const token = state.push("forum_opaque", "", 0);
    token.block = true;
    token.map = [start, next];
    token.content = state.getLines(start, next, state.blkIndent, false);
    state.line = next;
    return true;
  });
  md.renderer.rules.forum_opaque = (tokens, index) => `<p>${md.utils.escapeHtml(tokens[index].content)}</p>\n`;
  md.inline.ruler.before("link", "forum_mention", (state, silent) => {
    const environment = state.env as { sourceVersion?: number; references?: Record<string, unknown> };
    if (silent || environment?.sourceVersion === 0 || (state as StateInline & { linkLevel: number }).linkLevel > 0 || state.src[state.pos] !== "[") return false;
    const context = inlineContext(state);
    if (context.html.some(([start, end]) => state.pos >= start && state.pos < end)) return false;
    const source = state.src.slice(state.pos, state.pos + 1162);
    const header = parseExtensionHeader(source);
    if (!header || source[header.end] === "(") return false;
    if (source[header.end] === "[") {
      const reference = /^\[([^\]]*)\]/.exec(source.slice(header.end));
      if (reference && environment?.references?.[md.utils.normalizeReference(reference[1] || header.name)]) return false;
    }
    if (header.name !== "mention") {
      const end = context.pairs.get(state.pos);
      if (!end) return false;
      const token = state.push("text", "", 0);
      token.content = state.src.slice(state.pos, end);
      state.pos = end;
      return true;
    }
    if (Object.keys(header.attributes).length !== 1) return false;
    const id = header.attributes.user;
    if (!/^[1-9][0-9]*$/.test(id || "") || BigInt(id) > 18446744073709551615n) return false;
    const relative = source.slice(header.end, header.end + 138).indexOf("[/mention]");
    if (relative < 0) return false;
    const close = header.end + relative;
    const label = source.slice(header.end, close);
    if (new TextEncoder().encode(label).length > 128 || /[[\]<>\r\n]/.test(label)) return false;
    const token = state.push("mention", "a", 0);
    token.content = label;
    token.meta = { id, raw: source.slice(0, close + 10) };
    state.pos += close + 10;
    return true;
  });
  md.renderer.rules.mention = (tokens, index) => {
    const token = tokens[index];
    const { id } = token.meta as { id: string };
    return `<a class="mention" data-mention-user="${id}" href="/u/${id}">${md.utils.escapeHtml(token.content)}</a>`;
  };
}

export function mentionQuery(text: string) {
  const match = /(?:^|[^A-Za-z0-9_\-.@/:])@([A-Za-z0-9_-]{0,32})$/.exec(text);
  return match ? { query: match[1], length: match[1].length + 1 } : null;
}
