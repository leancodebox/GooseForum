import { copyFile, cp, mkdir } from "node:fs/promises";
import path from "node:path";

const appRoot = path.resolve(import.meta.dirname, "..");
const target = path.join(
  appRoot,
  ".next/standalone/apps/next/.next/static",
);

await mkdir(path.dirname(target), { recursive: true });
await cp(path.join(appRoot, ".next/static"), target, { recursive: true });
const standaloneRoot = path.join(appRoot, ".next/standalone/apps/next");
await Promise.all([
  copyFile(
    path.join(appRoot, "scripts/standalone-entry.mjs"),
    path.join(standaloneRoot, "start.mjs"),
  ),
  copyFile(
    path.join(appRoot, "scripts/launch-options.mjs"),
    path.join(standaloneRoot, "launch-options.mjs"),
  ),
]);

try {
  await cp(
    path.join(appRoot, "public"),
    path.join(standaloneRoot, "public"),
    { recursive: true },
  );
} catch (error) {
  if (error?.code !== "ENOENT") throw error;
}
