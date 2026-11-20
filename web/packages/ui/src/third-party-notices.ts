import { existsSync, readdirSync, readFileSync, realpathSync } from "node:fs";
import { dirname, join } from "node:path";
import type { Plugin } from "vite";

export const noticesFile = "third-party-notices.txt";

const header = `THIRD-PARTY NOTICES

This file gives the licence notices of the third-party software in this web UI. Each entry gives the location of
the source code of the software, at the version in this web UI. This web UI contains the software with no changes.
You can get the source code at that location at no charge.
`;

const separator = "=".repeat(80);

// The same file names as tools/notices: LICENSE.md and NOTICE.txt, but not a code file such as notice.js.
const licenceFile = /^(licen[cs]e|copying|notice|patents)([-.](txt|md|mit|apache|bsd|[\d.]+))*$/i;

const nodeModules = "/node_modules/";

export function thirdPartyNotices(): Plugin {
  let root = "";

  return {
    name: "cdr-lab:third-party-notices",
    apply: "build",
    configResolved(config) {
      root = config.root;
    },
    generateBundle(_options, bundle) {
      const files = Object.values(bundle);
      const ids = files.flatMap((file) =>
        file.type === "chunk"
          ? Object.entries(file.modules)
              .filter(([, info]) => info.renderedLength > 0)
              .map(([id]) => id)
          : [],
      );
      const stylesheets = files.flatMap((file) =>
        file.type === "asset" && file.fileName.endsWith(".css") ? [String(file.source)] : [],
      );
      this.emitFile({
        type: "asset",
        fileName: noticesFile,
        source: notices(bundledPackages(ids, stylesheets, root)),
      });
    },
  };
}

export function bundledPackages(
  moduleIds: string[],
  stylesheets: string[],
  root: string,
): string[] {
  const dirs = new Set<string>();
  for (const id of moduleIds) {
    const at = id.lastIndexOf(nodeModules);
    if (at !== -1) {
      const [first = "", second = ""] = id.slice(at + nodeModules.length).split("/");
      const name = first.startsWith("@") ? `${first}/${second}` : first;
      dirs.add(id.slice(0, at + nodeModules.length) + name);
    } else if (id === "\0rolldown/runtime.js") {
      // The helpers that Rolldown adds to a bundle, through Vite.
      dirs.add(findPackage("rolldown", findPackage("vite", root)));
    } else if (id.startsWith("\0")) {
      throw new Error(`the bundle has the virtual module ${id}, which is not in a known package`);
    }
  }
  // Tailwind CSS puts its banner at the start of its CSS in a stylesheet, for example the preflight styles.
  if (stylesheets.some((css) => css.includes("tailwindcss v"))) {
    dirs.add(findPackage("tailwindcss", findPackage("@tailwindcss/vite", root)));
  }
  return [...dirs].toSorted();
}

// findPackage finds a package as Node does, in node_modules of the directory or of a parent directory.
function findPackage(name: string, from: string): string {
  for (let dir = from; ; dir = dirname(dir)) {
    const candidate = join(dir, "node_modules", name);
    if (existsSync(join(candidate, "package.json"))) return realpathSync(candidate);
    if (dirname(dir) === dir) throw new Error(`cannot find the package ${name} from ${from}`);
  }
}

function readManifest(dir: string): { name: string; version: string } {
  const manifest: unknown = JSON.parse(readFileSync(join(dir, "package.json"), "utf8"));
  if (
    typeof manifest === "object" &&
    manifest !== null &&
    "name" in manifest &&
    "version" in manifest &&
    typeof manifest.name === "string" &&
    typeof manifest.version === "string"
  ) {
    return { name: manifest.name, version: manifest.version };
  }
  throw new Error(`the package.json in ${dir} has no name or version`);
}

export function notices(packageDirs: string[]): string {
  const entries = packageDirs
    .map((dir) => {
      const { name, version } = readManifest(dir);
      const files = readdirSync(dir, { withFileTypes: true })
        .filter((f) => f.isFile() && licenceFile.test(f.name))
        .map((f) => f.name)
        .toSorted();
      if (files.length === 0)
        throw new Error(`the package ${name} ${version} has no licence file in ${dir}`);
      return { name, version, dir, files };
    })
    .toSorted((a, b) => (a.name < b.name ? -1 : 1));

  let text = header;
  for (const { name, version, dir, files } of entries) {
    const archive = `${name.slice(name.lastIndexOf("/") + 1)}-${version}.tgz`;
    text += `\n${separator}\n${name} ${version}\nSource: https://registry.npmjs.org/${name}/-/${archive}\n`;
    for (const file of files) {
      const content = readFileSync(join(dir, file), "utf8");
      text += `\n--- ${file} ---\n${content}${content.endsWith("\n") ? "" : "\n"}`;
    }
  }
  return text;
}
