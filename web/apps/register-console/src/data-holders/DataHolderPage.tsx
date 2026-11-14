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
  changeDataHolderBrandStatus,
  changeDataHolderStatus,
  getDataHolder,
  issueServerCertificate,
  problemOf,
  type DataHolderBrand,
  type DataHolderDetail,
} from "../api/client";
import { IssueCertificate } from "../IssueCertificate";
import { StatusActions } from "../StatusActions";
import { brandAndProductTones, dataHolderTones } from "../statuses";
import { brandIdOf } from "./brand-id";

const route = getRouteApi("/data-holders/$legalEntityId");

type Refresh = () => Promise<unknown>;

function LegalEntity({ detail, refresh }: { detail: DataHolderDetail; refresh: Refresh }) {
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
              to="/data-holders/$legalEntityId/change"
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
        <DescriptionItem term="Logo URI">
          <Identifier value={detail.logoUri} label={`logo URI of ${of}`} />
        </DescriptionItem>
        <DescriptionItem term="Status">
          <StatusBadge tone={dataHolderTones[detail.status]}>{detail.status}</StatusBadge>
        </DescriptionItem>
        <DescriptionItem term="Last changed">
          <LocalTime value={detail.lastUpdated} />
        </DescriptionItem>
      </DescriptionList>
      <ActionRow>
        <StatusActions
          record={{ name: of, status: detail.status }}
          actions={detail.actions}
          run={(a) => changeDataHolderStatus(detail.legalEntityId, a.action, detail.version)}
          refresh={refresh}
        />
      </ActionRow>
    </Section>
  );
}

type BrandProps = {
  brand: DataHolderBrand;
  detail: DataHolderDetail;
  refresh: Refresh;
  opened: string | undefined;
};

function Brand({ brand, detail, refresh, opened }: BrandProps) {
  const of = brand.brandName;
  const brandId = brandIdOf(brand);
  const endpoints = brand.endpointDetail;
  // An optional value that is not set gives undefined, which the description list shows as "Not set".
  const identifier = (value: string | undefined, name: string) =>
    value && <Identifier value={value} label={`${name} of ${of}`} />;

  return (
    <Section
      card
      collapsible
      defaultOpen={opened === brandId}
      title={brand.brandName}
      status={<StatusBadge tone={brandAndProductTones[brand.status]}>{brand.status}</StatusBadge>}
      links={
        <Link
          icon={Pencil}
          render={(props) => (
            <RouterLink
              {...props}
              to="/data-holders/$legalEntityId/brands/$brandId/change"
              params={{ legalEntityId: detail.legalEntityId, brandId }}
            />
          )}
        >
          Change brand<span className="sr-only"> {of}</span>
        </Link>
      }
    >
      <DescriptionList>
        <DescriptionItem term="Brand ID">
          {identifier(brand.dataHolderBrandId, "brand ID")}
        </DescriptionItem>
        <DescriptionItem term="Interim ID">
          {identifier(brand.interimId, "interim ID")}
        </DescriptionItem>
        <DescriptionItem term="Brand group">{brand.brandGroup}</DescriptionItem>
        <DescriptionItem term="Industries">{brand.industries.join(", ")}</DescriptionItem>
        <DescriptionItem term="Logo URI">{identifier(brand.logoUri, "logo URI")}</DescriptionItem>
        <DescriptionItem term="Endpoint version">{endpoints.version}</DescriptionItem>
        <DescriptionItem term="Public base URI">
          {identifier(endpoints.publicBaseUri, "public base URI")}
        </DescriptionItem>
        <DescriptionItem term="Product base URI">
          {identifier(endpoints.productBaseUri, "product base URI")}
        </DescriptionItem>
        <DescriptionItem term="Resource base URI">
          {identifier(endpoints.resourceBaseUri, "resource base URI")}
        </DescriptionItem>
        <DescriptionItem term="Infosec base URI">
          {identifier(endpoints.infosecBaseUri, "infosec base URI")}
        </DescriptionItem>
        <DescriptionItem term="Extension base URI">
          {identifier(endpoints.extensionBaseUri, "extension base URI")}
        </DescriptionItem>
        <DescriptionItem term="Website URI">
          {identifier(endpoints.websiteUri, "website URI")}
        </DescriptionItem>
        <DescriptionItem term="JWKS endpoint">
          {identifier(brand.jwksEndpoint, "JWKS endpoint")}
        </DescriptionItem>
      </DescriptionList>
      <ActionRow>
        <StatusActions
          record={{ name: of, status: brand.status }}
          actions={brand.actions}
          run={(a) =>
            changeDataHolderBrandStatus(detail.legalEntityId, brandId, a.action, detail.version)
          }
          refresh={refresh}
        />
        <IssueCertificate
          kind="server"
          recordName={of}
          subject={`The CN must be the primary DNS name, and the O must be the brand name "${of}" or the legal entity name "${detail.legalEntityName}".`}
          fileName={`server-certificate-${brandId}.pem`}
          permitted={brand.serverCertificateActions.includes("issue")}
          run={(csr) => issueServerCertificate(detail.legalEntityId, brandId, detail.version, csr)}
          refresh={refresh}
        />
      </ActionRow>
    </Section>
  );
}

export function DataHolderPage() {
  const { legalEntityId } = route.useParams();
  const focusTitle = useLocation({ select: (location) => location.state.focusTitle ?? false });
  const opened = useLocation({ select: (location) => location.state.open });
  const detail = useQuery({
    queryKey: ["data-holder", legalEntityId],
    queryFn: ({ signal }) => getDataHolder(legalEntityId, signal),
  });
  const queryClient = useQueryClient();
  // The detail is on the page, so its refetch completes before the promise resolves.
  const refresh = () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: ["data-holder", legalEntityId] }),
      queryClient.invalidateQueries({ queryKey: ["data-holder-history", legalEntityId] }),
      queryClient.invalidateQueries({ queryKey: ["data-holders"] }),
    ]);

  return (
    <>
      <PageHeader
        title={detail.data?.legalEntityName ?? "Data holder"}
        focusTitle={focusTitle}
        description="A data holder legal entity of the Register, with its brands."
        actions={
          <Link
            icon={History}
            render={(props) => (
              <RouterLink
                {...props}
                to="/data-holders/$legalEntityId/history"
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
                    to="/data-holders/$legalEntityId/add-brand"
                    params={{ legalEntityId }}
                  />
                )}
              >
                Add brand
              </Link>
            }
          >
            {detail.data.dataHolderBrands.length === 0 && (
              <EmptyState title="This data holder has no brands" />
            )}
            {detail.data.dataHolderBrands.map((brand) => (
              <Brand
                key={brandIdOf(brand)}
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
