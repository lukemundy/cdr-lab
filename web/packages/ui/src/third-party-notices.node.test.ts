import { mkdir, mkdtemp, realpath, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { beforeAll, describe, expect, test } from "vitest";

import { bundledPackages, notices } from "./third-party-notices";

let workspace = "";
let app = "";

async function addPackage(
  dir: string,
  name: string,
  version: string,
  files: Record<string, string>,
) {
  const root = join(dir, "node_modules", name);
  await mkdir(root, { recursive: true });
  await writeFile(join(root, "package.json"), JSON.stringify({ name, version }));
  await Promise.all(Object.entries(files).map(([file, text]) => writeFile(join(root, file), text)));
  return root;
}

// A pnpm layout: each package is in node_modules/.pnpm/<name>@<version>/node_modules, next to its dependencies.
beforeAll(async () => {
  workspace = await realpath(await mkdtemp(join(tmpdir(), "notices-")));
  app = join(workspace, "apps", "app");
  await mkdir(app, { recursive: true });
  const store = join(workspace, "node_modules", ".pnpm");
  await addPackage(join(store, "react@19.0.0"), "react", "19.0.0", { LICENSE: "React licence\n" });
  await addPackage(join(store, "@scope+lib@1.0.0"), "@scope/lib", "1.0.0", {
    "LICENSE.md": "Lib licence",
    NOTICE: "Lib notice\n",
    "notice.js": "export {};\n",
  });
  await addPackage(join(store, "bare@1.0.0"), "bare", "1.0.0", {});
  await addPackage(workspace, "vite", "8.0.0", { LICENSE: "Vite licence\n" });
  await addPackage(join(workspace, "node_modules", "vite"), "rolldown", "1.0.0", {
    LICENSE: "Rolldown licence\n",
  });
  const plugin = await addPackage(workspace, "@tailwindcss/vite", "4.0.0", {
    LICENSE: "Plugin licence\n",
  });
  await addPackage(plugin, "tailwindcss", "4.0.0", { LICENSE: "Tailwind licence\n" });
});

const inStore = (pkg: string, name: string, file: string) =>
  join(workspace, "node_modules", ".pnpm", pkg, "node_modules", name, file);

describe("the packages of a bundle", () => {
  // Verifies: UI-0182
  test("are the packages of the bundled modules from node_modules, one time each", () => {
    const ids = [
      join(app, "src", "main.tsx"),
      inStore("react@19.0.0", "react", "index.js"),
      inStore("react@19.0.0", "react", "cjs/react.production.js?commonjs-exports"),
      inStore("@scope+lib@1.0.0", "@scope/lib", "dist/index.js"),
    ];

    expect(bundledPackages(ids, [], app)).toEqual([
      inStore("@scope+lib@1.0.0", "@scope/lib", ""),
      inStore("react@19.0.0", "react", ""),
    ]);
  });

  // Verifies: UI-0182
  test("include the bundler runtime and the Tailwind CSS that a stylesheet contains", () => {
    const css = "/*! tailwindcss v4.0.0 | MIT License | https://tailwindcss.com */\n*{margin:0}";

    expect(bundledPackages(["\0rolldown/runtime.js"], [css], app)).toEqual([
      join(workspace, "node_modules", "@tailwindcss", "vite", "node_modules", "tailwindcss"),
      join(workspace, "node_modules", "vite", "node_modules", "rolldown"),
    ]);
    expect(bundledPackages([], ["a{color:red}"], app)).toEqual([]);
  });

  test("cannot include a module that is not in a known package", () => {
    expect(() => bundledPackages(["\0virtual:unknown"], [], app)).toThrow("\0virtual:unknown");
  });
});

describe("the notices", () => {
  // Verifies: UI-0182
  test("give each package with its source archive and its licence files", () => {
    const text = notices([
      inStore("react@19.0.0", "react", ""),
      inStore("@scope+lib@1.0.0", "@scope/lib", ""),
    ]);

    expect(text).toMatch(
      /^THIRD-PARTY NOTICES\n\nThis file gives the licence notices of the third-party software in this web UI\./,
    );
    expect(text.slice(text.indexOf("=====")))
      .toBe(`================================================================================
@scope/lib 1.0.0
Source: https://registry.npmjs.org/@scope/lib/-/lib-1.0.0.tgz

--- LICENSE.md ---
Lib licence

--- NOTICE ---
Lib notice

================================================================================
react 19.0.0
Source: https://registry.npmjs.org/react/-/react-19.0.0.tgz

--- LICENSE ---
React licence
`);
  });

  // Verifies: UI-0182
  test("fail for a package with no licence file", () => {
    expect(() => notices([inStore("bare@1.0.0", "bare", "")])).toThrow(
      "bare 1.0.0 has no licence file",
    );
  });
});
