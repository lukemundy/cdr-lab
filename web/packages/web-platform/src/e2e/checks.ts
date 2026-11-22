import { AxeBuilder } from "@axe-core/playwright";
import { expect, type Locator, type Page } from "@playwright/test";

// The same tags as the component tests of the design system.
const axeTags = ["wcag2a", "wcag2aa", "wcag21a", "wcag21aa", "wcag22aa"];

export async function expectNoAxeViolations(page: Page) {
  // A fade gives a lower contrast, and a toast can start to fade after the wait, so a check that fails runs again.
  await expect(async () => {
    await waitForAnimations(page);
    const { violations } = await new AxeBuilder({ page }).withTags(axeTags).analyze();
    expect(
      violations.map(({ id, nodes }) => ({ id, nodes: nodes.map((node) => node.html) })),
    ).toEqual([]);
  }).toPass();
}

// A transition starts one frame after its element appears, so the wait starts two frames later.
async function waitForAnimations(page: Page) {
  await page.evaluate(async () => {
    await new Promise((resolve) => {
      requestAnimationFrame(() => requestAnimationFrame(resolve));
    });
    const finite = document
      .getAnimations()
      .filter((animation) => animation.effect?.getTiming().iterations !== Infinity);
    await Promise.allSettled(finite.map((animation) => animation.finished));
  });
}

// The value of the term in a description list (dl) in within.
export function definitionOf(within: Page | Locator, term: string): Locator {
  return within
    .locator("dt", { hasText: new RegExp(`^${term}$`) })
    .locator("xpath=following-sibling::dd[1]");
}
