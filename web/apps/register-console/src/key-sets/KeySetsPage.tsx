import { ActionRow } from "@cdr-lab/ui/action-row";
import { DescriptionItem, DescriptionList } from "@cdr-lab/ui/description-list";
import { Identifier } from "@cdr-lab/ui/identifier";
import { PageHeader } from "@cdr-lab/ui/page-header";
import { Section } from "@cdr-lab/ui/section";
import { EmptyState, ErrorState, LoadingState } from "@cdr-lab/ui/states";
import { LocalTime } from "@cdr-lab/ui/local-time";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { listKeySets, problemOf, type KeySet, type KeyState } from "../api/client";
import { KeyActions, type KeySetNames } from "./KeyActions";

const keySets = {
  "access-token": {
    title: "Access token keys",
    description:
      "Sign the access tokens of the Register. The jwks_uri of the OpenID Provider configuration publishes them.",
    noun: "access token",
    signs: "the access tokens",
  },
  ssa: {
    title: "SSA keys",
    description: "Sign the software statement assertions. Get JWKS publishes them.",
    noun: "SSA",
    signs: "the SSAs",
  },
} as const satisfies Record<KeySet["keySet"], KeySetNames & { description: string }>;

const states = {
  signing: { title: "Key that signs", keyOf: (noun: string) => `the ${noun} key that signs` },
  next: { title: "Next key", keyOf: (noun: string) => `the next ${noun} key` },
  previous: { title: "Previous key", keyOf: (noun: string) => `the previous ${noun} key` },
} as const satisfies Record<KeyState, { title: string; keyOf: (noun: string) => string }>;

function KeySetSection({ set, refresh }: { set: KeySet; refresh: () => Promise<unknown> }) {
  const names = keySets[set.keySet];
  const { title, description, noun } = names;
  const { keys } = set;

  return (
    <Section card title={title} description={description}>
      <ActionRow>
        <KeyActions keySet={set} names={names} refresh={refresh} />
      </ActionRow>
      {keys.length === 0 && <EmptyState title="This key set has no keys" />}
      {keys.map(({ keyId, state, createdAt }) => (
        <Section card key={keyId} title={states[state].title}>
          <DescriptionList>
            <DescriptionItem term="Key ID">
              <Identifier value={keyId} label={`key ID of ${states[state].keyOf(noun)}`} />
            </DescriptionItem>
            <DescriptionItem term="Created">
              <LocalTime value={createdAt} seconds />
            </DescriptionItem>
          </DescriptionList>
        </Section>
      ))}
    </Section>
  );
}

export function KeySetsPage() {
  const list = useQuery({ queryKey: ["key-sets"], queryFn: ({ signal }) => listKeySets(signal) });
  const queryClient = useQueryClient();
  // The key sets are on the page, so their refetch completes before the promise resolves.
  const refresh = () => queryClient.invalidateQueries({ queryKey: ["key-sets"] });

  return (
    <>
      <PageHeader
        title="Signing keys"
        description="The key sets of the Register, and the state of each key. The JWKS of a key set publishes each key in it."
      />
      {list.isPending && <LoadingState />}
      {list.isError && <ErrorState problem={problemOf(list.error)} />}
      {list.data?.data.map((set) => (
        <KeySetSection key={set.keySet} set={set} refresh={refresh} />
      ))}
    </>
  );
}
