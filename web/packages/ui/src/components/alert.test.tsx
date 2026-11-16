import { useState } from "react";
import { expect, test } from "vitest";
import { page } from "vitest/browser";

import { expectNoAxeViolations } from "../test/checks";
import { render } from "../test/render";
import { Alert } from "./alert";
import { Button } from "./button";

// Verifies: UI-0333
test("an alert gives its title and its text, and passes the accessibility checks", async () => {
  const container = render(
    <Alert title="Conflict">The data recipient changed. See the current status.</Alert>,
  );

  const alert = page.getByRole("alert");
  await expect.element(alert).toMatchTextContent("Conflict");
  await expect.element(alert).toMatchTextContent("The data recipient changed.");
  await expectNoAxeViolations(container);
});

test("an alert can have only a title", async () => {
  render(<Alert title="Not Found" />);

  await expect.element(page.getByRole("alert")).toHaveTextContent("Not Found");
});

function FailingForm() {
  const [failed, setFailed] = useState(false);
  return (
    <>
      <Button onClick={() => setFailed(true)}>Send</Button>
      {failed && <Alert title="Bad Request">The request is not valid.</Alert>}
    </>
  );
}

// A screen reader announces an element with the alert role when it is added to the page.
// Verifies: UI-0333
test("an alert that appears is an alert when it is added", async () => {
  render(<FailingForm />);
  expect(page.getByRole("alert").query()).toBeNull();

  await page.getByRole("button", { name: "Send" }).click();

  await expect.element(page.getByRole("alert")).toMatchTextContent("Bad Request");
});

// Verifies: UI-0334
test("an alert has no close button, so it stays until its cause changes", () => {
  render(<Alert title="Conflict">Detail.</Alert>);

  expect(page.getByRole("button").query()).toBeNull();
});
