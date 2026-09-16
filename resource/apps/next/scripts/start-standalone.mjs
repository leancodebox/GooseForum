import { pathToFileURL } from "node:url";
import path from "node:path";
import { applyLaunchOptions } from "./launch-options.mjs";

const server = path.resolve(
  import.meta.dirname,
  "../.next/standalone/apps/next/server.js",
);
if (applyLaunchOptions()) await import(pathToFileURL(server).href);
