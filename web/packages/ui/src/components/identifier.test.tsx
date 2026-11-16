import { afterEach, expect, test, vi } from "vitest";
import { page } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize } from "../test/checks";
import { render } from "../test/render";
import { Identifier } from "./identifier";

const id = "a8f1c3d2-6b1e-4c5a-9f0e-2d7b3c4e5f60";

afterEach(() => {
  vi.restoreAllMocks();
});

// Verifies: UI-0162, UI-0257
test("the identifier shows the full value in the monospace font, with a copy button", async () => {
  const container = render(<Identifier value={id} label="legal entity ID" />);

  const code = page.getByText(id);
  await expect.element(code).toBeVisible();
  expect(code.element().tagName).toBe("CODE");
  expect(getComputedStyle(code.element()).fontFamily).toMatch(/JetBrains Mono/);
  const button = page.getByRole("button", { name: "Copy legal entity ID" });
  expectTargetSize(button.element());
  await expectNoAxeViolations(container);
});

// Verifies: UI-0259
test("a long identifier with no break points wraps in a narrow container", () => {
  const value = "x".repeat(200);
  const container = render(
    <div style={{ width: "200px" }}>
      <Identifier value={value} label="URI" />
    </div>,
  );

  const box = container.firstElementChild;
  if (!box) throw new Error("no container");
  expect(box.scrollWidth).toBeLessThanOrEqual(box.clientWidth);
});

// Verifies: UI-0259
test("with whitespace-nowrap, the identifier stays on one line", () => {
  const container = render(
    <div style={{ width: "100px" }}>
      <Identifier value={id} label="legal entity ID" className="whitespace-nowrap" />
    </div>,
  );

  const code = container.querySelector("code");
  if (!code) throw new Error("no code element");
  expect(code.getBoundingClientRect().height).toBeLessThanOrEqual(
    Number.parseFloat(getComputedStyle(code).lineHeight) * 1.5,
  );
});

// Verifies: UI-0258
test("with a display value, the identifier shows that value and the copy button copies the value", async () => {
  const writeText = vi.spyOn(navigator.clipboard, "writeText").mockResolvedValue();
  const container = render(<Identifier value="062000" display="062-000" label="BSB" />);

  const code = page.getByText("062-000", { exact: true });
  await expect.element(code).toBeVisible();
  expect(code.element().tagName).toBe("CODE");
  await page.getByRole("button", { name: "Copy BSB" }).click();

  expect(writeText).toHaveBeenCalledWith("062000");
  await expect.element(page.getByRole("status")).toHaveTextContent("Copied");
  await expectNoAxeViolations(container);
});
