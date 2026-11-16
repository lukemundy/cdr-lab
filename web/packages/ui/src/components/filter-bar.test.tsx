import { useState } from "react";
import { expect, onTestFinished, test, vi } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize, expectVisibleFocus } from "../test/checks";
import { render } from "../test/render";
import { FilterBar, FilterSelect } from "./filter-bar";

const statuses = [
  { value: "", label: "All statuses" },
  { value: "ACTIVE", label: "Active" },
  { value: "SUSPENDED", label: "Suspended" },
];

function renderFilterBar(search = "", status = "") {
  const calls = { searches: [] as string[], statuses: [] as string[] };
  const container = render(
    <FilterBar
      label="Filter the data recipients"
      searchLabel="Name"
      search={search}
      onSearch={(value) => calls.searches.push(value)}
    >
      <FilterSelect
        label="Status"
        value={status}
        options={statuses}
        onChange={(value) => calls.statuses.push(value)}
      />
    </FilterBar>,
  );
  return { container, calls };
}

// Verifies: UI-0269
test("the filter bar is a search landmark, and each field has a visible label", async () => {
  const { container } = renderFilterBar();

  await expect
    .element(page.getByRole("search", { name: "Filter the data recipients" }))
    .toBeVisible();
  await expect.element(page.getByText("Name", { exact: true })).toBeVisible();
  await expect.element(page.getByText("Status", { exact: true })).toBeVisible();
  await expect.element(page.getByRole("searchbox", { name: "Name" })).toBeVisible();
  await expect.element(page.getByRole("combobox", { name: "Status" })).toBeVisible();
  await expectNoAxeViolations(container);
});

// The debounce of the filter bar is 300 ms.
const afterTheDebounce = () => new Promise((resolve) => setTimeout(resolve, 400));

// The parent gives the applied search back to the filter bar after a delay, as the router of a list page does.
function ControlledFilterBar({
  initial = "",
  onSearch,
}: {
  initial?: string;
  onSearch?: (search: string) => void;
}) {
  const [search, setSearch] = useState(initial);
  return (
    <>
      <FilterBar
        label="Filter the data recipients"
        searchLabel="Name"
        search={search}
        onSearch={(value) => {
          onSearch?.(value);
          setTimeout(() => setSearch(value), 50);
        }}
      />
      <output aria-label="Applied search">{search}</output>
      <button type="button" onClick={() => setSearch("Other")}>
        Go back
      </button>
    </>
  );
}

// Fake timers, so that a slow test machine does not make a pause between two key presses.
function fakeTimers() {
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
  onTestFinished(() => {
    vi.useRealTimers();
  });
}

// Verifies: UI-0271
test("the search applies 300 ms after the last key press, one time for each pause, with no submit", async () => {
  fakeTimers();
  const { calls } = renderFilterBar();
  const field = page.getByRole("searchbox", { name: "Name" });

  await userEvent.type(field, "  Mock");
  vi.advanceTimersByTime(299);
  await userEvent.type(field, " Recipient ");
  vi.advanceTimersByTime(299);
  expect(calls.searches).toEqual([]);

  vi.advanceTimersByTime(1);
  expect(calls.searches).toEqual(["Mock Recipient"]);
  vi.advanceTimersByTime(1000);
  expect(calls.searches).toEqual(["Mock Recipient"]);
});

// Verifies: UI-0271
test("the Enter key applies the search at once, and the delay then does not apply it again", async () => {
  fakeTimers();
  const { calls } = renderFilterBar();

  await userEvent.type(page.getByRole("searchbox", { name: "Name" }), "Bank");
  await userEvent.keyboard("{Enter}");
  expect(calls.searches).toEqual(["Bank"]);

  vi.advanceTimersByTime(1000);
  expect(calls.searches).toEqual(["Bank"]);
});

test("the search field shows the current search, and an empty search clears it", async () => {
  const { calls } = renderFilterBar("Bank");
  const field = page.getByRole("searchbox", { name: "Name" });

  await expect.element(field).toHaveValue("Bank");
  await userEvent.clear(field);
  await userEvent.keyboard("{Enter}");

  expect(calls.searches).toEqual([""]);
});

test("there is no Search button", async () => {
  renderFilterBar("Bank");

  await expect.element(page.getByRole("button", { name: "Search" })).not.toBeInTheDocument();
});

// Verifies: UI-0272
test("the clear button shows only when the field has text", async () => {
  renderFilterBar();
  const clear = page.getByRole("button", { name: "Clear search" });

  await expect.element(clear).not.toBeInTheDocument();
  await userEvent.type(page.getByRole("searchbox", { name: "Name" }), "B");
  await expect.element(clear).toBeVisible();
});

// Verifies: UI-0272
test("the clear button empties the field, applies the empty search at once and puts the focus in the field", async () => {
  const searches: string[] = [];
  const container = render(
    <ControlledFilterBar initial="Bank" onSearch={(value) => searches.push(value)} />,
  );
  const field = page.getByRole("searchbox", { name: "Name" });
  await expectNoAxeViolations(container);

  await page.getByRole("button", { name: "Clear search" }).click();

  expect(searches).toEqual([""]);
  await expect.element(field).toHaveValue("");
  expect(document.activeElement).toBe(field.element());
  await expect.element(page.getByRole("button", { name: "Clear search" })).not.toBeInTheDocument();
  await afterTheDebounce();
  expect(searches).toEqual([""]);
});

// Verifies: UI-0273
test("the field keeps the focus and the text that the user types while the search applies", async () => {
  render(<ControlledFilterBar />);
  const field = page.getByRole("searchbox", { name: "Name" });

  await userEvent.type(field, "Bank ");
  await expect
    .element(page.getByRole("status", { name: "Applied search" }))
    .toHaveTextContent("Bank");

  expect(document.activeElement).toBe(field.element());
  await expect.element(field).toHaveValue("Bank ");
  await userEvent.keyboard("A");
  await expect
    .element(page.getByRole("status", { name: "Applied search" }))
    .toHaveTextContent("Bank A");
  await expect.element(field).toHaveValue("Bank A");
});

test("a change of the search outside the field, for example in the URL, updates the field", async () => {
  render(<ControlledFilterBar />);
  const field = page.getByRole("searchbox", { name: "Name" });

  await userEvent.type(field, "Bank");
  await expect
    .element(page.getByRole("status", { name: "Applied search" }))
    .toHaveTextContent("Bank");
  await page.getByRole("button", { name: "Go back" }).click();

  await expect.element(field).toHaveValue("Other");
});

// Verifies: UI-0274
test("a filter applies when the user selects a value with the keyboard", async () => {
  const { calls } = renderFilterBar();

  await userEvent.selectOptions(page.getByRole("combobox", { name: "Status" }), "SUSPENDED");

  expect(calls.statuses).toEqual(["SUSPENDED"]);
});

// Verifies: UI-0201, UI-0202
test("the keyboard reaches each field and the clear button with a visible focus", async () => {
  renderFilterBar("Bank");

  await userEvent.tab();
  expectVisibleFocus(page.getByRole("searchbox", { name: "Name" }).element());
  await userEvent.tab();
  expectVisibleFocus(page.getByRole("button", { name: "Clear search" }).element());
  await userEvent.tab();
  expectVisibleFocus(page.getByRole("combobox", { name: "Status" }).element());
});

// Verifies: UI-0094
test("each control is 24 by 24 CSS px or more", () => {
  renderFilterBar("Bank");

  expectTargetSize(page.getByRole("searchbox", { name: "Name" }).element());
  expectTargetSize(page.getByRole("button", { name: "Clear search" }).element());
  expectTargetSize(page.getByRole("combobox", { name: "Status" }).element());
});

test("a filter bar with no search label has only its filters", async () => {
  const onChange = vi.fn<(value: string) => void>();
  const container = render(
    <FilterBar label="Filter the software products">
      <FilterSelect label="Status" value="" options={statuses} onChange={onChange} />
    </FilterBar>,
  );

  const bar = page.getByRole("search", { name: "Filter the software products" });
  await expect.element(bar).toBeVisible();
  expect(container.querySelector("input")).toBeNull();
  await userEvent.selectOptions(bar.getByRole("combobox", { name: "Status" }), "ACTIVE");
  expect(onChange).toHaveBeenCalledWith("ACTIVE");
  await expectNoAxeViolations(container);
});
