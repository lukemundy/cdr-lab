import { expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize, expectVisibleFocus } from "../test/checks";
import { render } from "../test/render";
import { Pagination } from "./pagination";

function renderPagination(current: number, pageCount: number) {
  const pages: number[] = [];
  const container = render(
    <Pagination
      page={current}
      pageCount={pageCount}
      rowCount={95}
      onPageChange={(p) => pages.push(p)}
    />,
  );
  return { container, pages };
}

test("the pagination tells the user the current page, the number of pages and the number of records", async () => {
  const { container } = renderPagination(2, 4);

  const navigation = page.getByRole("navigation", { name: "Pagination" });
  await expect.element(navigation).toMatchTextContent("Page 2 of 4");
  await expect.element(navigation).toMatchTextContent("95 records");
  await expectNoAxeViolations(container);
});

test("the page text is a status message, so that a change of page is announced", async () => {
  renderPagination(2, 4);

  await expect
    .element(page.getByRole("status").filter({ hasText: "Page" }))
    .toHaveTextContent("Page 2 of 4");
});

test("the number of records is a status message, so that the result of a new search is announced", async () => {
  renderPagination(2, 4);

  await expect
    .element(page.getByRole("status").filter({ hasText: "records" }))
    .toHaveTextContent("95 records");
});

test("Previous and Next ask for the page before and after", async () => {
  const { pages } = renderPagination(2, 4);

  await page.getByRole("button", { name: "Previous page" }).click();
  await page.getByRole("button", { name: "Next page" }).click();

  expect(pages).toEqual([1, 3]);
});

test("on the first page, Previous is disabled", async () => {
  const { container } = renderPagination(1, 4);

  await expect.element(page.getByRole("button", { name: "Previous page" })).toBeDisabled();
  await expect.element(page.getByRole("button", { name: "Next page" })).toBeEnabled();
  await expectNoAxeViolations(container);
});

test("on the last page, Next is disabled", async () => {
  renderPagination(4, 4);

  await expect.element(page.getByRole("button", { name: "Next page" })).toBeDisabled();
});

test("one record gives the singular", async () => {
  render(<Pagination page={1} pageCount={1} rowCount={1} onPageChange={() => {}} />);

  await expect.element(page.getByText("1 record", { exact: true })).toBeVisible();
});

test("no records gives page 1 of 1", async () => {
  render(<Pagination page={1} pageCount={0} rowCount={0} onPageChange={() => {}} />);

  await expect
    .element(page.getByRole("status").filter({ hasText: "Page" }))
    .toHaveTextContent("Page 1 of 1");
});

test("the keyboard reaches and uses each button with a visible focus", async () => {
  const { pages } = renderPagination(2, 4);

  await userEvent.tab();
  expectVisibleFocus(page.getByRole("button", { name: "Previous page" }).element());
  await userEvent.tab();
  expectVisibleFocus(page.getByRole("button", { name: "Next page" }).element());
  await userEvent.keyboard("{Enter}");

  expect(pages).toEqual([3]);
});

test("each button is 24 by 24 CSS px or more", () => {
  renderPagination(2, 4);

  expectTargetSize(page.getByRole("button", { name: "Previous page" }).element());
  expectTargetSize(page.getByRole("button", { name: "Next page" }).element());
});
