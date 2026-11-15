import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..");
const docsRoot = path.join(repoRoot, "docs");
const markdownFiles = fs.readdirSync(docsRoot)
  .filter((name) => name.endsWith(".md"))
  .map((name) => path.join(docsRoot, name));
const linkPattern = /\[[^\]]*\]\(([^)]+)\)/g;
// Code fences contain syntax examples, not links in the rendered document.
export function localLinkTargets(source) {
  let fence;
  const prose = source.split(/\r?\n/).filter(line => {
    const marker = line.match(/^ {0,3}(`{3,}|~{3,})(.*)$/);
    if (fence) {
      if (marker && marker[1][0] === fence[0] && marker[1].length >= fence.length && !marker[2].trim()) fence = undefined;
      return false;
    }
    if (marker) { fence = marker[1]; return false; }
    return true;
  }).join("\n");
  return Array.from(prose.matchAll(linkPattern), match => match[1].trim())
    .filter(raw => raw && !raw.startsWith("#") && !/^[a-z][a-z0-9+.-]*:/i.test(raw));
}

if (import.meta.main) {
  const missing = [];
  let localLinks = 0;
  for (const file of markdownFiles) {
    for (const raw of localLinkTargets(fs.readFileSync(file, "utf8"))) {
      const target = raw.split("#", 1)[0];
      if (!target) continue;
      localLinks++;
      const resolved = path.resolve(path.dirname(file), target);
      if (!fs.existsSync(resolved)) missing.push(`${path.relative(repoRoot, file)} -> ${raw}`);
    }
  }
  if (missing.length) {
    console.error(`Documentation link check failed: ${missing.length} missing target(s)`);
    for (const item of missing) console.error(`- ${item}`);
    process.exitCode = 1;
  } else {
    console.log(`Documentation link check passed: ${markdownFiles.length} files, ${localLinks} local links.`);
  }
}
