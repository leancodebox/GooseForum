import path from "node:path";

/** @type {import('next').NextConfig} */
const config = {
  output: "standalone",
  outputFileTracingRoot: path.join(import.meta.dirname, "../.."),
  transpilePackages: [
    "@gooseforum/client",
    "@gooseforum/runtime",
    "@gooseforum/theme-default",
    "@gooseforum/ui",
  ],
  turbopack: {
    root: path.join(import.meta.dirname, "../.."),
  },
};

export default config;
