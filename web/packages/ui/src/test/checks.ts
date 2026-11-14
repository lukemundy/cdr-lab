import axe from "axe-core";
import { expect } from "vitest";

// DESIGN.md, Accessibility.
const axeTags = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];

// A fade gives a lower contrast while it runs, so axe waits for each animation that ends, which a busy spinner does not.
async function transitionsEnd(): Promise<void> {
  const finite = document
    .getAnimations()
    .filter((animation) => animation.effect?.getTiming().iterations !== Infinity);
  // A transition that stops before its end rejects its promise.
  await Promise.allSettled(finite.map((animation) => animation.finished));
}

export async function expectNoAxeViolations(context: Element): Promise<void> {
  await transitionsEnd();
  const { violations } = await axe.run(context, { runOnly: { type: "tag", values: axeTags } });
  expect(
    violations.map(({ id, nodes }) => ({ id, nodes: nodes.map((node) => node.html) })),
  ).toEqual([]);
}

export function expectVisibleFocus(element: Element): void {
  expect(document.activeElement).toBe(element);
  expect(element.matches(":focus-visible")).toBe(true);
  const style = getComputedStyle(element);
  expect(style.outlineStyle).not.toBe("none");
  expect(Number.parseFloat(style.outlineWidth)).toBeGreaterThanOrEqual(2);
}

// DESIGN.md, Layout: in a console, each control is 24 by 24 CSS px or more.
export function expectTargetSize(element: Element): void {
  const { width, height } = element.getBoundingClientRect();
  expect(width).toBeGreaterThanOrEqual(24);
  expect(height).toBeGreaterThanOrEqual(24);
}
