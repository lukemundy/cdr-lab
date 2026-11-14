import { refusalOf } from "@cdr-lab/console/problem";
import { textOf } from "@cdr-lab/console/form-values";
import { ConsumerCard } from "@cdr-lab/ui/consumer-page";
import { Form, type FormRefusal } from "@cdr-lab/ui/form";
import { TextField } from "@cdr-lab/ui/form-field";
import { Link as TextLink } from "@cdr-lab/ui/link";
import { useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate } from "@tanstack/react-router";

import { signUp } from "../api/client";
import { useBrand, usePageTitle } from "../brand";
import { sessionKey } from "./session";

export function SignUpPage() {
  usePageTitle("Sign up");
  const { name } = useBrand();
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  async function submit(data: FormData): Promise<FormRefusal | undefined> {
    const firstName = textOf(data, "firstName");
    try {
      const customer = await signUp({
        ...(firstName !== "" && { firstName }),
        lastName: textOf(data, "lastName"),
        emailAddress: textOf(data, "emailAddress"),
      });
      // A sign-up starts a session, so the customer of the session is the new customer.
      queryClient.setQueryData(sessionKey, customer);
    } catch (error) {
      return refusalOf(error);
    }
    await navigate({ to: "/customer-number" });
    return undefined;
  }

  return (
    <ConsumerCard>
      <h1>Become a customer of {name}</h1>
      <p>
        {name} is a demonstration bank, not a real bank. Give only invented details. The bank
        creates your customer number, and asks for no password and no identity document.
      </p>
      <Form label="Sign up" submitLabel="Sign up" cancel={<Link to="/" />} onSubmit={submit}>
        <TextField label="First name (optional)" name="firstName" autoComplete="given-name" />
        <TextField label="Last name" name="lastName" autoComplete="family-name" />
        <TextField
          label="Email address"
          name="emailAddress"
          type="email"
          autoComplete="email"
          spellCheck={false}
          description="Use an invented address, for example name@example.test. The bank sends each one-time password to this address."
        />
      </Form>
      <p>
        Already a customer? <TextLink render={<Link to="/sign-in" />}>Sign in</TextLink>
      </p>
    </ConsumerCard>
  );
}
