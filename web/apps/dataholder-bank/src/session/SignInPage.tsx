import { refusalOf } from "@cdr-lab/web-platform/problem";
import { textOf } from "@cdr-lab/web-platform/form-values";
import { ConsumerCard } from "@cdr-lab/ui/consumer-page";
import { Form, type FormRefusal } from "@cdr-lab/ui/form";
import { TextField } from "@cdr-lab/ui/form-field";
import { Link as TextLink } from "@cdr-lab/ui/link";
import { useQueryClient } from "@tanstack/react-query";
import { Link, useNavigate } from "@tanstack/react-router";

import { signIn } from "../api/client";
import { useBrand, usePageTitle } from "../brand";
import {
  clearSessionEnded,
  isSessionEnded,
  sessionKey,
  useSession,
  useSessionEndedRecord,
} from "./session";

// The page requires a password, but the site does not send it (services/dataholder/design/online-bank-site-sign-in.md,
// DH-2021 and DH-2022).
const noPassword: FormRefusal = {
  problem: { title: "Enter a password", detail: "Enter a password. Any password works." },
  fieldErrors: [{ pointer: "#/password", detail: "Enter a password. Any password works." }],
};

export function SignInPage() {
  usePageTitle("Sign in");
  const { name } = useBrand();
  const session = useSession();
  const endedRecord = useSessionEndedRecord();
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  async function submit(data: FormData): Promise<FormRefusal | undefined> {
    if (textOf(data, "password") === "") return noPassword;
    try {
      await signIn(textOf(data, "identifier"));
    } catch (error) {
      return refusalOf(error);
    }
    clearSessionEnded(queryClient);
    await queryClient.invalidateQueries({ queryKey: sessionKey });
    await navigate({ to: "/accounts" });
    return undefined;
  }

  return (
    <ConsumerCard>
      <h1>Sign in to online banking</h1>
      {(endedRecord || isSessionEnded(session.error)) && (
        <output className="block font-semibold">{name} signed you out. Sign in again.</output>
      )}
      {/* A note that looks different from the lock note of the CDR pages (design/online-bank-site-sign-in.md,
          DH-2054). */}
      <div className="rounded-md border-2 border-border-strong p-4">
        <p className="m-0">
          This is {name} online banking. This is a demonstration bank: any password works here, and
          you must not use a real password. When an app asks for your {name} data under the Consumer
          Data Right, you do not sign in here. {name} then asks only for your email address or
          customer number and a one-time password, and never for your password.
        </p>
      </div>
      <Form label="Sign in" submitLabel="Sign in" cancel={<Link to="/" />} onSubmit={submit}>
        <TextField
          label="Email address or customer number"
          name="identifier"
          autoComplete="username"
          autoCapitalize="none"
          spellCheck={false}
        />
        <TextField
          label="Password"
          name="password"
          type="password"
          autoComplete="current-password"
        />
      </Form>
      <p>
        New to {name}? <TextLink render={<Link to="/sign-up" />}>Sign up</TextLink>
      </p>
    </ConsumerCard>
  );
}
