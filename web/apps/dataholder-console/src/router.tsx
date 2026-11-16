import { AppShell, AppShellNavItem } from "@cdr-lab/ui/app-shell";
import { useQuery } from "@tanstack/react-query";
import {
  createRootRoute,
  createRoute,
  createRouter,
  Link,
  Outlet,
  useMatches,
  type RouterHistory,
} from "@tanstack/react-router";
import { AppWindow, LayoutDashboard, RadioTower, Users } from "lucide-react";
import { useEffect } from "react";

import { getBrand } from "./api/client";
import { CreateAccountPage } from "./customers/CreateAccountPage";
import { CreateCustomerPage } from "./customers/CreateCustomerPage";
import { CreateTransactionPage } from "./customers/CreateTransactionPage";
import { CustomerPage } from "./customers/CustomerPage";
import { CustomersPage } from "./customers/CustomersPage";
import { validateCustomerSearch } from "./customers/search";
import { TransactionsPage } from "./customers/TransactionsPage";
import { NotFoundPage } from "./NotFoundPage";
import { OverviewPage } from "./overview/OverviewPage";
import { RegistrationPage } from "./registrations/RegistrationPage";
import { RegistrationsPage } from "./registrations/RegistrationsPage";
import { validateRegistrationSearch } from "./registrations/search";
import { validateStatusSearch } from "./statuses/search";
import { StatusesPage } from "./statuses/StatusesPage";
import { parseSearch, stringifySearch } from "@cdr-lab/console/router-search";
import { validatePageSearch } from "@cdr-lab/console/list-search";

type Section = "overview" | "customers" | "registrations" | "register-statuses";

// The name of the console until the data holder gives the brand name, and when it cannot.
const fallbackName = "Data holder";

function Layout() {
  const section = useMatches({ select: (matches) => matches.at(-1)?.staticData.section });
  const brand = useQuery({
    queryKey: ["brand"],
    queryFn: ({ signal }) => getBrand(signal),
    staleTime: Infinity,
  });
  const name = brand.data?.brandName ?? fallbackName;

  useEffect(() => {
    document.title = name;
  }, [name]);

  return (
    <AppShell
      name={name}
      logo="/brand/logo.svg"
      label="Operator console"
      navigationLabel="Main"
      navigation={
        <>
          <AppShellNavItem
            icon={LayoutDashboard}
            current={section === "overview"}
            render={(props) => <Link {...props} to="/" />}
          >
            Overview
          </AppShellNavItem>
          <AppShellNavItem
            icon={Users}
            current={section === "customers"}
            render={(props) => <Link {...props} to="/customers" />}
          >
            Customers
          </AppShellNavItem>
          <AppShellNavItem
            icon={AppWindow}
            current={section === "registrations"}
            render={(props) => <Link {...props} to="/registrations" />}
          >
            Client registrations
          </AppShellNavItem>
          <AppShellNavItem
            icon={RadioTower}
            current={section === "register-statuses"}
            render={(props) => <Link {...props} to="/register-statuses" />}
          >
            Register statuses
          </AppShellNavItem>
        </>
      }
    >
      <Outlet />
    </AppShell>
  );
}

const rootRoute = createRootRoute({ component: Layout, notFoundComponent: NotFoundPage });

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  staticData: { section: "overview" },
  component: OverviewPage,
});

const customersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/customers",
  staticData: { section: "customers" },
  validateSearch: validateCustomerSearch,
  component: CustomersPage,
});

const customerRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/customers/$customerNumber",
  staticData: { section: "customers" },
  component: CustomerPage,
});

// The URL of a form is never the URL of a record (services/register/DESIGN.md, Operator console).
const createCustomerRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/create/customer",
  staticData: { section: "customers" },
  component: CreateCustomerPage,
});

const createAccountRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/customers/$customerNumber/create-account",
  staticData: { section: "customers" },
  component: CreateAccountPage,
});

const transactionsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/customers/$customerNumber/accounts/$accountId/transactions",
  staticData: { section: "customers" },
  validateSearch: validatePageSearch,
  component: TransactionsPage,
});

// The URL of a form is never the URL of a record (services/register/DESIGN.md, Operator console).
const createTransactionRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/customers/$customerNumber/accounts/$accountId/create-transaction",
  staticData: { section: "customers" },
  component: CreateTransactionPage,
});

const statusesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/register-statuses",
  staticData: { section: "register-statuses" },
  validateSearch: validateStatusSearch,
  component: StatusesPage,
});

const registrationsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/registrations",
  staticData: { section: "registrations" },
  validateSearch: validateRegistrationSearch,
  component: RegistrationsPage,
});

const registrationRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/registrations/$clientId",
  staticData: { section: "registrations" },
  validateSearch: validatePageSearch,
  component: RegistrationPage,
});

const routeTree = rootRoute.addChildren([
  indexRoute,
  customersRoute,
  customerRoute,
  createCustomerRoute,
  createAccountRoute,
  transactionsRoute,
  createTransactionRoute,
  registrationsRoute,
  registrationRoute,
  statusesRoute,
]);

export function createAppRouter(history?: RouterHistory) {
  return createRouter({
    routeTree,
    parseSearch,
    stringifySearch,
    ...(history && { history }),
  });
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof createAppRouter>;
  }
  interface StaticDataRouteOption {
    section?: Section;
  }
  interface HistoryState {
    // After a form, the detail moves the focus to its heading.
    focusTitle?: boolean;
  }
}
