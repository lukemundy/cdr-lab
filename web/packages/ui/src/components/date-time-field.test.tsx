import { expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { expectNoAxeViolations, expectTargetSize, expectVisibleFocus } from "../test/checks";
import { render } from "../test/render";
import { DateTimeField, utcOf } from "./date-time-field";

// The tests run in more than one time zone, so the local parts of the UTC time must be the parts that the user typed.
function localParts(utc: string): number[] {
  const date = new Date(utc);
  return [
    date.getFullYear(),
    date.getMonth() + 1,
    date.getDate(),
    date.getHours(),
    date.getMinutes(),
    date.getSeconds(),
  ];
}

// Verifies: UI-0313
test("a time in the time zone of the browser becomes the same instant in UTC", () => {
  const utc = utcOf("2026-10-26T14:05:30");

  expect(utc).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/);
  expect(localParts(utc)).toEqual([2026, 10, 26, 14, 5, 30]);
  expect(localParts(utcOf("2026-01-15T23:59"))).toEqual([2026, 1, 15, 23, 59, 0]);
});

// Verifies: UI-0313
test("the offset of the time is the offset of the browser at that instant, also with daylight saving", () => {
  for (const local of ["2026-01-15T12:00", "2026-07-15T12:00"]) {
    const offsetMinutes = new Date(local).getTimezoneOffset();
    const utc = new Date(utcOf(local));

    expect(utc.getTime()).toBe(
      Date.UTC(2026, Number(local.slice(5, 7)) - 1, 15, 12) + offsetMinutes * 60_000,
    );
  }
});

test("an empty value stays empty, and a value that is not a time goes to the server as it is", () => {
  expect(utcOf("")).toBe("");
  expect(utcOf("zebracorn")).toBe("zebracorn");
});

// Verifies: UI-0302
test("a date-time field has a visible label, help text and error, and the keyboard reaches it", async () => {
  const container = render(
    <DateTimeField
      label="Posting time"
      description="In the time zone of this browser."
      error="Give a posting time."
      name="postingDateTime"
    />,
  );
  const field = page.getByLabelText("Posting time");

  expect(field.element()).toHaveProperty("type", "datetime-local");
  await expect.element(field).toHaveAttribute("step", "1");
  await expect.element(field).toHaveAttribute("aria-invalid", "true");
  await expect
    .element(field)
    .toHaveAccessibleDescription("In the time zone of this browser. Give a posting time.");
  await userEvent.tab();
  expectVisibleFocus(field.element());
  expectTargetSize(field.element());
  await expectNoAxeViolations(container);
});

// Verifies: UI-0311
test("the value of a date-time field has seconds and no offset", async () => {
  render(<DateTimeField label="Posting time" name="postingDateTime" />);
  const field = page.getByLabelText("Posting time");

  await userEvent.fill(field, "2026-10-26T14:05:30");

  await expect.element(field).toHaveValue("2026-10-26T14:05:30");
});
