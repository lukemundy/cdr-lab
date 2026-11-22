import { Form, type FormRefusal } from "@cdr-lab/ui/form";
import { TextField } from "@cdr-lab/ui/form-field";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { useToast } from "@cdr-lab/ui/toast";
import { useQueryClient } from "@tanstack/react-query";
import { Link as RouterLink, useNavigate } from "@tanstack/react-router";

import { createCustomer } from "../api/client";
import { textOf } from "@cdr-lab/app/form-values";
import { refusalOf } from "@cdr-lab/app/problem";

export function CreateCustomerPage() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const toast = useToast();

  async function submit(data: FormData): Promise<FormRefusal | undefined> {
    const firstName = textOf(data, "firstName");
    let customerNumber: string;
    try {
      customerNumber = await createCustomer({
        ...(firstName !== "" && { firstName }),
        lastName: textOf(data, "lastName"),
        emailAddress: textOf(data, "emailAddress"),
      });
    } catch (error) {
      return refusalOf(error);
    }
    await queryClient.invalidateQueries({ queryKey: ["customers"] });
    toast.show({ title: `Customer ${customerNumber} created` });
    await navigate({
      to: "/customers/$customerNumber",
      params: { customerNumber },
      state: { focusTitle: true },
    });
    return undefined;
  }

  return (
    <>
      <PageHeader
        title="Create a customer"
        description="The data holder checks each value and creates the customer number."
      />
      <Form
        label="Create customer"
        submitLabel="Create customer"
        cancel={<RouterLink to="/customers" />}
        onSubmit={submit}
      >
        <TextField label="First name (optional)" name="firstName" autoComplete="off" />
        <TextField label="Last name" name="lastName" autoComplete="off" />
        <TextField
          label="Email address"
          name="emailAddress"
          autoComplete="off"
          spellCheck={false}
          description="The data holder sends each one-time password of the customer to this address."
        />
      </Form>
    </>
  );
}
