import { applyLaunchOptions } from "./launch-options.mjs";

if (applyLaunchOptions()) await import("./server.js");
