import { Form, type FormRefusal } from "@cdr-lab/ui/form";
import { Select, TextField } from "@cdr-lab/ui/form-field";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { useToast } from "@cdr-lab/ui/toast";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink, useNavigate } from "@tanstack/react-router";

import { createAccount, problemOf, productCategoryValues } from "../api/client";
import { textOf } from "@cdr-lab/console/form-values";
import { refusalOf } from "@cdr-lab/console/problem";
import { customerQuery, nameOf } from "./CustomerPage";

const route = getRouteApi("/customers/$customerNumber/create-account");

const categoryOptions = productCategoryValues.map((category) => ({
  value: category,
  label: category,
}));

function AccountForm({ customerNumber }: { customerNumber: string }) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const toast = useToast();

  async function submit(data: FormData): Promise<FormRefusal | undefined> {
    const displayName = textOf(data, "displayName");
    const creditLimit = textOf(data, "creditLimit");
    try {
      await createAccount(customerNumber, {
        displayName,
        productCategory: textOf(data, "productCategory"),
        productName: textOf(data, "productName"),
        bsb: textOf(data, "bsb"),
        accountNumber: textOf(data, "accountNumber"),
        ...(creditLimit !== "" && { creditLimit }),
      });
    } catch (error) {
      return refusalOf(error);
    }
    // The list of the customers has the count of the open accounts.
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["customer", customerNumber] }),
      queryClient.invalidateQueries({ queryKey: ["customers"] }),
    ]);
    toast.show({ title: `Account ${displayName} created` });
    await navigate({
      to: "/customers/$customerNumber",
      params: { customerNumber },
      state: { focusTitle: true },
    });
    return undefined;
  }

  return (
    <Form
      label="Create account"
      submitLabel="Create account"
      cancel={<RouterLink to="/customers/$customerNumber" params={{ customerNumber }} />}
      onSubmit={submit}
    >
      <TextField label="Display name" name="displayName" autoComplete="off" />
      <Select
        label="Product category"
        name="productCategory"
        options={categoryOptions}
        placeholder="Select a product category"
        defaultValue=""
      />
      <TextField label="Product name" name="productName" autoComplete="off" />
      <TextField label="BSB" name="bsb" autoComplete="off" spellCheck={false} inputMode="numeric" />
      <TextField
        label="Account number"
        name="accountNumber"
        autoComplete="off"
        spellCheck={false}
        inputMode="numeric"
      />
      <TextField
        label="Credit limit (optional)"
        name="creditLimit"
        autoComplete="off"
        spellCheck={false}
        description="An amount in AUD, for example 500.00. With no value, the credit limit is 0.00."
      />
    </Form>
  );
}

export function CreateAccountPage() {
  const { customerNumber } = route.useParams();
  const detail = useQuery(customerQuery(customerNumber));

  return (
    <>
      <PageHeader
        title={`Create an account for ${detail.data ? nameOf(detail.data) : "a customer"}`}
        description="The data holder checks each value and creates the account ID. The account is open."
      />
      {detail.isPending && <LoadingState />}
      {detail.isError && <ErrorState problem={problemOf(detail.error)} />}
      {detail.data && <AccountForm key={customerNumber} customerNumber={customerNumber} />}
    </>
  );
}
