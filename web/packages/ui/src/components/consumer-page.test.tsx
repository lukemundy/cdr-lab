import "../consumer.css";

import { expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations } from "../test/checks";
import { render } from "../test/render";
import {
  ConsumerButton,
  ConsumerCard,
  ConsumerFooter,
  ConsumerNavItem,
  ConsumerShell,
  PasswordStatement,
} from "./consumer-page";

const logo = `data:image/svg+xml,${encodeURIComponent(
  '<svg xmlns="http://www.w3.org/2000/svg" width="96" height="32"><rect width="96" height="32" fill="white"/></svg>',
)}`;

function shell() {
  return render(
    <ConsumerShell
      brandName="Wattle Bank"
      logo={logo}
      navigationLabel="Main"
      navigation={<ConsumerNavItem href="/products">Products</ConsumerNavItem>}
      footer={
        <ConsumerFooter>
          <PasswordStatement>Third parties never need your password.</PasswordStatement>
          <p>Wattle Bank is not a real bank.</p>
        </ConsumerFooter>
      }
    >
      <ConsumerCard>
        <h1>Banking with Wattle Bank</h1>
        <ConsumerButton href="/products">See all products</ConsumerButton>
        <ConsumerButton variant="secondary" href="/">
          Home
        </ConsumerButton>
      </ConsumerCard>
    </ConsumerShell>,
  );
}

test("the shell gives the logo with the brand name as its text alternative, the navigation, the main content and the footer", async () => {
  const container = shell();

  await expect
    .element(page.getByRole("banner").getByRole("img", { name: "Wattle Bank" }))
    .toBeVisible();
  await expect
    .element(page.getByRole("navigation", { name: "Main" }).getByRole("link", { name: "Products" }))
    .toBeVisible();
  await expect
    .element(
      page.getByRole("main").getByRole("heading", { level: 1, name: "Banking with Wattle Bank" }),
    )
    .toBeVisible();
  await expect
    .element(page.getByRole("contentinfo"))
    .toHaveTextContent("Third parties never need your password.Wattle Bank is not a real bank.");
  await expectNoAxeViolations(container);
});

test("the logo links to the home page", async () => {
  shell();

  await expect
    .element(page.getByRole("link", { name: "Wattle Bank" }))
    .toHaveAttribute("href", "/");
});

test("the shell has no Secure label, because the site must not look like the CDR authentication", async () => {
  shell();

  await expect.element(page.getByRole("banner")).not.toHaveTextContent("Secure");
});

test("the skip link moves the focus to the main content", async () => {
  shell();

  await userEvent.keyboard("{Tab}");
  const skip = page.getByRole("link", { name: "Skip to content" });
  await expect.element(skip).toHaveFocus();

  await userEvent.keyboard("{Enter}");
  await expect.element(page.getByRole("main")).toHaveFocus();
});

test("the footer has no link", () => {
  shell();

  expect(document.querySelectorAll("footer a")).toHaveLength(0);
});

// DESIGN.md (Layout): in a consumer UI, each control is 44 by 44 CSS px or more.
test("each link of the shell and each button link is 44 CSS px high or more", () => {
  shell();

  const lowLinks = ["See all products", "Home", "Products", "Wattle Bank"].filter(
    (name) =>
      page.getByRole("link", { name, exact: true }).element().getBoundingClientRect().height < 44,
  );
  expect(lowLinks).toEqual([]);
});
