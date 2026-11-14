import { expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations } from "../test/checks";
import { render } from "../test/render";
import { Link } from "./link";
import { PageHeader } from "./page-header";
import { Section } from "./section";

test("a section is a region that its heading names, with its description and links", async () => {
  const container = render(
    <Section
      title="Legal entity"
      description="The accredited person."
      links={<Link href="#change">Change details</Link>}
    >
      <p>Content</p>
    </Section>,
  );

  const region = page.getByRole("region", { name: "Legal entity" });
  await expect.element(region).toBeVisible();
  await expect
    .element(region.getByRole("heading", { level: 2, name: "Legal entity" }))
    .toBeVisible();
  await expect.element(region.getByText("The accredited person.")).toBeVisible();
  await expect.element(region.getByRole("link", { name: "Change details" })).toBeVisible();
  await expect.element(region.getByText("Content")).toBeVisible();
  await expectNoAxeViolations(container);
});

test("a section under the page header has a level 2 heading, and each nested section is one level below its parent", async () => {
  const container = render(
    <>
      <PageHeader title="Mock Recipient One" />
      <Section title="Brands">
        <Section title="Mock Brand">
          <Section title="Software products">
            <Section title="Mock Product">
              <Section title="Metadata">
                <Section title="Certificates">
                  <p>Deepest</p>
                </Section>
              </Section>
            </Section>
          </Section>
        </Section>
      </Section>
      <Section title="Status">
        <p>Second top-level section</p>
      </Section>
    </>,
  );

  const levels = [...container.querySelectorAll("h1, h2, h3, h4, h5, h6")].map(
    (heading) => `${heading.tagName} ${heading.textContent}`,
  );
  // HTML has no heading below h6, so a section below that level keeps h6.
  expect(levels).toEqual([
    "H1 Mock Recipient One",
    "H2 Brands",
    "H3 Mock Brand",
    "H4 Software products",
    "H5 Mock Product",
    "H6 Metadata",
    "H6 Certificates",
    "H2 Status",
  ]);
  await expectNoAxeViolations(container);
});

test("a section needs only a title", async () => {
  const container = render(<Section title="Keys">Content</Section>);

  await expect.element(page.getByRole("region", { name: "Keys" })).toBeVisible();
  await expectNoAxeViolations(container);
});

test("the links of a section are directly after its heading, on the same line", () => {
  render(
    <Section title="Legal entity" links={<Link href="#change">Change details</Link>}>
      <p>Content</p>
    </Section>,
  );

  const heading = page
    .getByRole("heading", { name: "Legal entity" })
    .element()
    .getBoundingClientRect();
  const link = page.getByRole("link", { name: "Change details" }).element().getBoundingClientRect();
  expect(link.left).toBeGreaterThan(heading.right);
  // A gap separates the link from the heading, but the link is not at the far side of the section.
  expect(link.left - heading.right).toBeLessThan(48);
  expect(link.top).toBeLessThan(heading.bottom);
  expect(link.bottom).toBeGreaterThan(heading.top);
});

test("a card section has a border, a title bar with its heading and links, and a nested card has a different background", async () => {
  const container = render(
    <Section card title="Mock Brand" links={<Link href="#change">Change brand</Link>}>
      <p>Brand details</p>
      <Section card title="Mock Product">
        <p>Product details</p>
      </Section>
    </Section>,
  );

  const brand = page.getByRole("region", { name: "Mock Brand" }).element();
  const product = page.getByRole("region", { name: "Mock Product" }).element();
  expect(getComputedStyle(brand).borderTopWidth).not.toBe("0px");
  expect(getComputedStyle(product).backgroundColor).not.toBe(
    getComputedStyle(brand).backgroundColor,
  );

  const bar = brand.firstElementChild!;
  expect(bar).toContainElement(page.getByRole("heading", { name: "Mock Brand" }).element());
  expect(bar).toContainElement(page.getByRole("link", { name: "Change brand" }).element());
  expect(getComputedStyle(bar).borderBottomWidth).not.toBe("0px");
  await expectNoAxeViolations(container);
});

const fontSizeOf = (name: string) =>
  Number.parseFloat(getComputedStyle(page.getByRole("heading", { name }).element()).fontSize);

test("each heading level from 2 to 4 has a smaller font than the level above it", () => {
  render(
    <Section title="Brands">
      <Section title="Mock Brand">
        <Section title="Software products">
          <p>Content</p>
        </Section>
      </Section>
    </Section>,
  );

  expect(fontSizeOf("Brands")).toBeGreaterThan(fontSizeOf("Mock Brand"));
  expect(fontSizeOf("Mock Brand")).toBeGreaterThan(fontSizeOf("Software products"));
});

test("a collapsible card starts collapsed, with its heading, status and links visible, and its button opens it", async () => {
  const container = render(
    <Section
      card
      collapsible
      title="Mock Brand"
      status={<span>REMOVED</span>}
      links={<Link href="#change">Change brand</Link>}
    >
      <p>Brand details</p>
    </Section>,
  );

  const toggle = page.getByRole("button", { name: "Mock Brand" });
  await expect.element(toggle).toHaveAttribute("aria-expanded", "false");
  await expect
    .element(page.getByRole("heading", { name: "Mock Brand" }))
    .toContainElement(toggle.element());
  await expect.element(page.getByText("REMOVED")).toBeVisible();
  await expect.element(page.getByRole("link", { name: "Change brand" })).toBeVisible();
  await expect.element(page.getByText("Brand details")).not.toBeVisible();
  expect(page.getByText("Brand details").element().closest("[hidden]")).toHaveAttribute(
    "hidden",
    "until-found",
  );
  await expectNoAxeViolations(container);

  await toggle.click();

  await expect.element(toggle).toHaveAttribute("aria-expanded", "true");
  await expect.element(page.getByText("Brand details")).toBeVisible();
  await expect.element(toggle).toHaveFocus();
  await userEvent.keyboard("{Enter}");
  await expect.element(page.getByText("Brand details")).not.toBeVisible();
});

test("a collapsible card can start open", async () => {
  render(
    <Section card collapsible defaultOpen title="Mock Brand">
      <p>Brand details</p>
    </Section>,
  );

  await expect
    .element(page.getByRole("button", { name: "Mock Brand" }))
    .toHaveAttribute("aria-expanded", "true");
  await expect.element(page.getByText("Brand details")).toBeVisible();
});

test("a find in the page opens the collapsible card of the text that it finds", async () => {
  render(
    <Section card collapsible title="Mock Brand">
      <p>Brand details</p>
    </Section>,
  );

  page
    .getByText("Brand details")
    .element()
    .closest("[hidden]")!
    .dispatchEvent(new Event("beforematch"));

  await expect
    .element(page.getByRole("button", { name: "Mock Brand" }))
    .toHaveAttribute("aria-expanded", "true");
});

test("Show all and Hide all open and close each collapsible card in the section, also nested cards", async () => {
  const container = render(
    <Section title="Brands" showAll="brands">
      <Section card collapsible title="Mock Brand">
        <p>Brand details</p>
        <Section card collapsible title="Mock Product">
          <p>Product details</p>
        </Section>
      </Section>
      <Section card collapsible title="Mock Saver">
        <p>Saver details</p>
      </Section>
    </Section>,
  );

  await page.getByRole("button", { name: "Show all brands" }).click();

  await expect
    .element(page.getByRole("button", { name: "Mock Brand" }))
    .toHaveAttribute("aria-expanded", "true");
  await expect
    .element(page.getByRole("button", { name: "Mock Product" }))
    .toHaveAttribute("aria-expanded", "true");
  await expect
    .element(page.getByRole("button", { name: "Mock Saver" }))
    .toHaveAttribute("aria-expanded", "true");
  await expectNoAxeViolations(container);

  await page.getByRole("button", { name: "Hide all brands" }).click();

  await expect
    .element(page.getByRole("button", { name: "Mock Brand" }))
    .toHaveAttribute("aria-expanded", "false");
  await expect
    .element(page.getByRole("button", { name: "Mock Saver" }))
    .toHaveAttribute("aria-expanded", "false");
  await page.getByRole("button", { name: "Mock Brand" }).click();
  await expect
    .element(page.getByRole("button", { name: "Mock Product" }))
    .toHaveAttribute("aria-expanded", "false");
});

test("a collapsed card is only as tall as its title bar", () => {
  render(
    <Section card collapsible title="Mock Brand">
      <p>Brand details</p>
    </Section>,
  );

  const card = page.getByRole("region", { name: "Mock Brand" }).element();
  const bar = card.firstElementChild!;
  // The card has a border of 1 CSS px at the top and at the bottom.
  expect(card.getBoundingClientRect().height).toBeLessThanOrEqual(
    bar.getBoundingClientRect().height + 2,
  );
});

test("the button of a collapsible card shows the pointer cursor", () => {
  render(
    <Section card collapsible title="Mock Brand">
      <p>Brand details</p>
    </Section>,
  );

  expect(getComputedStyle(page.getByRole("button", { name: "Mock Brand" }).element()).cursor).toBe(
    "pointer",
  );
});

test("Show all of a nested section opens only its cards, and Show all of the outer section also opens them", async () => {
  render(
    <Section title="Brands" showAll="brands">
      {["Mock Brand", "Mock Saver"].map((brand) => (
        <Section key={brand} card collapsible defaultOpen title={brand}>
          <Section title="Software products" showAll={`software products of ${brand}`}>
            <Section card collapsible title={`${brand} App`}>
              <p>Product details</p>
            </Section>
          </Section>
        </Section>
      ))}
    </Section>,
  );

  await page.getByRole("button", { name: "Show all software products of Mock Brand" }).click();

  await expect
    .element(page.getByRole("button", { name: "Mock Brand App" }))
    .toHaveAttribute("aria-expanded", "true");
  await expect
    .element(page.getByRole("button", { name: "Mock Saver App" }))
    .toHaveAttribute("aria-expanded", "false");

  await page.getByRole("button", { name: "Show all brands" }).click();

  await expect
    .element(page.getByRole("button", { name: "Mock Saver App" }))
    .toHaveAttribute("aria-expanded", "true");
});
