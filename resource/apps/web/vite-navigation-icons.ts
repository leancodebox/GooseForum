import { readFile } from "node:fs/promises";
import { createRequire } from "node:module";
import { dirname, join } from "node:path";
import iconImports from "lucide-react/dynamicIconImports";
import type { Plugin } from "vite";

const groupPrefix = "virtual:goose-navigation-icons/";

// Keep Lucide's dynamic lookup API, but package icon data by first letter.
// Individual dynamic entries otherwise split the static Lucide icons used by both shells.
export function navigationIconChunks(): Plugin {
  const groups = new Map<string, Array<keyof typeof iconImports>>();
  for (const name of Object.keys(iconImports) as Array<keyof typeof iconImports>) {
    const group = name[0];
    const names = groups.get(group) ?? [];
    names.push(name);
    groups.set(group, names);
  }

  return {
    name: "goose-navigation-icon-chunks",
    apply: "build",
    enforce: "pre",
    async buildStart() {
      const require = createRequire(import.meta.url);
      const packageRoot = dirname(require.resolve("lucide-react/dynamicIconImports"));
      this.emitFile({ type: "asset", name: "lucide-icons-LICENSE.txt", source: await readFile(join(packageRoot, "LICENSE")) });
    },
    transform(_code, id) {
      if (!id.endsWith("/lucide-react/dist/esm/dynamicIconImports.mjs")) return;
      const groupEntries = [...groups].map(([group, names]) => {
        const path = JSON.stringify(groupPrefix + group);
        return `[() => import(${path}), ${JSON.stringify(names)}]`;
      });
      return `export default Object.fromEntries([${groupEntries.join(",")}].flatMap(
        ([load, names]) => names.map(name => [name, async () => ({ __iconData: (await load()).default[name] })])
      ));`;
    },
    resolveId(id) {
      if (id.startsWith(groupPrefix)) return `\0${id}`;
    },
    async load(id) {
      if (!id.startsWith(`\0${groupPrefix}`)) return;
      const names = groups.get(id.slice(groupPrefix.length + 1));
      if (!names) return;
      const entries = await Promise.all(names.map(async name => [name, (await iconImports[name]()).__iconData]));
      return `export default ${JSON.stringify(Object.fromEntries(entries))};`;
    },
  };
}
