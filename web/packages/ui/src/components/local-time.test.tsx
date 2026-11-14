import { describe, expect, test } from "vitest";
import { page } from "vitest/browser";

import { expectNoAxeViolations } from "../test/checks";
import { render } from "../test/render";
import { LocalTime } from "./local-time";

// The workspace Vitest config runs this file in each of these browser time zones.
const zone = Intl.DateTimeFormat().resolvedOptions().timeZone;
const zones = ["Australia/Perth", "Australia/Sydney", "UTC", "America/St_Johns"];

function timeOf(container: HTMLElement): HTMLTimeElement {
  const time = container.querySelector("time");
  if (!time) throw new Error("no time element");
  return time;
}

function textWithout(element: Element, selector: string): string {
  return [...element.childNodes]
    .filter((node) => !(node instanceof Element && node.matches(selector)))
    .map((node) => node.textContent)
    .join("");
}

const visibleText = (element: Element) => textWithout(element, ".sr-only");
const spokenText = (element: Element) => textWithout(element, "[aria-hidden='true']");

function renderTime(value: string, seconds = true) {
  return timeOf(render(<LocalTime value={value} seconds={seconds} />));
}

test("the browser has one of the time zones of the tests", () => {
  expect(zones).toContain(zone);
});

test("the datetime value is the exact instant in UTC", () => {
  expect(renderTime("2026-10-24T08:05:07+08:00").dateTime).toBe("2026-10-24T00:05:07.000Z");
});

// HTML permits 3 fractional digits or fewer in datetime, and Go gives up to 9.
test("a time with nanoseconds gives a valid datetime", () => {
  expect(renderTime("2026-10-23T14:05:07.123456789Z").dateTime).toBe("2026-10-23T14:05:07.123Z");
});

test("a value that is not a time shows as it is, with no time element", async () => {
  const container = render(<LocalTime value="not a time" />);

  await expect.element(page.getByText("not a time")).toBeVisible();
  expect(container.querySelector("time")).toBeNull();
});

test("the time has no accessibility violations", async () => {
  const time = renderTime("2026-10-23T06:05:07Z");

  await expectNoAxeViolations(time.parentElement ?? time);
});

describe.runIf(zone === "Australia/Perth")("in Australia/Perth", () => {
  test("the time shows in local time with its offset from UTC", () => {
    const time = renderTime("2026-10-23T06:05:07Z");

    expect(visibleText(time)).toBe("23 October 2026, 2:05:07 pm UTC+8");
    expect(spokenText(time)).toBe("23 October 2026, 2:05:07 pm UTC plus 8 hours");
  });

  test("a selection of the time has the visible text only", () => {
    const time = renderTime("2026-10-23T06:05:07Z");

    const selection = window.getSelection();
    selection?.selectAllChildren(time);
    expect(selection?.toString()).toBe("23 October 2026, 2:05:07 pm UTC+8");
  });

  test("without seconds, the time has no seconds", () => {
    expect(visibleText(renderTime("2026-10-23T06:05:07Z", false))).toBe(
      "23 October 2026, 2:05 pm UTC+8",
    );
  });

  test("a time near midnight UTC shows the local date", () => {
    expect(visibleText(renderTime("2026-10-23T23:30:00Z"))).toBe(
      "24 October 2026, 7:30:00 am UTC+8",
    );
  });
});

describe.runIf(zone === "Australia/Sydney")("in Australia/Sydney", () => {
  // Daylight saving starts on 4 October 2026 at 2:00 am, when the clock goes to 3:00 am.
  test("a time before and a time after the start of daylight saving have different offsets", () => {
    expect(visibleText(renderTime("2026-10-03T15:59:59Z"))).toBe(
      "4 October 2026, 1:59:59 am UTC+10",
    );
    expect(visibleText(renderTime("2026-10-03T16:00:00Z"))).toBe(
      "4 October 2026, 3:00:00 am UTC+11",
    );
  });

  // Daylight saving ends on 5 April 2026 at 3:00 am, when the clock goes back to 2:00 am, so 2:30 am occurs two times.
  test("the two times of a repeated hour at the end of daylight saving have different offsets", () => {
    const first = renderTime("2026-04-04T15:30:00Z");
    const second = renderTime("2026-04-04T16:30:00Z");

    expect(visibleText(first)).toBe("5 April 2026, 2:30:00 am UTC+11");
    expect(visibleText(second)).toBe("5 April 2026, 2:30:00 am UTC+10");
    expect(spokenText(second)).toBe("5 April 2026, 2:30:00 am UTC plus 10 hours");
  });
});

describe.runIf(zone === "UTC")("in UTC", () => {
  test("the time shows in UTC with the label UTC", () => {
    const time = renderTime("2026-10-23T06:05:07Z");

    expect(visibleText(time)).toBe("23 October 2026, 6:05:07 am UTC");
    expect(spokenText(time)).toBe("23 October 2026, 6:05:07 am UTC");
  });

  test("a time with an offset shows in UTC", () => {
    expect(visibleText(renderTime("2026-10-24T08:05:07+08:00"))).toBe(
      "24 October 2026, 12:05:07 am UTC",
    );
  });
});

describe.runIf(zone === "America/St_Johns")("in America/St_Johns", () => {
  test("a negative offset shows its hours and minutes, and a screen reader says minus", () => {
    const time = renderTime("2026-01-15T12:00:00Z");

    expect(visibleText(time)).toBe("15 January 2026, 8:30:00 am UTC-3:30");
    expect(spokenText(time)).toBe("15 January 2026, 8:30:00 am UTC minus 3 hours 30 minutes");
  });
});
