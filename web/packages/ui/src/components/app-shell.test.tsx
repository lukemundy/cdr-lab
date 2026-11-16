import { AppWindow, Landmark } from "lucide-react";
import { expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize, expectVisibleFocus } from "../test/checks";
import { render } from "../test/render";
import { AppShell, AppShellNavItem } from "./app-shell";

function renderShell(brand: { logo?: string; label?: string } = {}) {
  return render(
    <AppShell
      name="Register console"
      {...brand}
      navigationLabel="Main"
      navigation={
        <>
          <AppShellNavItem href="#data-recipients" icon={AppWindow} current>
            Data recipients
          </AppShellNavItem>
          <AppShellNavItem href="#data-holders" icon={Landmark}>
            Data holders
          </AppShellNavItem>
        </>
      }
    >
      <h1>Data recipients</h1>
      <button type="button">In the page</button>
    </AppShell>,
  );
}

test("the shell has a banner with the name, a navigation and a main area", async () => {
  const container = renderShell();

  await expect.element(page.getByRole("banner")).toHaveTextContent("Register console");
  await expect.element(page.getByRole("navigation", { name: "Main" })).toBeVisible();
  await expect.element(page.getByRole("main")).toMatchTextContent("Data recipients");
  await expectNoAxeViolations(container);
});

test("with no logo, the header has the name as text and not the brand colour", async () => {
  renderShell();

  const banner = page.getByRole("banner");
  await expect.element(banner).toHaveTextContent("Register console");
  expect(banner.element().querySelector("img")).toBeNull();
  expect(banner.element().className).not.toContain("bg-brand");
});

// Verifies: UI-0075
test("with a logo, the header shows the logo with the name as its alternative and the label on the brand colour", async () => {
  const container = renderShell({ logo: "/brand/logo.svg", label: "Operator console" });

  const banner = page.getByRole("banner");
  await expect
    .element(banner.getByRole("img", { name: "Register console" }))
    .toHaveAttribute("src", "/brand/logo.svg");
  await expect.element(banner).toHaveTextContent("Operator console");
  expect(banner.element().className).toContain("bg-brand");
  expect(banner.element().className).toContain("text-fg-on-brand");
  await expectNoAxeViolations(container);
});

// Verifies: UI-0075
test("the header has a rule in the accent colour, with and without a logo", async () => {
  renderShell();
  expect(page.getByRole("banner").element().className).toContain("border-accent");
});

// Verifies: UI-0076, UI-0081
test("the navigation is on the hero colour, and the main area on the shell page colour", async () => {
  const container = renderShell({ logo: "/brand/logo.svg", label: "Operator console" });

  const navigation = page.getByRole("navigation", { name: "Main" }).element();
  expect(navigation.className).toContain("bg-hero");
  expect(navigation.className).toContain("text-fg-on-hero");
  expect(page.getByRole("main").element().className).toContain("bg-shell-page");
  await expectNoAxeViolations(container);
});

// Verifies: UI-0078
test("the focus ring in the navigation is the text colour on the hero", async () => {
  renderShell();

  await userEvent.tab();
  await userEvent.tab();
  const item = page.getByRole("link", { name: "Data recipients" }).element();
  expectVisibleFocus(item);
  const style = getComputedStyle(item);
  expect(style.outlineColor).toBe(style.color);
});

// Verifies: UI-0077
test("the current item has its own background and a bar that does not use the accent", async () => {
  renderShell();

  const current = getComputedStyle(page.getByRole("link", { name: "Data recipients" }).element());
  const other = getComputedStyle(page.getByRole("link", { name: "Data holders" }).element());
  expect(current.backgroundColor).not.toBe(other.backgroundColor);
  expect(current.borderLeftWidth).toBe("3px");
  expect(current.borderLeftColor).toBe(current.color);
});

test("the navigation marks the current page", async () => {
  renderShell();

  await expect
    .element(page.getByRole("link", { name: "Data recipients" }))
    .toHaveAttribute("aria-current", "page");
  await expect
    .element(page.getByRole("link", { name: "Data holders" }))
    .not.toHaveAttribute("aria-current");
});

test("each navigation item shows its icon, which assistive technology does not announce", async () => {
  renderShell();
  await expect
    .element(page.getByRole("link", { name: "Data recipients", exact: true }))
    .toBeVisible();

  // The exact accessible name shows that the icon adds no text to the name.
  for (const name of ["Data recipients", "Data holders"]) {
    const icon = page.getByRole("link", { name, exact: true }).element().querySelector("svg");
    expect(icon?.getAttribute("aria-hidden")).toBe("true");
    expect(icon?.getBoundingClientRect().width).toBe(16);
  }
});

test("a navigation item with no icon shows only its text", async () => {
  render(
    <AppShell
      name="CDR Register"
      navigationLabel="Main"
      navigation={<AppShellNavItem href="#a">Signing keys</AppShellNavItem>}
    >
      <h1>Signing keys</h1>
    </AppShell>,
  );

  const link = page.getByRole("link", { name: "Signing keys" });
  await expect.element(link).toBeVisible();
  expect(link.element().querySelector("svg")).toBeNull();
});

test("the first Tab shows the skip link, and the skip link moves the focus to the main area", async () => {
  renderShell();

  await userEvent.tab();
  const skip = page.getByRole("link", { name: "Skip to content" });
  expectVisibleFocus(skip.element());
  await expect.element(skip).toBeVisible();
  await userEvent.keyboard("{Enter}");

  expect(document.activeElement).toBe(page.getByRole("main").element());
  await userEvent.tab();
  expect(document.activeElement).toBe(page.getByRole("button", { name: "In the page" }).element());
});

test("the keyboard reaches each navigation item with a visible focus", async () => {
  renderShell();

  await userEvent.tab();
  await userEvent.tab();
  expectVisibleFocus(page.getByRole("link", { name: "Data recipients" }).element());
  await userEvent.tab();
  expectVisibleFocus(page.getByRole("link", { name: "Data holders" }).element());
});

// Verifies: UI-0094
test("each navigation item is 24 by 24 CSS px or more", () => {
  renderShell();

  expectTargetSize(page.getByRole("link", { name: "Data recipients" }).element());
  expectTargetSize(page.getByRole("link", { name: "Data holders" }).element());
});

test("a router can give the element of a navigation item", async () => {
  render(
    <AppShell
      name="Register console"
      navigationLabel="Main"
      navigation={
        <AppShellNavItem
          href="#routed"
          icon={AppWindow}
          render={({ children, ...props }) => (
            <a data-router="" aria-current="page" {...props}>
              {children}
            </a>
          )}
        >
          Data recipients
        </AppShellNavItem>
      }
    >
      <h1>Data recipients</h1>
    </AppShell>,
  );

  const item = page.getByRole("link", { name: "Data recipients" });
  await expect.element(item).toHaveAttribute("data-router");
  await expect.element(item).toHaveAttribute("aria-current", "page");
  expect(item.element().querySelector("svg")).not.toBeNull();
});
