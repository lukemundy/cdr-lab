import { validatePageSearch } from "@cdr-lab/web-platform/list-search";
import { ConsumerFooter, ConsumerNavItem, PasswordStatement } from "@cdr-lab/ui/consumer-page";
import { useQueryClient } from "@tanstack/react-query";
import {
  createRootRoute,
  createRoute,
  createRouter,
  Link,
  Outlet,
  type RouterHistory,
  useNavigate,
} from "@tanstack/react-router";

import { signOut } from "./api/client";
import { AccountPage } from "./accounts/AccountPage";
import { AccountsPage } from "./accounts/AccountsPage";
import { ConsentsPage } from "./consents/ConsentsPage";
import { useBrand } from "./brand";
import { HomePage } from "./HomePage";
import { NotFoundPage } from "./NotFoundPage";
import { BrandPhoto } from "./photo";
import { OpenAccountPage } from "./products/OpenAccountPage";
import { ProductPage } from "./products/ProductPage";
import { ProductsPage } from "./products/ProductsPage";
import { CustomerNumberPage } from "./session/CustomerNumberPage";
import { sessionKey, useSession } from "./session/session";
import { SignInPage } from "./session/SignInPage";
import { SignUpPage } from "./session/SignUpPage";
import { BankLogo, BankShell, WithPhoto } from "./shell";

function Layout() {
  const { demonstrationNotice } = useBrand();

  return (
    <BankShell
      home={
        <Link to="/" className="bank-home">
          <BankLogo />
        </Link>
      }
      navigation={
        <>
          <ConsumerNavItem render={(props) => <Link {...props} to="/products" />}>
            Products
          </ConsumerNavItem>
          <SessionNavItem />
        </>
      }
      footer={
        <ConsumerFooter>
          {/* CDS: Common Authentication Standards. A consumer can open each page first, so each page says it. */}
          <PasswordStatement>
            Third parties never need your password to get your data under the Consumer Data Right.
          </PasswordStatement>
          {demonstrationNotice && <p>{demonstrationNotice}</p>}
        </ConsumerFooter>
      }
    >
      <Outlet />
    </BankShell>
  );
}

// Each page shows if the consumer is signed in: the accounts, the data sharing and a sign-out control, or a link to
// the sign-in page.
function SessionNavItem() {
  const session = useSession();
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  if (!session.isSuccess) {
    return (
      <ConsumerNavItem render={(props) => <Link {...props} to="/sign-in" />}>
        Sign in
      </ConsumerNavItem>
    );
  }

  // After a failed sign-out, the session can still be valid, and the new query of the session shows it.
  async function end() {
    const signedOut = await signOut().then(
      () => true,
      () => false,
    );
    if (signedOut) await navigate({ to: "/sign-in" });
    await queryClient.resetQueries({ queryKey: sessionKey });
  }

  return (
    <>
      <ConsumerNavItem render={(props) => <Link {...props} to="/accounts" />}>
        Accounts
      </ConsumerNavItem>
      {/* CDR Rules, rule 1.15(1)(f): the dashboard is prominently displayed and readily accessible. */}
      <ConsumerNavItem render={(props) => <Link {...props} to="/data-sharing" />}>
        Data sharing
      </ConsumerNavItem>
      <ConsumerNavItem
        render={(props) => <button {...props} type="button" onClick={() => void end()} />}
      >
        Sign out
      </ConsumerNavItem>
    </>
  );
}

const rootRoute = createRootRoute({ component: Layout, notFoundComponent: NotFoundPage });

const homeRoute = createRoute({ getParentRoute: () => rootRoute, path: "/", component: HomePage });

const productsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/products",
  component: ProductsPage,
});

const productRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/products/$productId",
  component: ProductPage,
});

const openAccountRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/products/$productId/open",
  component: OpenAccountPage,
});

// The sign-up and the sign-in show the hero photo of the brand next to the form.
const heroPhoto = <BrandPhoto file="hero.webp" eager />;

const signUpRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/sign-up",
  component: () => (
    <WithPhoto photo={heroPhoto}>
      <SignUpPage />
    </WithPhoto>
  ),
});

const customerNumberRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/customer-number",
  component: CustomerNumberPage,
});

const signInRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/sign-in",
  component: () => (
    <WithPhoto photo={heroPhoto}>
      <SignInPage />
    </WithPhoto>
  ),
});

const accountsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/accounts",
  component: AccountsPage,
});

const accountRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/accounts/$accountId",
  validateSearch: validatePageSearch,
  component: AccountPage,
});

const consentsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/data-sharing",
  component: ConsentsPage,
});

const routeTree = rootRoute.addChildren([
  homeRoute,
  productsRoute,
  productRoute,
  openAccountRoute,
  signUpRoute,
  customerNumberRoute,
  signInRoute,
  accountsRoute,
  accountRoute,
  consentsRoute,
]);

export function createAppRouter(history?: RouterHistory) {
  return createRouter({ routeTree, ...(history && { history }) });
}

declare module "@tanstack/react-router" {
  interface Register {
    router: ReturnType<typeof createAppRouter>;
  }
}
