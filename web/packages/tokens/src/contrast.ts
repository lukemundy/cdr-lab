export interface ContrastPair {
  fg: string;
  bg: string;
  min: number;
}

const backgrounds = ["color.page", "color.surface", "color.raised"];
const statuses = ["success", "warning", "danger", "info"];

const pairs = (fgs: string[], bgs: string[], min: number) =>
  fgs.flatMap((fg) => bgs.map((bg): ContrastPair => ({ fg, bg, min })));

// WCAG 2.2 SC 1.4.3. The primary action colour is not a text colour: a link uses the link role.
const text = [
  ...pairs(["color.fg", "color.fg-muted", "color.heading", "color.link"], backgrounds, 4.5),
  ...pairs(
    ["color.fg-on-primary"],
    ["color.primary", "color.primary-hover", "color.primary-pressed"],
    4.5,
  ),
  ...pairs(["color.fg", "color.heading"], ["color.secondary"], 4.5),
  ...pairs(["color.fg-on-brand"], ["color.brand"], 4.5),
  ...pairs(["color.fg-on-hero"], ["color.hero"], 4.5),
  ...statuses.flatMap((status) => pairs([`color.${status}-fg`], [`color.${status}-bg`], 4.5)),
];

// WCAG 2.2 SC 1.4.11. The border role, the status borders and the accent are decorative. A button with a text label
// needs no boundary, so the primary action colour is not here.
const nonText = pairs(["color.border-strong", "color.focus"], backgrounds, 3);

export const contrastPairs = [...text, ...nonText];

const shell = ["shell.band", "shell.page"];

// The pairs of the console shell, where a role of the brand of a participant can meet a role of the lab brand
// (web/packages/ui/DESIGN.md, Tokens).
export const consoleShellPairs: ContrastPair[] = [
  ...pairs(["color.fg-on-hero"], ["color.hero", "shell.nav-hover", "shell.nav-current"], 4.5),
  ...pairs(["color.fg", "color.fg-muted", "color.heading", "color.link"], shell, 4.5),
  ...pairs(["color.heading", "color.link"], backgrounds, 4.5),
  ...pairs(["color.fg-on-brand"], ["color.brand"], 4.5),
  ...pairs(
    ["color.fg-on-primary"],
    ["color.primary", "color.primary-hover", "color.primary-pressed"],
    4.5,
  ),
  // The focus ring of the side navigation is the text on the hero.
  ...pairs(["color.border-strong", "color.focus"], shell, 3),
];
