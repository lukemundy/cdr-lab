import { Button } from "@cdr-lab/ui/button";
import { render } from "@cdr-lab/ui/test/render";
import { ToastProvider } from "@cdr-lab/ui/toast";
import { useState } from "react";
import { describe, expect, test } from "vitest";
import { page, userEvent } from "vitest/browser";

import { ProblemError } from "./problem";
import { RecordActions, type ActionDialog } from "./record-actions";

type Action = { action: "close"; count: number };

const conflict = {
  title: "Conflict",
  detail: "The record is closed. The service changed nothing.",
};

function closeDialog(state: string, action: Action): ActionDialog {
  return {
    trigger: <Button variant="danger">Close</Button>,
    title: "Close Everyday?",
    name: "Close",
    effects: [`Everyday changes from ${state} to CLOSED (${action.count}).`],
    irreversible: true,
    result: "Everyday is now CLOSED",
  };
}

// The record has the close action until the refresh after the first request.
function Record({ refuse, requests }: { refuse: boolean; requests: Action[] }) {
  const [actions, setActions] = useState<Action[]>([{ action: "close", count: 1 }]);
  const [state, setState] = useState("OPEN");

  return (
    <ToastProvider>
      <p>State: {state}</p>
      <RecordActions
        label="Actions of Everyday"
        actions={actions}
        dialogOf={(action) => closeDialog(state, action)}
        run={async (action) => {
          requests.push(action);
          if (refuse) throw new ProblemError(conflict, 409);
        }}
        refresh={async () => {
          setActions([]);
          setState("CLOSED");
        }}
      />
    </ToastProvider>
  );
}

const group = () => page.getByRole("group", { name: "Actions of Everyday", exact: true });
const dialog = () => page.getByRole("alertdialog", { name: "Close Everyday?" });

describe("the actions of a record", () => {
  // Verifies: WCON-0015, WCON-0016, WCON-0018
  test("a confirmed action runs one time, refreshes the views and shows the result in a toast", async () => {
    const requests: Action[] = [];
    render(<Record refuse={false} requests={requests} />);

    await userEvent.click(group().getByRole("button", { name: "Close" }));
    await expect
      .element(dialog())
      .toHaveAccessibleDescription(
        "Everyday changes from OPEN to CLOSED (1). You cannot reverse this action.",
      );
    await dialog().getByRole("button", { name: "Close", exact: true }).click();

    await expect.element(dialog()).not.toBeInTheDocument();
    await expect.element(page.getByText("State: CLOSED")).toBeVisible();
    await expect
      .element(page.getByRole("dialog", { name: "Everyday is now CLOSED" }))
      .toBeVisible();
    expect(requests).toEqual([{ action: "close", count: 1 }]);
    // The action does not exist now, so the focus goes to the group.
    await expect.poll(() => document.activeElement).toBe(group().element());
  });

  // Verifies: WCON-0014, WCON-0015, WCON-0017
  test("a refused action keeps its dialog as it opened, with the problem, after the refresh removes the action", async () => {
    const requests: Action[] = [];
    render(<Record refuse requests={requests} />);

    await userEvent.click(group().getByRole("button", { name: "Close" }));
    await dialog().getByRole("button", { name: "Close", exact: true }).click();

    await expect.element(dialog().getByRole("alert")).toMatchTextContent(conflict.detail);
    await expect.element(page.getByText("State: CLOSED")).toBeVisible();
    await expect
      .element(dialog())
      .toHaveAccessibleDescription(
        "Everyday changes from OPEN to CLOSED (1). You cannot reverse this action.",
      );
    expect(page.getByRole("dialog", { name: /is now/ }).query()).toBeNull();
    expect(requests).toHaveLength(1);

    await dialog().getByRole("button", { name: "Cancel" }).click();

    await expect.element(dialog()).not.toBeInTheDocument();
    expect(group().getByRole("button").query()).toBeNull();
  });
});
