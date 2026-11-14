import { ProblemError } from "@cdr-lab/console/problem";
import { ConsumerButton, ConsumerCard } from "@cdr-lab/ui/consumer-page";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { useQuery } from "@tanstack/react-query";
import { getRouteApi, Link } from "@tanstack/react-router";
import type { ReactNode } from "react";

import { getProduct, problemOf, type ProductDetail } from "../api/client";
import type { components } from "../api/schema.gen";
import { usePageTitle } from "../brand";
import { NotFoundPage } from "../NotFoundPage";
import { Term, Terms } from "../terms";
import { BrandPhoto } from "../photo";
import { categoryLabels, categoryPhoto } from "./categories";
import { additionalValue, formatAmount, formatDate, formatFrequency, formatRate } from "./format";
import {
  applicationLabels,
  depositRateLabels,
  featureLabels,
  feeLabels,
  feeMethodLabels,
  feeRateLabels,
  lendingRateLabels,
  loanPurposeLabels,
  repaymentLabels,
} from "./labels";

type Schemas = components["schemas"];

const route = getRouteApi("/products/$productId");
const notFound = 404;

export function ProductPage() {
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
      {product.data && <Detail product={product.data} />}
    </ConsumerCard>
  );
}

function Detail({ product }: { product: ProductDetail }) {
  usePageTitle(product.name);
  const { depositRates = [], lendingRates = [], fees = [], features = [] } = product;

  return (
    <>
      <div className="product-head">
        <BrandPhoto file={categoryPhoto(product.productCategory)} className="product-photo" />
        <div>
          <p className="mb-1 text-sm font-semibold text-fg-muted">
            {categoryLabels[product.productCategory]}
          </p>
          <h1>{product.name}</h1>
          <p>{product.description}</p>
        </div>
      </div>
      <ConsumerButton
        render={(props) => (
          <Link
            {...props}
            to="/products/$productId/open"
            params={{ productId: product.productId }}
          />
        )}
      >
        Open an account
      </ConsumerButton>
      {product.isTailored && (
        <p className="rounded-md border border-info-border bg-info-bg px-4 py-3 text-info-fg">
          The rates and fees can be different for you.
        </p>
      )}
      {(product.effectiveFrom || product.effectiveTo) && (
        <Terms>
          {product.effectiveFrom && (
            <Term term="Available from">{formatDate(product.effectiveFrom)}</Term>
          )}
          {product.effectiveTo && (
            <Term term="Available until">{formatDate(product.effectiveTo)}</Term>
          )}
        </Terms>
      )}
      {(depositRates.length > 0 || lendingRates.length > 0) && (
        <Group heading="Interest rates">
          {depositRates.map((rate, index) => (
            // oxlint-disable-next-line react/no-array-index-key -- a rate has no identity, and the list does not change order.
            <DepositRateItem key={index} rate={rate} />
          ))}
          {lendingRates.map((rate, index) => (
            // oxlint-disable-next-line react/no-array-index-key -- a rate has no identity, and the list does not change order.
            <LendingRateItem key={index} rate={rate} />
          ))}
        </Group>
      )}
      {fees.length > 0 && (
        <Group heading="Fees">
          {fees.map((fee, index) => (
            // oxlint-disable-next-line react/no-array-index-key -- a fee has no identity, and the list does not change order.
            <FeeItem key={index} fee={fee} />
          ))}
        </Group>
      )}
      {features.length > 0 && (
        <Group heading="Features">
          {features.map((feature, index) => (
            // oxlint-disable-next-line react/no-array-index-key -- a feature has no identity, and the list does not change order.
            <FeatureItem key={index} feature={feature} />
          ))}
        </Group>
      )}
    </>
  );
}

function DepositRateItem({ rate }: { rate: Schemas["DepositRate"] }) {
  return (
    <Item heading={depositRateLabels[rate.depositRateType]}>
      <Term term="Rate">{formatRate(rate.rate)}</Term>
      <Term term="Interest is paid">{applicationLabels[rate.applicationType]}</Term>
      {rate.applicationFrequency && (
        <Term term="How often">{formatFrequency(rate.applicationFrequency)}</Term>
      )}
      {rate.additionalValue && (
        <Extra {...additionalValue("depositRate", rate.depositRateType, rate.additionalValue)} />
      )}
      {rate.additionalInfo && <Term term="More information">{rate.additionalInfo}</Term>}
    </Item>
  );
}

function LendingRateItem({ rate }: { rate: Schemas["LendingRate"] }) {
  return (
    <Item heading={lendingRateLabels[rate.lendingRateType]}>
      <Term term="Rate">{formatRate(rate.rate)}</Term>
      {rate.comparisonRate && <Term term="Comparison rate">{formatRate(rate.comparisonRate)}</Term>}
      <Term term="Interest is charged">{applicationLabels[rate.applicationType]}</Term>
      {rate.applicationFrequency && (
        <Term term="How often">{formatFrequency(rate.applicationFrequency)}</Term>
      )}
      <Term term="Repayments">{repaymentLabels[rate.repaymentType]}</Term>
      <Term term="Loan purpose">{loanPurposeLabels[rate.loanPurpose]}</Term>
      {rate.additionalValue && (
        <Extra {...additionalValue("lendingRate", rate.lendingRateType, rate.additionalValue)} />
      )}
      {rate.additionalInfo && <Term term="More information">{rate.additionalInfo}</Term>}
    </Item>
  );
}

function FeeItem({ fee }: { fee: Schemas["Fee"] }) {
  return (
    <Item heading={fee.name}>
      <Term term="Type">{feeLabels[fee.feeType]}</Term>
      <Term term="Worked out as">{feeMethodLabels[fee.feeMethodUType]}</Term>
      {fee.fixedAmount && <Term term="Amount">{formatAmount(fee.fixedAmount.amount)}</Term>}
      {fee.rateBased && (
        <Term term="Rate">
          {formatRate(fee.rateBased.rate)} of {feeRateLabels[fee.rateBased.rateType]}
        </Term>
      )}
      {fee.variable?.feeMinimum && (
        <Term term="Minimum">{formatAmount(fee.variable.feeMinimum)}</Term>
      )}
      {fee.variable?.feeMaximum && (
        <Term term="Maximum">{formatAmount(fee.variable.feeMaximum)}</Term>
      )}
      {fee.additionalValue && (
        <Extra {...additionalValue("fee", fee.feeType, fee.additionalValue)} />
      )}
      {fee.additionalInfo && <Term term="More information">{fee.additionalInfo}</Term>}
    </Item>
  );
}

function FeatureItem({ feature }: { feature: Schemas["Feature"] }) {
  return (
    <Item heading={featureLabels[feature.featureType]}>
      {feature.additionalValue && (
        <Extra {...additionalValue("feature", feature.featureType, feature.additionalValue)} />
      )}
      {feature.additionalInfo && <Term term="More information">{feature.additionalInfo}</Term>}
    </Item>
  );
}

function Group({ heading, children }: { heading: string; children: ReactNode }) {
  return (
    <section aria-label={heading}>
      <h2>{heading}</h2>
      <ul className="flex list-none flex-col gap-4 p-0">{children}</ul>
    </section>
  );
}

function Item({ heading, children }: { heading: string; children: ReactNode }) {
  return (
    <li>
      <h3>{heading}</h3>
      <Terms>{children}</Terms>
    </li>
  );
}

function Extra({ term, text }: { term: string; text: string }) {
  return <Term term={term}>{text}</Term>;
}
