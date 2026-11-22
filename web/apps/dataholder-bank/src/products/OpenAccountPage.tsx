import { textOf } from "@cdr-lab/app/form-values";
import { ProblemError, refusalOf } from "@cdr-lab/app/problem";
import { ConsumerCard } from "@cdr-lab/ui/consumer-page";
import { Form, type FormRefusal } from "@cdr-lab/ui/form";
import { TextField } from "@cdr-lab/ui/form-field";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { getRouteApi, Link, useNavigate } from "@tanstack/react-router";

import {
  getProduct,
  openAccount,
  problemOf,
  type OpenAccountRequest,
  type ProductDetail,
} from "../api/client";
import { useBrand, usePageTitle } from "../brand";
import { NotFoundPage } from "../NotFoundPage";
import { recordSessionEnded, SignedIn, sessionKey } from "../session/session";
import { categoryLabels } from "./categories";
import { amountStringOf, openValueOf, type OpenValue } from "./opening";

const route = getRouteApi("/products/$productId/open");
const notFound = 404;
const forbidden = 403;

export function OpenAccountPage() {
  return <SignedIn>{() => <OpenOfProduct />}</SignedIn>;
}

function OpenOfProduct() {
  const { productId } = route.useParams();
  const product = useQuery({
    queryKey: ["products", productId],
    queryFn: ({ signal }) => getProduct(productId, signal),
  });

  if (product.error instanceof ProblemError && product.error.status === notFound)
    return <NotFoundPage />;

  return (
    <ConsumerCard>
      {product.isPending && <LoadingState label="Loading the product" />}
      {product.isError && <ErrorState problem={problemOf(product.error)} />}
      {product.data && <Confirmation product={product.data} />}
    </ConsumerCard>
  );
}

// The page is the confirmation of the open, because an account cannot be removed (WCAG 2.2, SC 3.3.4).
function Confirmation({ product }: { product: ProductDetail }) {
  usePageTitle(`Open ${product.name}`);
  const { name } = useBrand();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const back = <Link to="/products/$productId" params={{ productId: product.productId }} />;

  const value = openValueOf(product.productCategory);

  async function submit(data: FormData): Promise<FormRefusal | undefined> {
    const body: OpenAccountRequest = { productId: product.productId };
    if (value) {
      const amount = amountStringOf(textOf(data, value.name));
      if (amount === undefined) return notAnAmount(value);
      body[value.name] = amount;
    }

    let opened;
    try {
      opened = await openAccount(body);
    } catch (error) {
      // The session ended, so the site reads it again and goes to the sign-in page.
      if (error instanceof ProblemError && error.status === forbidden) {
        recordSessionEnded(queryClient, error);
        await queryClient.invalidateQueries({ queryKey: sessionKey });
      }
      return refusalOf(error);
    }
    await queryClient.invalidateQueries({ queryKey: ["accounts"] });
    await navigate({
      to: "/accounts/$accountId",
      params: { accountId: opened.accountId },
      state: { status: `Your ${product.name} account is open.` },
    });
    return undefined;
  }

  return (
    <>
      <p className="mb-1 text-sm font-semibold text-fg-muted">
        {categoryLabels[product.productCategory]}
      </p>
      <h1>Open {product.name}</h1>
      <p>{name} opens an account of this product in your name.</p>
      {value && (
        <p>
          {name} is a demonstration bank, so it does no credit check. It approves the account
          immediately.
        </p>
      )}
      <Form
        label={`Open ${product.name}`}
        submitLabel="Open account"
        cancel={back}
        onSubmit={submit}
      >
        {value && (
          <TextField
            label={value.label}
            name={value.name}
            inputMode="decimal"
            autoComplete="off"
            description={value.description}
          />
        )}
      </Form>
    </>
  );
}

function notAnAmount({ name, label }: OpenValue): FormRefusal {
  const detail = `Enter the ${label.toLowerCase()} in dollars, for example 5,000 or 5,000.50.`;
  return {
    problem: { title: `Enter the ${label.toLowerCase()}`, detail },
    fieldErrors: [{ pointer: `#/${name}`, detail }],
  };
}
