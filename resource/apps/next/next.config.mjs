import path from "node:path";

/** @type {import('next').NextConfig} */
const config = {
  output: "standalone",
  images: { unoptimized: true },
  outputFileTracingRoot: path.join(import.meta.dirname, "../.."),
  outputFileTracingExcludes: {
    "/*": [
      "../../node_modules/.pnpm/sharp@*/**/*",
      "../../node_modules/.pnpm/@img+*/**/*",
    ],
  },
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
