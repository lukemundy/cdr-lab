import { ChevronRight, ChevronsDownUp, ChevronsUpDown } from "lucide-react";
import {
  createContext,
  useContext,
  useEffect,
  useId,
  useLayoutEffect,
  useRef,
  useState,
  type ReactNode,
} from "react";

import { cn } from "../lib/cn";
import { focusRing } from "../lib/focus";
import { Button } from "./button";

// The page header is the h1, so a section that is not in another section has an h2.
const HeadingLevel = createContext(2);
const InCard = createContext(false);
type ShowAllClick = { open: boolean; count: number };

// The last Show all or Hide all around the card. The count orders the clicks of all sections.
const ShowAll = createContext<ShowAllClick | undefined>(undefined);
let clicks = 0;

function latest(a: ShowAllClick | undefined, b: ShowAllClick | undefined) {
  if (!a || !b) return a ?? b;
  return a.count > b.count ? a : b;
}

const headings = ["h2", "h3", "h4", "h5", "h6"] as const;

const headingSizes: Record<number, string> = { 2: "text-xl", 3: "text-lg" };

export type SectionProps = {
  title: string;
  description?: string;
  links?: ReactNode;
  card?: boolean;
  collapsible?: boolean;
  defaultOpen?: boolean;
  status?: ReactNode;
  // What Show all opens, for the accessible names of Show all and Hide all, for example "brands".
  showAll?: string;
  children: ReactNode;
};

function useOpen(collapsible: boolean, defaultOpen: boolean) {
  const showAll = useContext(ShowAll);
  const [open, setOpen] = useState(!collapsible || defaultOpen);
  const [seen, setSeen] = useState(showAll?.count);
  if (collapsible && showAll && showAll.count !== seen) {
    setSeen(showAll.count);
    setOpen(showAll.open);
  }
  return [open, setOpen] as const;
}

// React has no value "until-found" for hidden, so the body gets the attribute and the event from the DOM.
function useHiddenUntilFound(open: boolean, setOpen: (open: boolean) => void) {
  const body = useRef<HTMLDivElement>(null);
  useLayoutEffect(() => {
    if (open) body.current?.removeAttribute("hidden");
    else body.current?.setAttribute("hidden", "until-found");
  }, [open]);
  useEffect(() => {
    const element = body.current;
    const found = () => setOpen(true);
    element?.addEventListener("beforematch", found);
    return () => element?.removeEventListener("beforematch", found);
  }, [setOpen]);
  return body;
}

export function Section({
  title,
  description,
  links,
  card = false,
  collapsible = false,
  defaultOpen = false,
  status,
  showAll,
  children,
}: SectionProps) {
  const level = useContext(HeadingLevel);
  const inCard = useContext(InCard);
  const parentShowAll = useContext(ShowAll);
  const [shown, setShown] = useState<ShowAllClick>();
  const showAllClick = latest(parentShowAll, shown);
  const [open, setOpen] = useOpen(collapsible, defaultOpen);
  const body = useHiddenUntilFound(open, setOpen);
  const id = useId();
  const bodyId = useId();
  // HTML has no heading below h6.
  const Heading = headings[Math.min(level, 6) - 2] ?? "h6";

  const header = (
    <div
      className={cn(
        "flex flex-col gap-1",
        card && "px-4 py-3",
        card && open && "border-b border-border",
      )}
    >
      <div className="flex flex-wrap items-baseline gap-x-4 gap-y-1">
        <Heading
          id={id}
          className={cn("font-semibold text-fg", headingSizes[level] ?? "text-base")}
        >
          {collapsible ? (
            <button
              type="button"
              aria-expanded={open}
              aria-controls={bodyId}
              onClick={() => setOpen(!open)}
              className={cn(
                "-mx-1 inline-flex cursor-pointer items-center gap-1 rounded-sm px-1 hover:bg-secondary",
                focusRing,
              )}
            >
              <ChevronRight
                aria-hidden="true"
                className={cn(
                  "size-4 shrink-0 motion-safe:transition-transform",
                  open && "rotate-90",
                )}
              />
              {title}
            </button>
          ) : (
            title
          )}
        </Heading>
        {status}
        {links && (
          <div className="flex flex-wrap items-center gap-x-4 gap-y-1 text-sm">{links}</div>
        )}
        {showAll && (
          <div className="ml-auto flex gap-2 self-center">
            {[
              { open: true, label: "Show all", icon: ChevronsUpDown },
              { open: false, label: "Hide all", icon: ChevronsDownUp },
            ].map((option) => (
              <Button
                key={option.label}
                variant="quiet"
                icon={option.icon}
                onClick={() => setShown({ open: option.open, count: ++clicks })}
              >
                {option.label}
                <span className="sr-only"> {showAll}</span>
              </Button>
            ))}
          </div>
        )}
      </div>
      {description && <p className="text-sm text-fg-muted">{description}</p>}
    </div>
  );
  const content = (
    <HeadingLevel value={level + 1}>
      <InCard value={inCard || card}>
        <ShowAll value={showAllClick}>{children}</ShowAll>
      </InCard>
    </HeadingLevel>
  );

  if (!card) {
    return (
      <section
        aria-labelledby={id}
        className={cn("flex flex-col gap-4", level === 2 ? "py-6" : "py-2")}
      >
        {header}
        {content}
      </section>
    );
  }

  return (
    <section
      aria-labelledby={id}
      className={cn(
        "rounded-lg border border-border",
        inCard ? "bg-page" : "bg-surface",
        level === 2 && "my-6",
      )}
    >
      {header}
      {/* A body with hidden="until-found" keeps its box, so the padding is on the element in it. */}
      <div ref={body} id={bodyId}>
        <div className="flex flex-col gap-4 p-4">{content}</div>
      </div>
    </section>
  );
}
