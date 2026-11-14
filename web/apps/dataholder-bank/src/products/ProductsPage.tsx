import { ConsumerCard } from "@cdr-lab/ui/consumer-page";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { Link, useLocation } from "@tanstack/react-router";
import { useEffect } from "react";

import { problemOf } from "../api/client";
import { usePageTitle } from "../brand";
import { BrandPhoto } from "../photo";
import { categoryAnchor, categoryLabels, categoryPhoto, groupByCategory } from "./categories";
import { useProducts } from "./useProducts";

export function ProductsPage() {
  usePageTitle("Products");
  const products = useProducts();
  const hash = useLocation({ select: (location) => location.hash });

  // The link of a category on the home page goes to its group, which shows only when the products load.
  useEffect(() => {
    if (products.data && hash) document.getElementById(hash)?.scrollIntoView();
  }, [products.data, hash]);

  return (
    <ConsumerCard>
      <h1>Products</h1>
      {products.isPending && <LoadingState label="Loading the products" />}
      {products.isError && <ErrorState problem={problemOf(products.error)} />}
      {products.data &&
        groupByCategory(products.data.products).map(({ category, products: inCategory }) => (
          <section key={category} aria-labelledby={categoryAnchor(category)}>
            <div className="group-head">
              <BrandPhoto file={categoryPhoto(category)} className="group-photo" />
              <h2 id={categoryAnchor(category)} className="scroll-mt-4">
                {categoryLabels[category]}
              </h2>
            </div>
            <ul className="flex list-none flex-col gap-4 p-0">
              {inCategory.map((product) => (
                <li key={product.productId}>
                  <h3>
                    <Link
                      to="/products/$productId"
                      params={{ productId: product.productId }}
                      className="inline-flex min-h-control items-center rounded-sm text-link underline underline-offset-2 hover:decoration-2 focus-visible:outline-3 focus-visible:outline-offset-2 focus-visible:outline-focus"
                    >
                      {product.name}
                    </Link>
                  </h3>
                  <p className="mb-0 text-fg-muted">{product.description}</p>
                </li>
              ))}
            </ul>
          </section>
        ))}
      {products.data?.products.length === 0 && <p>The bank has no products now.</p>}
    </ConsumerCard>
  );
}
