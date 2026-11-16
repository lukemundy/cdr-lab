# Design system: Non-functional requirements

This area gives the non-functional requirements of all the design system.

## Non-functional requirements

- <a id="ui-0180"></a>**UI-0180** Supply chain: the design system has 0 dependencies that run an install script. ([Judgement: the npm supply-chain controls (ADR 0015)][adr-0015])
- <a id="ui-0181"></a>**UI-0181** Single source: a UI keeps no copy of a token or a component. A change to a token or a component shows in each UI at its next build. ([Judgement: one design system (ADR 0016)][adr-0016])

[adr-0015]: ../../../../docs/adr/0015-web-toolchain-and-npm-supply-chain-controls.md
[adr-0016]: ../../../../docs/adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md
