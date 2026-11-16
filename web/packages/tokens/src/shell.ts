import Color from "colorjs.io";

// The colours of the console shell, each a mix of two roles (web/packages/ui/design/layout.md, UI-0083). theme.css
// gives each one as a Tailwind colour.
export interface ShellMix {
  id: string;
  role: string;
  percent: number;
  base: string;
}

export const shellMixes: ShellMix[] = [
  { id: "shell.nav-hover", role: "color.fg-on-hero", percent: 8, base: "color.hero" },
  { id: "shell.nav-current", role: "color.fg-on-hero", percent: 14, base: "color.hero" },
  { id: "shell.band", role: "color.accent", percent: 10, base: "color.surface" },
  { id: "shell.page", role: "color.hero", percent: 3, base: "color.page" },
];

const variable = (id: string) => `var(--ds-${id.replaceAll(".", "-")})`;

export function shellMixCSS({ id, role, percent, base }: ShellMix) {
  return `--color-${id.replace(".", "-")}: color-mix(in oklab, ${variable(role)} ${percent}%, ${variable(base)});`;
}

// The roles that the brand of a participant gives to its console (web/packages/ui/design/tokens-and-brands.md,
// UI-0035). The data holder serves the same list (services/dataholder/internal/adapter/branddir, consoleProperties).
export const participantConsoleRoles = [
  "color.brand",
  "color.fg-on-brand",
  "color.hero",
  "color.fg-on-hero",
  "color.accent",
  "color.heading",
  "color.link",
  "color.primary",
  "color.primary-hover",
  "color.primary-pressed",
  "color.fg-on-primary",
];

// consoleColor gives a role or a shell colour as the browser computes it: color-mix() in OKLab.
export function consoleColor(id: string, role: (id: string) => Color): Color {
  const mix = shellMixes.find((m) => m.id === id);
  if (!mix) return role(id);
  return role(mix.base).mix(role(mix.role), mix.percent / 100, {
    space: "oklab",
    outputSpace: "srgb",
  });
}
