import { expect, test } from "vitest";
import { page } from "vitest/browser";

import { expectNoAxeViolations } from "../test/checks";
import { render } from "../test/render";
import { AppShell } from "./app-shell";
import { DescriptionItem, DescriptionList } from "./description-list";
import { Identifier } from "./identifier";

function definitionOf(container: HTMLElement, term: string): HTMLElement {
  const dt = [...container.querySelectorAll("dt")].find((element) => element.textContent === term);
  const dd = dt?.nextElementSibling;
  if (!(dd instanceof HTMLElement) || dd.tagName !== "DD") throw new Error(`no value for ${term}`);
  return dd;
}

// Verifies: UI-0247
test("each value is in the description list with its term", async () => {
  const container = render(
    <DescriptionList>
      <DescriptionItem term="Name">Mock Recipient One</DescriptionItem>
      <DescriptionItem term="Accreditation number">ADR-000001</DescriptionItem>
    </DescriptionList>,
  );

  expect(container.querySelector("dl")).not.toBeNull();
  expect(definitionOf(container, "Name").textContent).toBe("Mock Recipient One");
  expect(definitionOf(container, "Accreditation number").textContent).toBe("ADR-000001");
  await expect.element(page.getByRole("term").first()).toHaveTextContent("Name");
  await expect
    .element(page.getByRole("definition").first())
    .toHaveTextContent("Mock Recipient One");
  await expectNoAxeViolations(container);
});

// Verifies: UI-0249
test("an empty value shows the text Not set", async () => {
  const container = render(
    <DescriptionList>
      <DescriptionItem term="Logo URI" />
      <DescriptionItem term="Brand group">{null}</DescriptionItem>
      <DescriptionItem term="Interim ID">{""}</DescriptionItem>
      <DescriptionItem term="Industries" values={[]} />
    </DescriptionList>,
  );

  for (const term of ["Logo URI", "Brand group", "Interim ID", "Industries"]) {
    expect(definitionOf(container, term).textContent).toBe("Not set");
  }
  await expectNoAxeViolations(container);
});

// Verifies: UI-0250
test("an empty list of values can give its own text", async () => {
  const container = render(
    <DescriptionList>
      <DescriptionItem term="Accounts" values={[]} empty="None" />
      <DescriptionItem term="Scopes" values={["openid"]} empty="None" />
    </DescriptionList>,
  );

  expect(definitionOf(container, "Accounts").textContent).toBe("None");
  expect(definitionOf(container, "Scopes").textContent).toBe("openid");
  await expectNoAxeViolations(container);
});

// Verifies: UI-0248
test("a value can be a list of values", async () => {
  const container = render(
    <DescriptionList>
      <DescriptionItem term="Industries" values={["banking", "energy"]} />
    </DescriptionList>,
  );

  const items = [...definitionOf(container, "Industries").querySelectorAll("ul > li")];
  expect(items.map((item) => item.textContent)).toEqual(["banking", "energy"]);
  await expectNoAxeViolations(container);
});

// Verifies: UI-0251
test("a long thumbprint and a long URI wrap, and the page has no horizontal scroll at 1280 CSS px", async () => {
  await page.viewport(1280, 800);
  const thumbprint = "kZ2lQnTXy1vbQ8mcJ0aA7oPqRsTuVwXyZ01234567-_";
  const uri = "https://mock-recipient.example/".padEnd(200, "callback");
  expect(thumbprint).toHaveLength(43);
  expect(uri).toHaveLength(200);

  const container = render(
    <AppShell name="Console" navigationLabel="Main" navigation={null}>
      <DescriptionList>
        <DescriptionItem term="Thumbprint">
          <Identifier value={thumbprint} label="thumbprint" />
        </DescriptionItem>
        <DescriptionItem term="Redirect URIs" values={[uri, `${uri}/2`]} />
        <DescriptionItem term="Plain text">{uri}</DescriptionItem>
      </DescriptionList>
    </AppShell>,
  );

  const root = document.documentElement;
  expect(root.scrollWidth).toBeLessThanOrEqual(root.clientWidth);
  const value = definitionOf(container, "Plain text");
  expect(value.getBoundingClientRect().height).toBeGreaterThan(
    Number.parseFloat(getComputedStyle(value).lineHeight) * 1.5,
  );
  await expectNoAxeViolations(container);
});

// Verifies: UI-0086
test("in a narrow container, each value is below its term, and the list has no horizontal scroll", () => {
  const container = render(
    <div style={{ width: "280px" }}>
      <DescriptionList>
        <DescriptionItem term="Account number">xxx-xxx xxxx5678</DescriptionItem>
      </DescriptionList>
    </div>,
  );

  const term = container.querySelector("dt")!.getBoundingClientRect();
  const value = container.querySelector("dd")!.getBoundingClientRect();
  expect(value.top).toBeGreaterThanOrEqual(term.bottom);
  expect(value.left).toBe(term.left);
  const list = container.querySelector("dl")!;
  expect(list.scrollWidth).toBeLessThanOrEqual(280);
});

test("in a wide container, each value is next to its term", () => {
  const container = render(
    <div style={{ width: "800px" }}>
      <DescriptionList>
        <DescriptionItem term="Account number">xxx-xxx xxxx5678</DescriptionItem>
      </DescriptionList>
    </div>,
  );

  const term = container.querySelector("dt")!.getBoundingClientRect();
  const value = container.querySelector("dd")!.getBoundingClientRect();
  expect(value.top).toBe(term.top);
  expect(value.left).toBeGreaterThan(term.right);
});
