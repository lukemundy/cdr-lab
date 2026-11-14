import { ConsumerButton, ConsumerCard } from "@cdr-lab/ui/consumer-page";
import { Link } from "@tanstack/react-router";

import { usePageTitle } from "./brand";

export function NotFoundPage() {
  usePageTitle("Page not found");

  return (
    <ConsumerCard>
      <h1>Page not found</h1>
      <p>This page does not exist, or it is not available to you now.</p>
      <ConsumerButton render={(props) => <Link {...props} to="/products" />}>
        See all products
      </ConsumerButton>
    </ConsumerCard>
  );
}
