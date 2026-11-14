import { ConsumerButton } from "@cdr-lab/ui/consumer-page";
import { ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { Link } from "@tanstack/react-router";

import { problemOf } from "./api/client";
import { useBrand, usePageTitle } from "./brand";
import { BrandPhoto } from "./photo";
import {
  categoryAnchor,
  categoryLabels,
  categoryPhoto,
  type CategoryGroup,
  groupByCategory,
  productCount,
} from "./products/categories";
import { useProducts } from "./products/useProducts";

export function HomePage() {
  usePageTitle("Home");
  const { name } = useBrand();

  return (
    <div className="home">
      <section className="hero" aria-labelledby="home-heading">
        <div className="hero-text">
          <h1 id="home-heading">Banking with {name}</h1>
          <p>See the products that {name} offers now, with their rates, fees and features.</p>
          <ConsumerButton render={(props) => <Link {...props} to="/products" />}>
            See all products
          </ConsumerButton>
        </div>
        <BrandPhoto file="hero.webp" className="hero-photo" eager />
      </section>
      <section className="types" aria-labelledby="types-heading">
        <h2 id="types-heading">Products by type</h2>
        <Categories />
      </section>
      <section className="sharing" aria-labelledby="sharing-heading">
        <h2 id="sharing-heading">Sharing your data</h2>
        <p>
          {name} shares your data under the Consumer Data Right only when you give your consent.
        </p>
      </section>
    </div>
  );
}

// The updraft layout shows everyday and savings accounts first, in its large tile (services/dataholder/DESIGN.md,
// Online bank site: Presentation). The other groups stay in the order of the CDS.
const isEveryday = (group: CategoryGroup) => group.category === "TRANS_AND_SAVINGS_ACCOUNTS";

function inLayoutOrder(groups: CategoryGroup[], layout: string | undefined): CategoryGroup[] {
  if (layout !== "updraft") return groups;
  return [...groups.filter(isEveryday), ...groups.filter((group) => !isEveryday(group))];
}

function Categories() {
  const products = useProducts();
  const { layout } = useBrand();

  if (products.isPending) return <LoadingState label="Loading the products" />;
  if (products.isError) return <ErrorState problem={problemOf(products.error)} />;

  return (
    <ul className="type-list">
      {inLayoutOrder(groupByCategory(products.data.products), layout).map(
        ({ category, products: inCategory }) => (
          <li key={category}>
            <Link to="/products" hash={categoryAnchor(category)} className="type-card">
              <BrandPhoto file={categoryPhoto(category)} className="type-photo" />
              <span className="type-body">
                <span className="type-name">{categoryLabels[category]}</span>
                <span className="type-count">{productCount(inCategory.length)}</span>
              </span>
            </Link>
          </li>
        ),
      )}
    </ul>
  );
}
