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
const missing = [];
let localLinks = 0;

for (const file of markdownFiles) {
  const source = fs.readFileSync(file, "utf8");
  for (const match of source.matchAll(linkPattern)) {
    const raw = match[1].trim();
    if (!raw || raw.startsWith("#") || /^[a-z][a-z0-9+.-]*:/i.test(raw)) continue;
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
  process.exit(1);
}
console.log(`Documentation link check passed: ${markdownFiles.length} files, ${localLinks} local links.`);
