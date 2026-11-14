import { ConsumerButton, ConsumerCard } from "@cdr-lab/ui/consumer-page";
import { CopyButton } from "@cdr-lab/ui/copy-button";
import { Link } from "@tanstack/react-router";

import { useBrand, usePageTitle } from "../brand";
import { SignedIn } from "./session";

export function CustomerNumberPage() {
  usePageTitle("Your customer number");
  const { name } = useBrand();

  return (
    <SignedIn>
      {(customer) => (
        <ConsumerCard>
          <h1>Your customer number</h1>
          <p className="flex flex-wrap items-center gap-2">
            <span className="font-mono text-2xl font-semibold">{customer.customerNumber}</span>
            <CopyButton value={customer.customerNumber} label="customer number" />
          </p>
          <p>
            You sign in to {name} online banking with this customer number or with your email
            address.
          </p>
          <p>
            When an app asks for your {name} data under the Consumer Data Right, {name} asks for
            your email address or this customer number. {name} then sends a one-time password to
            your email address.
          </p>
          <ConsumerButton render={(props) => <Link {...props} to="/" />}>Continue</ConsumerButton>
        </ConsumerCard>
      )}
    </SignedIn>
  );
}
