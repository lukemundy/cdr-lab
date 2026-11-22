import { DateTimeField, utcOf } from "@cdr-lab/ui/date-time-field";
import { Fieldset } from "@cdr-lab/ui/fieldset";
import { Form, type FormRefusal } from "@cdr-lab/ui/form";
import { Select, TextField } from "@cdr-lab/ui/form-field";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { useToast } from "@cdr-lab/ui/toast";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getRouteApi, Link as RouterLink } from "@tanstack/react-router";

import {
  createTransaction,
  extensionUTypeValues,
  nPPServiceValues,
  transactionStatusValues,
  transactionTypeValues,
  type TransactionCreate,
} from "../api/client";
import { textOf } from "@cdr-lab/app/form-values";
import { refusalOf } from "@cdr-lab/app/problem";
import { customerQuery } from "./CustomerPage";

const route = getRouteApi("/customers/$customerNumber/accounts/$accountId/create-transaction");

const optionsOf = (values: readonly string[]) => values.map((value) => ({ value, label: value }));

function choiceOf<T extends string>(values: readonly T[], data: FormData, name: string) {
  return values.find((value) => value === textOf(data, name));
}

const optionalTexts = [
  "reference",
  "merchantName",
  "merchantCategoryCode",
  "billerCode",
  "billerName",
  "crn",
  "apcaNumber",
] as const;
const times = ["postingDateTime", "executionDateTime", "valueDateTime"] as const;
const extendedNames = [
  "extendedData/payer",
  "extendedData/payee",
  "extendedData/extensionUType",
  "extendedData/nppPayload/extendedDescription",
  "extendedData/nppPayload/service",
  "extendedData/nppPayload/serviceVersion",
];

function extendedDataOf(data: FormData): TransactionCreate["extendedData"] {
  if (extendedNames.every((name) => textOf(data, name) === "")) return undefined;
  const payer = textOf(data, "extendedData/payer");
  const payee = textOf(data, "extendedData/payee");
  return {
    ...(payer !== "" && { payer }),
    ...(payee !== "" && { payee }),
    extensionUType: choiceOf(extensionUTypeValues, data, "extendedData/extensionUType"),
    nppPayload: {
      extendedDescription: textOf(data, "extendedData/nppPayload/extendedDescription"),
      service: choiceOf(nPPServiceValues, data, "extendedData/nppPayload/service"),
      serviceVersion: textOf(data, "extendedData/nppPayload/serviceVersion"),
    },
  };
}

// An optional field or group with no value is not in the request.
function transactionOf(data: FormData): TransactionCreate {
  const create: TransactionCreate = {
    type: choiceOf(transactionTypeValues, data, "type"),
    status: choiceOf(transactionStatusValues, data, "status"),
    description: textOf(data, "description"),
    amount: textOf(data, "amount"),
  };
  for (const name of optionalTexts) {
    const value = textOf(data, name);
    if (value !== "") create[name] = value;
  }
  for (const name of times) {
    const value = utcOf(textOf(data, name));
    if (value !== "") create[name] = value;
  }
  const extendedData = extendedDataOf(data);
  if (extendedData) create.extendedData = extendedData;
  return create;
}

const inBrowserTime = "In the time zone of this browser.";

export function CreateTransactionPage() {
  const { customerNumber, accountId } = route.useParams();
  const navigate = route.useNavigate();
  const queryClient = useQueryClient();
  const toast = useToast();
  const detail = useQuery(customerQuery(customerNumber));
  const name = detail.data?.accounts.find((a) => a.accountId === accountId)?.displayName;
  const transactions = (
    <RouterLink
      to="/customers/$customerNumber/accounts/$accountId/transactions"
      params={{ customerNumber, accountId }}
    />
  );

  async function submit(data: FormData): Promise<FormRefusal | undefined> {
    const create = transactionOf(data);
    try {
      await createTransaction(customerNumber, accountId, create);
    } catch (error) {
      return refusalOf(error);
    }
    // The balances of the account change with the transaction.
    await Promise.all([
      queryClient.invalidateQueries({ queryKey: ["transactions", customerNumber, accountId] }),
      queryClient.invalidateQueries({ queryKey: ["customer", customerNumber] }),
    ]);
    toast.show({ title: `Transaction ${create.description} created` });
    await navigate({
      to: "/customers/$customerNumber/accounts/$accountId/transactions",
      params: { customerNumber, accountId },
      state: { focusTitle: true },
    });
    return undefined;
  }

  return (
    <>
      <PageHeader
        title="Create a transaction"
        description={`${name ? `For ${name}. ` : ""}The data holder checks each value and creates the transaction ID.`}
      />
      <Form
        label="Create transaction"
        submitLabel="Create transaction"
        cancel={transactions}
        onSubmit={submit}
      >
        <Select
          label="Type"
          name="type"
          placeholder="Choose a type"
          options={optionsOf(transactionTypeValues)}
          defaultValue=""
        />
        <Select
          label="Status"
          name="status"
          placeholder="Choose a status"
          options={optionsOf(transactionStatusValues)}
          defaultValue=""
        />
        <TextField label="Description" name="description" autoComplete="off" />
        <TextField
          label="Amount"
          name="amount"
          autoComplete="off"
          spellCheck={false}
          description="In AUD, with 2 decimal digits. A debit has a minus sign, for example -12.50."
        />
        <TextField label="Reference (optional)" name="reference" autoComplete="off" />
        <DateTimeField
          label="Posting time (optional)"
          name="postingDateTime"
          description={`${inBrowserTime} A POSTED transaction must have it.`}
        />
        <DateTimeField
          label="Execution time (optional)"
          name="executionDateTime"
          description={`${inBrowserTime} A PENDING transaction must have it.`}
        />
        <DateTimeField
          label="Value time (optional)"
          name="valueDateTime"
          description={inBrowserTime}
        />
        <TextField label="Merchant name (optional)" name="merchantName" autoComplete="off" />
        <TextField
          label="Merchant category code (optional)"
          name="merchantCategoryCode"
          autoComplete="off"
        />
        <TextField label="Biller code (optional)" name="billerCode" autoComplete="off" />
        <TextField label="Biller name (optional)" name="billerName" autoComplete="off" />
        <TextField label="CRN (optional)" name="crn" autoComplete="off" />
        <TextField label="APCA number (optional)" name="apcaNumber" autoComplete="off" />
        <Fieldset legend="Extended data (optional)">
          <TextField
            label="Payer (optional)"
            name="extendedData/payer"
            autoComplete="off"
            description="For a positive amount."
          />
          <TextField
            label="Payee (optional)"
            name="extendedData/payee"
            autoComplete="off"
            description="For a negative amount."
          />
          <Select
            label="Extension type"
            name="extendedData/extensionUType"
            placeholder="Choose a type"
            options={optionsOf(extensionUTypeValues)}
            defaultValue=""
          />
          <Select
            label="NPP service"
            name="extendedData/nppPayload/service"
            placeholder="Choose a service"
            options={optionsOf(nPPServiceValues)}
            defaultValue=""
          />
          <TextField
            label="NPP service version"
            name="extendedData/nppPayload/serviceVersion"
            autoComplete="off"
          />
          <TextField
            label="Extended description"
            name="extendedData/nppPayload/extendedDescription"
            autoComplete="off"
          />
        </Fieldset>
      </Form>
    </>
  );
}
