import { DescriptionItem, DescriptionList } from "@cdr-lab/ui/description-list";
import { ProblemError } from "@cdr-lab/app/problem";
import { ConsumerButton, ConsumerCard } from "@cdr-lab/ui/consumer-page";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { useQuery } from "@tanstack/react-query";
import { getRouteApi, Link } from "@tanstack/react-router";
import type { ReactNode } from "react";

import { getProduct, problemOf, type ProductDetail } from "../api/client";
import type { components } from "../api/schema.gen";
import { usePageTitle } from "../brand";
import { NotFoundPage } from "../NotFoundPage";

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
        <DescriptionList>
          {product.effectiveFrom && (
            <DescriptionItem term="Available from">
              {formatDate(product.effectiveFrom)}
            </DescriptionItem>
          )}
          {product.effectiveTo && (
            <DescriptionItem term="Available until">
              {formatDate(product.effectiveTo)}
            </DescriptionItem>
          )}
        </DescriptionList>
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
      <DescriptionItem term="Rate">{formatRate(rate.rate)}</DescriptionItem>
      <DescriptionItem term="Interest is paid">
        {applicationLabels[rate.applicationType]}
      </DescriptionItem>
      {rate.applicationFrequency && (
        <DescriptionItem term="How often">
          {formatFrequency(rate.applicationFrequency)}
        </DescriptionItem>
      )}
      {rate.additionalValue && (
        <Extra {...additionalValue("depositRate", rate.depositRateType, rate.additionalValue)} />
      )}
      {rate.additionalInfo && (
        <DescriptionItem term="More information">{rate.additionalInfo}</DescriptionItem>
      )}
    </Item>
  );
}

function LendingRateItem({ rate }: { rate: Schemas["LendingRate"] }) {
  return (
    <Item heading={lendingRateLabels[rate.lendingRateType]}>
      <DescriptionItem term="Rate">{formatRate(rate.rate)}</DescriptionItem>
      {rate.comparisonRate && (
        <DescriptionItem term="Comparison rate">{formatRate(rate.comparisonRate)}</DescriptionItem>
      )}
      <DescriptionItem term="Interest is charged">
        {applicationLabels[rate.applicationType]}
      </DescriptionItem>
      {rate.applicationFrequency && (
        <DescriptionItem term="How often">
          {formatFrequency(rate.applicationFrequency)}
        </DescriptionItem>
      )}
      <DescriptionItem term="Repayments">{repaymentLabels[rate.repaymentType]}</DescriptionItem>
      <DescriptionItem term="Loan purpose">{loanPurposeLabels[rate.loanPurpose]}</DescriptionItem>
      {rate.additionalValue && (
        <Extra {...additionalValue("lendingRate", rate.lendingRateType, rate.additionalValue)} />
      )}
      {rate.additionalInfo && (
        <DescriptionItem term="More information">{rate.additionalInfo}</DescriptionItem>
      )}
    </Item>
  );
}

function FeeItem({ fee }: { fee: Schemas["Fee"] }) {
  return (
    <Item heading={fee.name}>
      <DescriptionItem term="Type">{feeLabels[fee.feeType]}</DescriptionItem>
      <DescriptionItem term="Worked out as">{feeMethodLabels[fee.feeMethodUType]}</DescriptionItem>
      {fee.fixedAmount && (
        <DescriptionItem term="Amount">{formatAmount(fee.fixedAmount.amount)}</DescriptionItem>
      )}
      {fee.rateBased && (
        <DescriptionItem term="Rate">
          {formatRate(fee.rateBased.rate)} of {feeRateLabels[fee.rateBased.rateType]}
        </DescriptionItem>
      )}
      {fee.variable?.feeMinimum && (
        <DescriptionItem term="Minimum">{formatAmount(fee.variable.feeMinimum)}</DescriptionItem>
      )}
      {fee.variable?.feeMaximum && (
        <DescriptionItem term="Maximum">{formatAmount(fee.variable.feeMaximum)}</DescriptionItem>
      )}
      {fee.additionalValue && (
        <Extra {...additionalValue("fee", fee.feeType, fee.additionalValue)} />
      )}
      {fee.additionalInfo && (
        <DescriptionItem term="More information">{fee.additionalInfo}</DescriptionItem>
      )}
    </Item>
  );
}

function FeatureItem({ feature }: { feature: Schemas["Feature"] }) {
  return (
    <Item heading={featureLabels[feature.featureType]}>
      {feature.additionalValue && (
        <Extra {...additionalValue("feature", feature.featureType, feature.additionalValue)} />
      )}
      {feature.additionalInfo && (
        <DescriptionItem term="More information">{feature.additionalInfo}</DescriptionItem>
      )}
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
      <DescriptionList>{children}</DescriptionList>
    </li>
  );
}

function Extra({ term, text }: { term: string; text: string }) {
  return <DescriptionItem term={term}>{text}</DescriptionItem>;
}
