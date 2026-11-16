import { expect, test } from "vitest";
import { page } from "vitest/browser";

import { expectNoAxeViolations } from "../test/checks";
import { render } from "../test/render";
import { EmptyState, ErrorState, LoadingState } from "./states";

// Verifies: UI-0275
test("the empty state gives its title and description", async () => {
  const container = render(
    <EmptyState title="No data recipients" description="Change the filters to see more." />,
  );

  await expect.element(page.getByText("No data recipients")).toBeVisible();
  await expect.element(page.getByText("Change the filters to see more.")).toBeVisible();
  await expectNoAxeViolations(container);
});

// Verifies: UI-0275
test("the loading state is a status message", async () => {
  const container = render(<LoadingState />);

  await expect.element(page.getByRole("status")).toHaveTextContent("Loading");
  await expectNoAxeViolations(container);
});

// Verifies: UI-0276
test("the error state is an alert with the title and the detail of the problem", async () => {
  const container = render(
    <ErrorState
      problem={{ title: "Service Unavailable", detail: "The Register cannot read its database." }}
    />,
  );

  const alert = page.getByRole("alert");
  await expect.element(alert).toMatchTextContent("Service Unavailable");
  await expect.element(alert).toMatchTextContent("The Register cannot read its database.");
  await expectNoAxeViolations(container);
});

test("the error state shows a problem with no detail", async () => {
  render(<ErrorState problem={{ title: "Not Found" }} />);

  await expect.element(page.getByRole("alert")).toMatchTextContent("Not Found");
});
