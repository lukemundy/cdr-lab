import type { ReactNode } from "react";

export function DescriptionList({ children }: { children: ReactNode }) {
  // A container, so that each value goes below its term when the list is narrow, for example in a consumer card.
  return <dl className="@container flex flex-col text-sm">{children}</dl>;
}

export type DescriptionItemProps = {
  term: string;
  children?: ReactNode;
  values?: readonly ReactNode[];
  // The text of an empty list of values, when no values is a normal state and not a value that is not set.
  empty?: string;
};

function isEmpty(value: ReactNode): boolean {
  return value === undefined || value === null || value === false || value === "";
}

function Value({ children, values, empty }: DescriptionItemProps) {
  if (values) {
    if (values.length === 0) return <Muted>{empty ?? "Not set"}</Muted>;
    return (
      <ul className="flex flex-col gap-1">
        {values.map((value, index) => (
          // oxlint-disable-next-line react/no-array-index-key -- the values have no identity, and the list does not change order.
          <li key={index}>{value}</li>
        ))}
      </ul>
    );
  }
  return isEmpty(children) ? <Muted>Not set</Muted> : children;
}

function Muted({ children }: { children: string }) {
  return <span className="text-fg-muted">{children}</span>;
}

export function DescriptionItem(props: DescriptionItemProps) {
  return (
    // HTML permits a div around each term and its value, which gives one row of the grid.
    <div className="grid grid-cols-1 gap-1 border-b @md:grid-cols-[12rem_minmax(0,1fr)] @md:gap-4 border-border py-2 last:border-b-0">
      <dt className="font-medium text-fg-muted">{props.term}</dt>
      <dd className="min-w-0 text-fg wrap-break-word">
        <Value {...props} />
      </dd>
    </div>
  );
}
