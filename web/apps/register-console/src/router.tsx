import { AppShell, AppShellNavItem } from "@cdr-lab/ui/app-shell";
import {
  createRootRoute,
  createRoute,
  createRouter,
  Link,
  Outlet,
  redirect,
  useMatches,
  type RouterHistory,
} from "@tanstack/react-router";
import { AppWindow, FileBadge, KeyRound, Landmark } from "lucide-react";

import { CDRCertificatesPage } from "./cdr-certificates/CDRCertificatesPage";
import { AddBrandPage as AddDataHolderBrandPage } from "./data-holders/AddBrandPage";
import { ChangeBrandPage as ChangeDataHolderBrandPage } from "./data-holders/ChangeBrandPage";
import { ChangeDetailsPage as ChangeDataHolderDetailsPage } from "./data-holders/ChangeDetailsPage";
import { CreateDataHolderPage } from "./data-holders/CreateDataHolderPage";
import { DataHolderPage } from "./data-holders/DataHolderPage";
import { DataHoldersPage } from "./data-holders/DataHoldersPage";
import { validateDataHolderSearch } from "./data-holders/search";
import { AddBrandPage } from "./data-recipients/AddBrandPage";
import { AddSoftwareProductPage } from "./data-recipients/AddSoftwareProductPage";
import { ChangeBrandPage } from "./data-recipients/ChangeBrandPage";
import { ChangeDetailsPage } from "./data-recipients/ChangeDetailsPage";
import { ChangeSoftwareProductPage } from "./data-recipients/ChangeSoftwareProductPage";
import { CreateDataRecipientPage } from "./data-recipients/CreateDataRecipientPage";
import { DataRecipientPage } from "./data-recipients/DataRecipientPage";
import { DataRecipientsPage } from "./data-recipients/DataRecipientsPage";
import { validateDataRecipientSearch } from "./data-recipients/search";
import { DataHolderHistoryPage, DataRecipientHistoryPage } from "./history/HistoryPage";
import { KeySetsPage } from "./key-sets/KeySetsPage";
import { NotFoundPage } from "./NotFoundPage";
import { parseSearch, stringifySearch } from "@cdr-lab/app/router-search";
import { validatePageSearch } from "@cdr-lab/app/list-search";

type Section = "data-recipients" | "data-holders" | "signing-keys" | "cdr-certificates";

// A form page has a URL that is not under the URL of its section, so each route gives its section.
function Layout() {
  const section = useMatches({ select: (matches) => matches.at(-1)?.staticData.section });

  return (
    <AppShell
      name="CDR Register"
      navigationLabel="Main"
      navigation={
        <>
          <AppShellNavItem
            icon={AppWindow}
            current={section === "data-recipients"}
            render={(props) => <Link {...props} to="/data-recipients" />}
          >
            Data recipients
          </AppShellNavItem>
          <AppShellNavItem
            icon={Landmark}
            current={section === "data-holders"}
            render={(props) => <Link {...props} to="/data-holders" />}
          >
            Data holders
          </AppShellNavItem>
          <AppShellNavItem
            icon={KeyRound}
            current={section === "signing-keys"}
            render={(props) => <Link {...props} to="/signing-keys" />}
          >
            Signing keys
          </AppShellNavItem>
          <AppShellNavItem
            icon={FileBadge}
            current={section === "cdr-certificates"}
            render={(props) => <Link {...props} to="/cdr-certificates" />}
          >
            CDR certificates
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
  beforeLoad: () => {
    // oxlint-disable-next-line typescript/only-throw-error -- TanStack Router redirects with a thrown redirect.
    throw redirect({ to: "/data-recipients" });
  },
});

const dataRecipientsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-recipients",
  staticData: { section: "data-recipients" },
  validateSearch: validateDataRecipientSearch,
  component: DataRecipientsPage,
});

const dataHoldersRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-holders",
  staticData: { section: "data-holders" },
  validateSearch: validateDataHolderSearch,
  component: DataHoldersPage,
});

const dataRecipientRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-recipients/$legalEntityId",
  staticData: { section: "data-recipients" },
  component: DataRecipientPage,
});

// The URL of a form is never the URL of a record (services/register/design/operator-console-forms.md, REG-1405).
const changeDataRecipientDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-recipients/$legalEntityId/change",
  staticData: { section: "data-recipients" },
  component: ChangeDetailsPage,
});

// A path under /data-recipients/ can be the detail of a data recipient with that ID, for example "new".
const createDataRecipientRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/create/data-recipient",
  staticData: { section: "data-recipients" },
  component: CreateDataRecipientPage,
});

const addDataRecipientBrandRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-recipients/$legalEntityId/add-brand",
  staticData: { section: "data-recipients" },
  component: AddBrandPage,
});

const addSoftwareProductRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-recipients/$legalEntityId/brands/$brandId/add-software-product",
  staticData: { section: "data-recipients" },
  component: AddSoftwareProductPage,
});

const changeDataRecipientBrandRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-recipients/$legalEntityId/brands/$brandId/change",
  staticData: { section: "data-recipients" },
  component: ChangeBrandPage,
});

const changeSoftwareProductRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-recipients/$legalEntityId/software-products/$softwareProductId/change",
  staticData: { section: "data-recipients" },
  component: ChangeSoftwareProductPage,
});

const dataHolderRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-holders/$legalEntityId",
  staticData: { section: "data-holders" },
  component: DataHolderPage,
});

const changeDataHolderDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-holders/$legalEntityId/change",
  staticData: { section: "data-holders" },
  component: ChangeDataHolderDetailsPage,
});

const addDataHolderBrandRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-holders/$legalEntityId/add-brand",
  staticData: { section: "data-holders" },
  component: AddDataHolderBrandPage,
});

const changeDataHolderBrandRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-holders/$legalEntityId/brands/$brandId/change",
  staticData: { section: "data-holders" },
  component: ChangeDataHolderBrandPage,
});

const createDataHolderRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/create/data-holder",
  staticData: { section: "data-holders" },
  component: CreateDataHolderPage,
});

const dataRecipientHistoryRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-recipients/$legalEntityId/history",
  staticData: { section: "data-recipients" },
  validateSearch: validatePageSearch,
  component: DataRecipientHistoryPage,
});

const dataHolderHistoryRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-holders/$legalEntityId/history",
  staticData: { section: "data-holders" },
  validateSearch: validatePageSearch,
  component: DataHolderHistoryPage,
});

const keySetsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/signing-keys",
  staticData: { section: "signing-keys" },
  component: KeySetsPage,
});

const cdrCertificatesRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/cdr-certificates",
  staticData: { section: "cdr-certificates" },
  validateSearch: validatePageSearch,
  component: CDRCertificatesPage,
});

const routeTree = rootRoute.addChildren([
  indexRoute,
  dataRecipientsRoute,
  dataRecipientRoute,
  createDataRecipientRoute,
  changeDataRecipientDetailsRoute,
  addDataRecipientBrandRoute,
  changeDataRecipientBrandRoute,
  addSoftwareProductRoute,
  changeSoftwareProductRoute,
  dataRecipientHistoryRoute,
  dataHoldersRoute,
  dataHolderRoute,
  createDataHolderRoute,
  changeDataHolderDetailsRoute,
  addDataHolderBrandRoute,
  changeDataHolderBrandRoute,
  dataHolderHistoryRoute,
  keySetsRoute,
  cdrCertificatesRoute,
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
    // A form that sends the operator to a page asks the page to move the focus to its title.
    focusTitle?: boolean;
    // The ID of the record that the form changed or added, which the page opens.
    open?: string;
  }
}
