import { Link } from "@cdr-lab/ui/link";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { Link as RouterLink } from "@tanstack/react-router";

export function NotFoundPage() {
  return (
    <>
      <PageHeader title="Page not found" description="The console has no page at this address." />
      <Link render={(props) => <RouterLink {...props} to="/data-recipients" />}>
        Go to the data recipients
      </Link>
    </>
  );
}
