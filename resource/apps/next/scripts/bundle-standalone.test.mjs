import assert from "node:assert/strict";
import { mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { buildArchive } from "./bundle-standalone.mjs";

test("adds pnpm aliases and excludes the unused Sharp runtime", async () => {
  const temporary = await mkdtemp(path.join(os.tmpdir(), "goose-next-bundle-"));
  const source = path.join(temporary, "standalone");
  const write = async (name) => {
    const target = path.join(source, name);
    await mkdir(path.dirname(target), { recursive: true });
    await writeFile(target, name);
  };

  try {
    await write("apps/next/start.mjs");
    await write("node_modules/.pnpm/node_modules/@swc/helpers/index.js");
    await write("node_modules/.pnpm/sharp@1.0.0/index.js");
    const output = path.join(temporary, "standalone.zip");

    await buildArchive(source, output);
    const archive = (await readFile(output)).toString("latin1");
    assert.match(archive, /apps\/next\/start\.mjs/);
    assert.match(archive, /node_modules\/@swc\/helpers\/index\.js/);
    assert.doesNotMatch(archive, /node_modules\/\.pnpm\/sharp@1\.0\.0/);
  } finally {
    await rm(temporary, { recursive: true, force: true });
  }
});
