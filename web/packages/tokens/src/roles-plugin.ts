import type { Plugin } from "@terrazzo/parser";
import { contrastPairs } from "./contrast.ts";
import { colorRoles, cssVariable } from "./tokens.ts";

// roles.html loads only tokens.css and fonts.css, as a Go page or an app does.
export default function rolesPlugin(): Plugin {
  return {
    name: "cdr-lab:roles",
    build({ resolver, outputFile }) {
      outputFile("roles.html", page(colorRoles(resolver.apply({}))));
    },
  };
}

function slug(id: string) {
  return id.replace(/^color\./, "");
}

function page(roles: string[]) {
  const rules = roles.map(
    (id) =>
      `    .bg-${slug(id)} { background: var(${cssVariable(id)}); }\n    .fg-${slug(id)} { color: var(${cssVariable(id)}); }\n    .bd-${slug(id)} { border-color: var(${cssVariable(id)}); }`,
  );
  const swatches = roles.map(
    (id) => `      <tr>
        <th scope="row"><code>${id}</code></th>
        <td><span class="swatch bg-${slug(id)}"></span></td>
      </tr>`,
  );
  const pairs = contrastPairs.map(({ fg, bg, min }) => {
    const sample =
      min === 3
        ? `<span class="boundary bd-${slug(fg)}"></span>`
        : `<span class="fg-${slug(fg)}">Aa 0123</span>`;
    return `      <tr>
        <td><code>${fg}</code></td>
        <td><code>${bg}</code></td>
        <td class="bg-${slug(bg)}">${sample}</td>
        <td>${min}:1</td>
      </tr>`;
  });
  return `<!doctype html>
<html lang="en-AU">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Colour roles</title>
  <link rel="stylesheet" href="tokens.css">
  <link rel="stylesheet" href="fonts.css">
  <style>
    body { margin: 0; padding: var(--ds-space-6); background: var(--ds-color-page); color: var(--ds-color-fg); font-family: var(--ds-font-body); font-size: var(--ds-font-size-base); line-height: var(--ds-line-height-normal); }
    table { border-collapse: collapse; margin-block-end: var(--ds-space-8); background: var(--ds-color-surface); box-shadow: var(--ds-elevation-1); border-radius: var(--ds-radius-md); }
    th, td { padding: var(--ds-space-cell-y) var(--ds-space-cell-x); border-bottom: 1px solid var(--ds-color-border); text-align: start; }
    code { font-family: var(--ds-font-code); font-size: var(--ds-font-size-sm); }
    .swatch { display: inline-block; inline-size: var(--ds-space-12); block-size: var(--ds-space-6); border: 1px solid var(--ds-color-border-strong); border-radius: var(--ds-radius-sm); vertical-align: middle; }
    .boundary { display: inline-block; inline-size: var(--ds-space-12); block-size: var(--ds-space-6); border: 2px solid; border-radius: var(--ds-radius-sm); vertical-align: middle; }
    td span { font-variant-numeric: tabular-nums; font-weight: var(--ds-font-weight-semibold); }
${rules.join("\n")}
  </style>
</head>
<body>
  <h1>Colour roles</h1>
  <p>The colour roles of the default brand. The page has no script. It loads only <code>tokens.css</code> and <code>fonts.css</code>.</p>
  <table>
    <caption>Roles</caption>
    <thead>
      <tr><th scope="col">Role</th><th scope="col">Swatch</th></tr>
    </thead>
    <tbody>
${swatches.join("\n")}
    </tbody>
  </table>
  <table>
    <caption>Pairs that the contrast test checks</caption>
    <thead>
      <tr><th scope="col">Foreground</th><th scope="col">Background</th><th scope="col">Sample</th><th scope="col">Minimum</th></tr>
    </thead>
    <tbody>
${pairs.join("\n")}
    </tbody>
  </table>
</body>
</html>
`;
}
