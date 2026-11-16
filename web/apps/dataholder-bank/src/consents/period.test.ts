import { expect, test } from "vitest";

import { formatPeriod } from "./period";

const hour = 3600;
const day = 24 * hour;

// Verifies: DH-1915
test("a sharing duration of 24 hours or less is one occasion, as on the consent page", () => {
  expect(formatPeriod(0)).toBe("One occasion");
  expect(formatPeriod(day)).toBe("One occasion");
});

test("a longer sharing duration shows in days and hours, and a part of an hour counts as one hour", () => {
  expect(formatPeriod(day + 1)).toBe("1 day and 1 hour");
  expect(formatPeriod(2 * day)).toBe("2 days");
  expect(formatPeriod(90 * day + 5 * hour)).toBe("90 days and 5 hours");
});

test("the longest sharing duration of the CDS is one year", () => {
  expect(formatPeriod(365 * day)).toBe("1 year");
});
