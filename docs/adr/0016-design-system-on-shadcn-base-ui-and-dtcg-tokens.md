# 0016. Build one design system from shadcn/ui on Base UI, Tailwind CSS and DTCG design tokens

- Date: 2026-10-22

## Context

The web UIs are React single-page applications ([ADR 0014](0014-web-uis-as-embedded-react-single-page-applications.md)). The authentication and consent pages of the data holder are Go pages with no script, outside that decision. Operator consoles are for desktop and are data-dense. Consumer UIs are for mobile. All UIs must use one design system. Each mock data holder and each data recipient is a different brand, so one set of components must take more than one brand. The project has no designer. In order of priority, the UIs must be maintainable, follow the CDR consumer experience (CX) standards closely, look like real products, and be accessible.

The research is in [Web front ends](../research/2026-10-22-web-front-ends.md), sections 7 to 10. We compared these options:

| | shadcn/ui on Base UI | shadcn/ui on React Aria | Mantine 9 | React Aria Components and our own CSS |
|---|---|---|---|---|
| Model | Component source copied into the repository | Same | A library with more than 100 styled components | Unstyled components, styled by us |
| Base | Base UI 1.x, stable since 2025-12-11, from the authors of Radix and Floating UI ([Base UI](https://base-ui.com/)). The shadcn default since July 2026 ([shadcn changelog](https://ui.shadcn.com/docs/changelog/2026-07-base-ui-default)). | React Aria (Adobe), tested with screen readers on devices, 30 or more languages ([React Aria](https://react-aria.adobe.com/)). A shadcn base since July 2026 ([shadcn changelog](https://ui.shadcn.com/docs/changelog/2026-07-react-aria)). | Own | React Aria |
| Several brands | CSS custom properties that we own | Same | A theme object | Our own |
| Effort for a good look with no designer | Low: start from a preset | Low, with fewer examples | Lowest for consoles | High |

A headless library gives keyboard support, focus management and ARIA attributes, so much accessibility comes with the components.

Tailwind Labs joined Shopify on 2026-09-09. Tailwind CSS stays MIT-licensed ([Tailwind blog](https://tailwindcss.com/blog/tailwind-is-joining-shopify)). Tailwind CSS v4 gives its theme values as CSS custom properties, and its output is static CSS.

The W3C Design Tokens Community Group published its first stable format, 2025.10. Its Resolver module gives sets of tokens and modifiers, for example a theme, in a resolution order ([DTCG Resolver](https://www.designtokens.org/TR/2025.10/resolver/)). Terrazzo 2 supports the full format with resolvers ([Terrazzo changelog](https://github.com/terrazzoapp/terrazzo/blob/main/packages/cli/CHANGELOG.md)). Style Dictionary v5 does not support the full format yet ([Style Dictionary: DTCG](https://styledictionary.com/info/dtcg/)).

Accessibility requirements:

- The CDS says that participants "MUST seek to comply" with WCAG guidelines 1.4, 2.1, 2.5, 3.1 and 3.3 "throughout the Consent Model" ([CDS: Accessibility Standards](https://consumerdatastandardsaustralia.github.io/standards/#accessibility-standards)). "WCAG 2.1" there is guideline 2.1 (Keyboard Accessible), not the WCAG version. The text gives no version and no conformance level.
- The Australian Human Rights Commission says: "Organisations should conform with WCAG 2.2 at a minimum Level AA" ([AHRC guidelines](https://humanrights.gov.au/our-work/disability-rights/publications/guidelines-equal-access-digital-goods-and-services)).
- The DSB says that its own CX Figma files are only "partially conformant" to WCAG 2.1 AA ([DSB: open source assets](https://cx.dsb.gov.au/overview/open-source-assets)).
- Automated checks find only part of the issues: Deque says that axe-core "automatically catches up to 57% of WCAG issues" (as cited in [Storybook: accessibility testing](https://storybook.js.org/docs/writing-tests/accessibility-testing)).

## Decision

This decision applies to all web UIs, and to the authentication and consent pages of the data holder.

- The components are shadcn/ui components on Base UI, in one shared workspace package. Apps import the components from the package and do not keep their own copies. The reason for Base UI: it is the shadcn default, with the most examples, and its accessibility is sufficient for the bar below. React Aria was the second choice.
- Styles use Tailwind CSS v4. Components use only semantic tokens, not raw colours or sizes.
- The Go pages of the data holder use the tokens, the fonts and a plain stylesheet of the design system, which uses only the semantic tokens. They do not use Tailwind CSS or the components. The reason: the pages have no React, and Tailwind CSS must scan the source that uses its classes, so the web build then depends on the templates of a service.
- The design tokens are DTCG 2025.10 files in a shared workspace package. They have a primitive tier (raw values), a semantic tier (roles, for example surface, foreground, primary action, danger) and a component tier only where a component needs its own setting. A `brand` modifier gives each brand its values. Terrazzo builds the tokens into CSS custom properties. Each mock participant has its own brand, and the consoles have a neutral lab brand.
- The UIs have a light theme only. The semantic tier lets a later change add a dark theme as a second modifier.
- Forms use React Hook Form with Zod schemas. Tables use TanStack Table. Icons are Lucide. Fonts are self-hosted, so that no third-party host is in the CSP. Dates, times and amounts are formatted with `Intl` for `en-AU`, at display only ([ADR 0012](0012-utc-for-all-times.md)).
- The accessibility bar has two levels. A consumer UI targets WCAG 2.2 Level AA. An operator console has no target beyond what the components give. The tests of all UIs include automated axe-core checks. There are no manual screen-reader audits. The reason for WCAG 2.2 AA: the CDS text cannot be tested as written, and the AHRC recommends this level.
- The code is the source of truth for the design. No design-tool file is a source. The DSB CX Figma files are a reference for consumer flows and content only, not for visual design.
- There is no component catalogue (for example Storybook) now. The component tests show the states of each component.

## Consequences

- We own the component source. An upstream fix to a shadcn component needs a merge with `shadcn --diff`. A local change to a component is easy.
- The components look like many other shadcn sites until the tokens of each brand are set.
- Each brand needs a contrast check of its colour roles.
- A dark theme later needs a second set of semantic values and a contrast check for each brand.
- Without manual audits, a screen-reader defect that axe-core does not find can stay in a consumer UI.
- Terrazzo adds a build step for the tokens. It is newer and has fewer users than Style Dictionary.
- Tailwind utility classes make the markup long, but each style is next to the element that it changes.
- The Go pages have a second stylesheet, with class names that their templates use. A change to the look of a control can need a change to a component and to that stylesheet.
- Without a catalogue, a person who wants to see a component runs an app or its tests.
