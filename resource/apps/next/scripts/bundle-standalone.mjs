import { ZipArchive } from "archiver";
import { createWriteStream } from "node:fs";
import {
  mkdir,
  readdir,
  realpath,
  rename,
  rm,
  stat,
} from "node:fs/promises";
import path from "node:path";
import { pathToFileURL } from "node:url";

const archiveDate = new Date("1980-01-01T00:00:00.000Z");
const appRoot = path.resolve(import.meta.dirname, "..");

export function excludedRuntimePath(name) {
  const normalized = `/${name.split(path.sep).join("/")}`;
  return (
    normalized.includes("/node_modules/.pnpm/sharp@") ||
    normalized.includes("/node_modules/.pnpm/@img+") ||
    normalized.includes("/node_modules/sharp/") ||
    normalized.includes("/node_modules/@img/")
  );
}

async function addPath(archive, source, name, ancestors = new Set()) {
  if (excludedRuntimePath(name)) return;

  const [info, resolvedSource] = await Promise.all([stat(source), realpath(source)]);
  if (info.isDirectory()) {
    if (ancestors.has(resolvedSource)) {
      throw new Error(`symlink cycle at ${source}`);
    }
    const nextAncestors = new Set(ancestors).add(resolvedSource);
    const entries = await readdir(source, { withFileTypes: true });
    entries.sort((left, right) => left.name.localeCompare(right.name));
    for (const entry of entries) {
      await addPath(
        archive,
        path.join(source, entry.name),
        name ? `${name}/${entry.name}` : entry.name,
        nextAncestors,
      );
    }
    return;
  }
  if (!info.isFile()) throw new Error(`unsupported file type ${source}`);

  archive.file(source, {
    name: name.split(path.sep).join("/"),
    date: archiveDate,
    mode: info.mode,
  });
}

async function addPnpmAliases(archive, root) {
  const aliases = path.join(root, "node_modules", ".pnpm", "node_modules");
  let entries;
  try {
    entries = await readdir(aliases, { withFileTypes: true });
  } catch (error) {
    if (error?.code === "ENOENT") return;
    throw error;
  }
  entries.sort((left, right) => left.name.localeCompare(right.name));
  for (const entry of entries) {
    await addPath(
      archive,
      path.join(aliases, entry.name),
      `node_modules/${entry.name}`,
    );
  }
}

export async function buildArchive(source, output) {
  const root = path.resolve(source);
  const sourceInfo = await stat(root).catch(() => null);
  if (!sourceInfo?.isDirectory()) {
    throw new Error(`Next standalone source "${root}" is unavailable`);
  }

  const target = path.resolve(output);
  const temporary = `${target}.tmp`;
  await mkdir(path.dirname(target), { recursive: true });
  await rm(temporary, { force: true });

  const destination = createWriteStream(temporary);
  const archive = new ZipArchive({ zlib: { level: 9 } });
  const completed = new Promise((resolve, reject) => {
    destination.once("close", resolve);
    destination.once("error", reject);
    archive.once("error", reject);
    archive.on("warning", (error) => {
      if (error.code !== "ENOENT") reject(error);
    });
  });
  archive.pipe(destination);

  try {
    await addPath(archive, root, "");
    await addPnpmAliases(archive, root);
    await archive.finalize();
    await completed;
    await rm(target, { force: true });
    await rename(temporary, target);
  } catch (error) {
    archive.abort();
    destination.destroy();
    await rm(temporary, { force: true });
    throw new Error(`archive Next standalone: ${error.message}`, {
      cause: error,
    });
  }
}

const invokedDirectly =
  process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href;
if (invokedDirectly) {
  const source = path.join(appRoot, ".next", "standalone");
  const output = path.resolve(appRoot, "..", "..", "next", "standalone.zip");
  await buildArchive(source, output);
  console.log(`Next standalone archive: ${output}`);
}
