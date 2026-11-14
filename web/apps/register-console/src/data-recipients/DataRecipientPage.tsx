import { ActionRow } from "@cdr-lab/ui/action-row";
import { DescriptionItem, DescriptionList } from "@cdr-lab/ui/description-list";
import { Identifier } from "@cdr-lab/ui/identifier";
import { Link } from "@cdr-lab/ui/link";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { Section } from "@cdr-lab/ui/section";
import { EmptyState, ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { StatusBadge } from "@cdr-lab/ui/status-badge";
import { LocalTime } from "@cdr-lab/ui/local-time";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { History, Pencil, Plus } from "lucide-react";
import { getRouteApi, Link as RouterLink, useLocation } from "@tanstack/react-router";

import {
  changeDataRecipientBrandStatus,
  changeDataRecipientStatus,
  changeSoftwareProductStatus,
  getDataRecipient,
  issueClientCertificate,
  problemOf,
  registerClientCertificate,
  removeClientCertificate,
  type DataRecipientBrand,
  type DataRecipientDetail,
  type SoftwareProduct,
} from "../api/client";
import { brandAndProductTones, dataRecipientTones } from "../statuses";
import { CertificateList, RegisterCertificate } from "./ClientCertificates";
import { IssueCertificate } from "../IssueCertificate";
import { StatusActions } from "../StatusActions";

const route = getRouteApi("/data-recipients/$legalEntityId");

type Refresh = () => Promise<unknown>;

function LegalEntity({ detail, refresh }: { detail: DataRecipientDetail; refresh: Refresh }) {
  const of = detail.legalEntityName;

  return (
    <Section
      card
      title="Legal entity"
      links={
        <Link
          icon={Pencil}
          render={(props) => (
            <RouterLink
              {...props}
              to="/data-recipients/$legalEntityId/change"
              params={{ legalEntityId: detail.legalEntityId }}
            />
          )}
        >
          Change details<span className="sr-only"> of {of}</span>
        </Link>
      }
    >
      <DescriptionList>
        <DescriptionItem term="Legal entity ID">
          <Identifier value={detail.legalEntityId} label={`legal entity ID of ${of}`} />
        </DescriptionItem>
        <DescriptionItem term="Name">{detail.legalEntityName}</DescriptionItem>
        <DescriptionItem term="Accreditation number">{detail.accreditationNumber}</DescriptionItem>
        <DescriptionItem term="Accreditation level">{detail.accreditationLevel}</DescriptionItem>
        <DescriptionItem term="Logo URI">
          <Identifier value={detail.logoUri} label={`logo URI of ${of}`} />
        </DescriptionItem>
        <DescriptionItem term="Status">
          <StatusBadge tone={dataRecipientTones[detail.status]}>{detail.status}</StatusBadge>
        </DescriptionItem>
        <DescriptionItem term="Last changed">
          <LocalTime value={detail.lastUpdated} />
        </DescriptionItem>
      </DescriptionList>
      <ActionRow>
        <StatusActions
          record={{ name: of, status: detail.status }}
          actions={detail.actions}
          run={(a) => changeDataRecipientStatus(detail.legalEntityId, a.action, detail.version)}
          refresh={refresh}
        />
      </ActionRow>
    </Section>
  );
}

type ProductProps = {
  product: SoftwareProduct;
  brandName: string;
  detail: DataRecipientDetail;
  refresh: Refresh;
  opened: string | undefined;
};

function Product({ product, brandName, detail, refresh, opened }: ProductProps) {
  const of = product.softwareProductName;
  // An optional value that is not set gives undefined, which the description list shows as "Not set".
  const identifier = (value: string | undefined, name: string) =>
    value && <Identifier value={value} label={`${name} of ${of}`} />;

  return (
    <Section
      card
      collapsible
      defaultOpen={opened === product.softwareProductId}
      title={product.softwareProductName}
      status={
        <StatusBadge tone={brandAndProductTones[product.status]}>{product.status}</StatusBadge>
      }
      links={
        <Link
          icon={Pencil}
          render={(props) => (
            <RouterLink
              {...props}
              to="/data-recipients/$legalEntityId/software-products/$softwareProductId/change"
              params={{
                legalEntityId: detail.legalEntityId,
                softwareProductId: product.softwareProductId,
              }}
            />
          )}
        >
          Change software product<span className="sr-only"> {of}</span>
        </Link>
      }
    >
      <DescriptionList>
        <DescriptionItem term="Software product ID">
          {identifier(product.softwareProductId, "software product ID")}
        </DescriptionItem>
        <DescriptionItem term="Description">{product.softwareProductDescription}</DescriptionItem>
        <DescriptionItem term="Logo URI">{identifier(product.logoUri, "logo URI")}</DescriptionItem>
        <DescriptionItem term="Client URI">
          {identifier(product.clientUri, "client URI")}
        </DescriptionItem>
        <DescriptionItem term="Sector identifier URI">
          {identifier(product.sectorIdentifierUri, "sector identifier URI")}
        </DescriptionItem>
        <DescriptionItem
          term="Redirect URIs"
          values={product.redirectUris.map((value) => identifier(value, "redirect URI"))}
        />
        <DescriptionItem term="Terms of service URI">
          {identifier(product.tosUri, "terms of service URI")}
        </DescriptionItem>
        <DescriptionItem term="Policy URI">
          {identifier(product.policyUri, "policy URI")}
        </DescriptionItem>
        <DescriptionItem term="JWKS URI">{identifier(product.jwksUri, "JWKS URI")}</DescriptionItem>
        <DescriptionItem term="Revocation URI">
          {identifier(product.revocationUri, "revocation URI")}
        </DescriptionItem>
        <DescriptionItem term="Recipient base URI">
          {identifier(product.recipientBaseUri, "recipient base URI")}
        </DescriptionItem>
        <DescriptionItem term="Scopes" values={product.scopes} />
        <DescriptionItem term="Client certificate thumbprints">
          <CertificateList
            productName={of}
            thumbprints={product.clientCertificateThumbprints}
            removable={product.clientCertificateActions.includes("remove")}
            remove={(thumbprint) =>
              removeClientCertificate(
                detail.legalEntityId,
                product.softwareProductId,
                thumbprint,
                detail.version,
              )
            }
            refresh={refresh}
          />
        </DescriptionItem>
      </DescriptionList>
      <ActionRow>
        <StatusActions
          record={{ name: of, status: product.status }}
          actions={product.actions}
          run={(a) =>
            changeSoftwareProductStatus(
              detail.legalEntityId,
              product.softwareProductId,
              a.action,
              detail.version,
            )
          }
          refresh={refresh}
        />
        <RegisterCertificate
          productName={of}
          permitted={product.clientCertificateActions.includes("register")}
          run={(certificate) =>
            registerClientCertificate(
              detail.legalEntityId,
              product.softwareProductId,
              detail.version,
              certificate,
            )
          }
          refresh={refresh}
        />
        <IssueCertificate
          kind="client"
          recordName={of}
          subject={`The CN must be the software product name "${of}", and the O must be the brand name "${brandName}".`}
          fileName={`client-certificate-${product.softwareProductId}.pem`}
          permitted={product.clientCertificateActions.includes("issue")}
          run={(csr) =>
            issueClientCertificate(
              detail.legalEntityId,
              product.softwareProductId,
              detail.version,
              csr,
            )
          }
          refresh={refresh}
        />
      </ActionRow>
    </Section>
  );
}

type BrandProps = {
  brand: DataRecipientBrand;
  detail: DataRecipientDetail;
  refresh: Refresh;
  opened: string | undefined;
};

function Brand({ brand, detail, refresh, opened }: BrandProps) {
  const of = brand.brandName;
  const open =
    opened === brand.dataRecipientBrandId ||
    brand.softwareProducts.some((product) => product.softwareProductId === opened);

  return (
    <Section
      card
      collapsible
      defaultOpen={open}
      title={brand.brandName}
      status={<StatusBadge tone={brandAndProductTones[brand.status]}>{brand.status}</StatusBadge>}
      links={
        <Link
          icon={Pencil}
          render={(props) => (
            <RouterLink
              {...props}
              to="/data-recipients/$legalEntityId/brands/$brandId/change"
              params={{
                legalEntityId: detail.legalEntityId,
                brandId: brand.dataRecipientBrandId,
              }}
            />
          )}
        >
          Change brand<span className="sr-only"> {of}</span>
        </Link>
      }
    >
      <DescriptionList>
        <DescriptionItem term="Brand ID">
          <Identifier value={brand.dataRecipientBrandId} label={`brand ID of ${of}`} />
        </DescriptionItem>
        <DescriptionItem term="Logo URI">
          <Identifier value={brand.logoUri} label={`logo URI of ${of}`} />
        </DescriptionItem>
      </DescriptionList>
      <ActionRow>
        <StatusActions
          record={{ name: of, status: brand.status }}
          actions={brand.actions}
          run={(a) =>
            changeDataRecipientBrandStatus(
              detail.legalEntityId,
              brand.dataRecipientBrandId,
              a.action,
              detail.version,
            )
          }
          refresh={refresh}
        />
      </ActionRow>
      <Section
        title="Software products"
        showAll={`software products of ${of}`}
        links={
          <Link
            icon={Plus}
            render={(props) => (
              <RouterLink
                {...props}
                to="/data-recipients/$legalEntityId/brands/$brandId/add-software-product"
                params={{
                  legalEntityId: detail.legalEntityId,
                  brandId: brand.dataRecipientBrandId,
                }}
              />
            )}
          >
            Add software product<span className="sr-only"> to {of}</span>
          </Link>
        }
      >
        {brand.softwareProducts.length === 0 && (
          <EmptyState title="This brand has no software products" />
        )}
        {brand.softwareProducts.map((product) => (
          <Product
            key={product.softwareProductId}
            product={product}
            brandName={brand.brandName}
            detail={detail}
            refresh={refresh}
            opened={opened}
          />
        ))}
      </Section>
    </Section>
  );
}

export function DataRecipientPage() {
  const { legalEntityId } = route.useParams();
  const focusTitle = useLocation({ select: (location) => location.state.focusTitle ?? false });
  const opened = useLocation({ select: (location) => location.state.open });
  const detail = useQuery({
    queryKey: ["data-recipient", legalEntityId],
    queryFn: ({ signal }) => getDataRecipient(legalEntityId, signal),
  });
  const queryClient = useQueryClient();
  // The detail is on the page, so its refetch completes before the promise resolves.
  const refresh = () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: ["data-recipient", legalEntityId] }),
      queryClient.invalidateQueries({ queryKey: ["data-recipient-history", legalEntityId] }),
      queryClient.invalidateQueries({ queryKey: ["data-recipients"] }),
    ]);

  return (
    <>
      <PageHeader
        title={detail.data?.legalEntityName ?? "Data recipient"}
        focusTitle={focusTitle}
        description="A data recipient legal entity of the Register, with its brands and software products."
        actions={
          <Link
            icon={History}
            render={(props) => (
              <RouterLink
                {...props}
                to="/data-recipients/$legalEntityId/history"
                params={{ legalEntityId }}
              />
            )}
          >
            History
          </Link>
        }
      />
      {detail.isPending && <LoadingState />}
      {detail.isError && <ErrorState problem={problemOf(detail.error)} />}
      {detail.data && (
        <>
          <LegalEntity detail={detail.data} refresh={refresh} />
          <Section
            title="Brands"
            showAll="brands"
            links={
              <Link
                icon={Plus}
                render={(props) => (
                  <RouterLink
                    {...props}
                    to="/data-recipients/$legalEntityId/add-brand"
                    params={{ legalEntityId }}
                  />
                )}
              >
                Add brand
              </Link>
            }
          >
            {detail.data.dataRecipientBrands.length === 0 && (
              <EmptyState title="This data recipient has no brands" />
            )}
            {detail.data.dataRecipientBrands.map((brand) => (
              <Brand
                key={brand.dataRecipientBrandId}
                brand={brand}
                detail={detail.data}
                refresh={refresh}
                opened={opened}
              />
            ))}
          </Section>
        </>
      )}
    </>
  );
}
