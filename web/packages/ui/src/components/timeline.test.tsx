import { expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize, expectVisibleFocus } from "../test/checks";
import { render } from "../test/render";
import { DescriptionItem, DescriptionList } from "./description-list";
import { Timeline, type TimelineEvent, type TimelineProps } from "./timeline";

const events: TimelineEvent[] = [
  {
    id: "3",
    time: "2026-10-23T14:05:07.123456Z",
    title: "Software product deactivated",
    details: (
      <DescriptionList>
        <DescriptionItem term="Software product">Mock Product</DescriptionItem>
      </DescriptionList>
    ),
  },
  { id: "2", time: "2026-10-23T14:05:07.123456Z", title: "Accreditation suspended" },
  { id: "1", time: "2026-10-22T01:00:00Z", title: "Loaded into the Register" },
];

function renderTimeline(props: Partial<TimelineProps> = {}) {
  const pages: number[] = [];
  const container = render(
    <Timeline
      label="History"
      events={events}
      empty={{ title: "No events" }}
      page={1}
      pageCount={2}
      rowCount={30}
      onPageChange={(next) => pages.push(next)}
      {...props}
    />,
  );
  return { container, pages };
}

test("the timeline is an ordered list of the events, in the order that it gets", async () => {
  const { container } = renderTimeline();

  const list = page.getByRole("list", { name: "History" });
  await expect.element(list).toBeVisible();
  expect(list.element().tagName).toBe("OL");
  const titles = [...list.element().children].map((item) => item.querySelector("p")?.textContent);
  expect(titles).toEqual([
    "Software product deactivated",
    "Accreditation suspended",
    "Loaded into the Register",
  ]);
  await expectNoAxeViolations(container);
});

test("each event has its time in a time element with the ISO 8601 UTC value, and the time has seconds and the offset from UTC", () => {
  const { container } = renderTimeline();

  const times = [...container.querySelectorAll("ol > li time")];
  expect(times.map((time) => time.getAttribute("datetime"))).toEqual([
    "2026-10-23T14:05:07.123Z",
    "2026-10-23T14:05:07.123Z",
    "2026-10-22T01:00:00.000Z",
  ]);
  // The tests run in Australia/Perth (vitest.config.ts).
  expect(times.map((time) => time.textContent)).toEqual([
    expect.stringContaining("23 October 2026, 10:05:07 pm UTC+8"),
    expect.stringContaining("23 October 2026, 10:05:07 pm UTC+8"),
    expect.stringContaining("22 October 2026, 9:00:00 am UTC+8"),
  ]);
});

test("an event shows its details", async () => {
  renderTimeline();

  await expect.element(page.getByText("Mock Product")).toBeVisible();
});

test("the pagination is below the list, and works with the keyboard", async () => {
  const { container, pages } = renderTimeline();

  const nav = page.getByRole("navigation", { name: "Pagination" });
  const list = page.getByRole("list", { name: "History" });
  expect(
    list.element().compareDocumentPosition(nav.element()) & Node.DOCUMENT_POSITION_FOLLOWING,
  ).toBeTruthy();
  await expect
    .element(page.getByRole("status").filter({ hasText: "Page" }))
    .toHaveTextContent("Page 1 of 2");

  const next = page.getByRole("button", { name: "Next page" });
  next.element().focus();
  await userEvent.keyboard("{Enter}");
  expectVisibleFocus(next.element());
  expectTargetSize(next.element());
  expect(pages).toEqual([2]);
  await expectNoAxeViolations(container);
});

test("the empty state shows when there are no events", async () => {
  const { container } = renderTimeline({
    events: [],
    rowCount: 0,
    pageCount: 0,
    empty: { title: "No events", description: "The Register has no events for this record." },
  });

  await expect.element(page.getByText("No events")).toBeVisible();
  await expect.element(page.getByText("The Register has no events for this record.")).toBeVisible();
  expect(container.querySelector("ol")).toBeNull();
  await expectNoAxeViolations(container);
});

test("the loading state is a status message", async () => {
  const { container } = renderTimeline({ events: [], status: { state: "loading" } });

  await expect.element(page.getByRole("status").first()).toHaveTextContent("Loading");
  expect(container.querySelector("ol")).toBeNull();
  await expectNoAxeViolations(container);
});

test("the error state is an alert with the problem", async () => {
  const { container } = renderTimeline({
    events: [],
    status: { state: "error", problem: { title: "Not Found", detail: "No such participant." } },
  });

  const alert = page.getByRole("alert");
  await expect.element(alert).toMatchTextContent("Not Found");
  await expect.element(alert).toMatchTextContent("No such participant.");
  expect(container.querySelector("ol")).toBeNull();
  await expectNoAxeViolations(container);
});
