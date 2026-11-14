import { expectNoAxeViolations } from "@cdr-lab/ui/test/checks";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import type { SoftwareProductStatusList } from "../api/client";
import { definitionOf, http, renderApp, worker } from "../test/app";

// Each request of the page, as the search of its URL.
function serve(body: SoftwareProductStatusList) {
  const requests: string[] = [];
  worker.use(
    http.get("/api/software-product-statuses", ({ request, response }) => {
      requests.push(new URL(request.url).search);
      return response(200).json(body);
    }),
  );
  return requests;
}

const table = () => page.getByRole("table", { name: "Register statuses" });
const rowOf = (id: string) => table().getByRole("row").filter({ hasText: id });
const registrationFilter = () => page.getByRole("combobox", { name: "Registration" });

describe("the Register statuses page", () => {
  test("shows each software product with its name, its status, and the time of the last response of the Register", async () => {
    serve({
      data: [
        {
          softwareProductId: "software-1",
          status: "ACTIVE",
          clientId: "client-1",
          name: "Recipient Web",
        },
        { softwareProductId: "software-2", status: "INACTIVE" },
        { softwareProductId: "software-3", status: "REMOVED" },
      ],
      receivedAt: "2026-10-23T02:05:07.123Z",
    });

    renderApp("/register-statuses");

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Register statuses" }))
      .toBeVisible();
    await expect.element(rowOf("software-1").getByText("ACTIVE", { exact: true })).toBeVisible();
    await expect
      .element(rowOf("software-1").getByRole("link", { name: "Recipient Web" }))
      .toHaveAttribute("href", "/registrations/client-1");
    await expect.element(rowOf("software-2").getByText("Not registered")).toBeVisible();
    await expect.element(rowOf("software-2").getByRole("link")).not.toBeInTheDocument();
    await expect.element(rowOf("software-2").getByText("INACTIVE", { exact: true })).toBeVisible();
    await expect.element(rowOf("software-3").getByText("REMOVED", { exact: true })).toBeVisible();
    await expect
      .element(page.getByRole("button", { name: "Copy software product ID of software-2" }))
      .toBeVisible();
    // The tests run in Australia/Perth (vitest.config.ts).
    const time = definitionOf(document.body, "Last got from the Register").querySelector("time");
    expect(time?.textContent).toContain("23 October 2026, 10:05:07 am UTC+8");
    expect(time?.dateTime).toBe("2026-10-23T02:05:07.123Z");
    await expect.element(page.getByText("3 records")).toBeVisible();
    await expectNoAxeViolations(document.body);
  });

  test("a URL gives the API the same filter and sort", async () => {
    const requests = serve({ data: [], receivedAt: "2026-10-23T02:05:07.123Z" });

    renderApp("/register-statuses?registration=notRegistered&sort=status&order=desc");

    await expect.element(registrationFilter()).toHaveValue("notRegistered");
    await expect
      .element(page.getByRole("columnheader", { name: "Status" }))
      .toHaveAttribute("aria-sort", "descending");
    await expect.element(table().getByText("No software products match")).toBeVisible();
    await expect.element(table().getByText("Change the registration filter.")).toBeVisible();
    expect(requests).toEqual(["?registration=notRegistered&sort=status&order=desc"]);
  });

  test("a value in the URL that is not valid is not sent", async () => {
    const requests = serve({ data: [] });

    renderApp("/register-statuses?registration=maybe&sort=softwareProductId&order=up&q=x&page=2");

    await expect.element(table().getByText("No statuses")).toBeVisible();
    await expect
      .element(page.getByRole("columnheader", { name: "Name" }))
      .toHaveAttribute("aria-sort", "ascending");
    expect(requests).toEqual([""]);
  });

  test("a filter and a sort change the URL and add a history entry", async () => {
    const requests = serve({ data: [{ softwareProductId: "software-1", status: "ACTIVE" }] });
    const router = renderApp("/register-statuses");
    await expect.element(rowOf("software-1")).toBeVisible();
    const historyLength = router.history.length;

    await userEvent.selectOptions(registrationFilter().element(), "Registered");
    await expect.poll(() => router.state.location.searchStr).toBe("?registration=registered");

    await userEvent.click(page.getByRole("button", { name: "Status" }));
    await expect
      .poll(() => router.state.location.searchStr)
      .toBe("?registration=registered&sort=status&order=asc");

    await userEvent.selectOptions(registrationFilter().element(), "All");
    await expect.poll(() => router.state.location.searchStr).toBe("?sort=status&order=asc");
    await expect.poll(() => requests.at(-1)).toBe("?sort=status&order=asc");
    expect(router.history.length).toBe(historyLength + 3);
  });

  test("says when the data holder has not got the statuses", async () => {
    serve({ data: [] });

    renderApp("/register-statuses");

    await expect.element(table().getByText("No statuses")).toBeVisible();
    await expect
      .element(page.getByText("The data holder has not got the statuses from the Register."))
      .toBeVisible();
  });

  test("shows a problem when the API cannot answer", async () => {
    worker.use(http.untyped.get("/api/software-product-statuses", () => Response.error()));

    renderApp("/register-statuses");

    await expect
      .element(
        page.getByRole("alert").getByText("The console cannot get data from the data holder"),
      )
      .toBeVisible();
  });

  test("the side navigation goes to the Register statuses, and the item is current", async () => {
    serve({ data: [] });
    worker.use(
      http.get("/api/customers", ({ response }) =>
        response(200).json({ data: [], totalRecords: 0 }),
      ),
    );

    renderApp("/customers");
    const navigation = page.getByRole("navigation", { name: "Main" });
    await navigation.getByRole("link", { name: "Register statuses" }).click();

    await expect
      .element(page.getByRole("heading", { level: 1, name: "Register statuses" }))
      .toBeVisible();
    await expect
      .element(navigation.getByRole("link", { name: "Register statuses" }))
      .toHaveAttribute("aria-current", "page");
  });
});
