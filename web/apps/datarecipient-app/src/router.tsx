import { createRootRoute, createRouter, Outlet, type RouterHistory } from "@tanstack/react-router";

// The pages of the web app (services/datarecipient/design/web-app.md) are routes below the root.
const rootRoute = createRootRoute({
  component: () => (
    <main>
      <Outlet />
    </main>
  ),
});

export function createAppRouter(history?: RouterHistory) {
  return createRouter({ routeTree: rootRoute, ...(history && { history }) });
}
