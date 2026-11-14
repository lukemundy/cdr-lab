# Preset bank brands: design

- Research date: 2026-11-01
- Scope: the visual design of the consumer authentication and consent pages of the data holder for four preset brands. The names, palettes and logos of the brands. How real Australian banks design their sign-in pages. What the CDR consumer experience (CX) standards and guidelines say about these pages. A proposed design direction for the design tokens of [ADR 0016](../adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md).
- Applies to: CDS 1.36.0, DSB CX Guidelines and CX Checklist v1.36.0, WCAG 2.2, the APRA ADI register of 15 October 2026, and the public web pages of the four banks on 2026-11-01.

This note does not repeat [Web front ends](2026-10-22-web-front-ends.md), sections 7 to 10 (design system, tokens, CX requirements, accessibility). It adds to them. The note [Branded data holder instances](2026-11-01-branded-data-holder-instances.md) gives the technical facts. The gallery of this research is a separate HTML file, which is not in the repository.

The intent is that each preset bank looks similar to one of the four major Australian banks, so that the lab pages look like the pages of a real bank. Thus the palettes use the colours of the real banks, and the fonts have a similar character. The names and the logos are original. When this note was written, the names, palettes and logos were candidates for review in the gallery.

## Legend

- **UNVERIFIED**: I did not find a primary source. Do a check before you depend on it.
- **JUDGEMENT**: my opinion, not a fact from a source.
- **Measured**: a result from a command in this research session.

## Sources

| ID | Source |
|---|---|
| CDS-A | CDS 1.36.0 source, CX authentication standards: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/slate/source/includes/cx_standards/authentication.md |
| CDS-Z | CDS 1.36.0 source, CX authorisation standards: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/slate/source/includes/cx_standards/authorisation.md |
| CDS-X | CDS 1.36.0 source, accessibility standards: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/slate/source/includes/cx_standards/accessibility.md |
| CDS-L | CDS 1.36.0 source, data language standards: https://github.com/ConsumerDataStandardsAustralia/standards/blob/v1.36.0/slate/source/includes/cx_standards/data_language.md |
| DSB-2AU | DSB CX Guidelines, Redirect with One Time Password: https://cx.dsb.gov.au/authenticate/redirect-with-one-time-password/cx-guidelines |
| DSB-3AU | DSB CX Guidelines, Authorisation to disclose: https://cx.dsb.gov.au/authorise/authorisation-to-disclose/cx-guidelines |
| APRA | APRA, Register of authorised deposit-taking institutions (updated 15 October 2026): https://www.apra.gov.au/register-of-authorised-deposit-taking-institutions |
| CBA | Measured: https://www.commbank.com.au/ , its stylesheet `/etc.clientlibs/cba/commbank/clientlibs/clientlib-main.min.css`, and the NetBank log-on page https://www.my.commbank.com.au/netbank/Logon/Logon.aspx |
| NAB | Measured: https://www.nab.com.au/ and its stylesheet `/etc.clientlibs/nab/clientlibs/clientlib-generated-components.lc-f3b53c12e410fec57b8bfdce5b055a02-lc.css` ; https://ib.nab.com.au/login |
| ANZ | Measured: https://www.anz.com.au/ and its stylesheet `/apps/settings/wcm/designs/anzcomau/clientlibs/latest-SNAPSHOT/css/homepage.css` ; https://login.anz.com/internetbanking |
| WBC | Measured: https://www.westpac.com.au/ and its stylesheet `/styles/dist/style-config.compiled.min.css` |
| GEL | Westpac GEL design system, colour article (a preview deployment, so the URL can change): https://gel-next-site-2b7nsf75a-westpacgel.vercel.app/articles/colour |
| WCAG | WCAG 2.2: https://www.w3.org/TR/WCAG22/ , Understanding 1.4.11: https://www.w3.org/WAI/WCAG22/Understanding/non-text-contrast.html |
| BSB | AusPayNet BSB directory search: https://bsb.auspaynet.com.au/ |
| RFC6761 | RFC 6761, Special-Use Domain Names (`.test`): https://www.rfc-editor.org/rfc/rfc6761 |
| GF | Google Fonts repository (licences of the fonts): https://github.com/google/fonts/tree/main/ofl |

Method for the bank colours: I downloaded each home page and its main stylesheet with `curl`, and counted each hex colour with `grep -o | sort | uniq -c`. A colour that a stylesheet uses many times is a brand colour. For Westpac, the stylesheet has named custom properties (`--wbc-primary-500`), which are better evidence than a count. The colours of a logo can be different from the colours of a stylesheet.

## 1. The current pages

- The consumer pages of the data holder are Go templates in `services/dataholder/internal/adapter/consumerpages/templates/`. They have no stylesheet. A browser shows them with its default styles: a serif font and default controls (measured).
- The consoles use the `lab` brand of `web/packages/tokens`: Inter, a neutral grey page and a blue primary colour, with small radii (4 to 8 px) and light shadows. This is the "flat" look of the project.
- The token resolver already has a `brand` modifier with one context, `lab`, and a `density` modifier with one context, `compact`. The semantic roles include `fg-on-primary`, so a brand with a light primary colour (yellow) is possible now.

## 2. Real Australian bank web pages

### 2.1 Colours (measured)

| Bank | Colour | Evidence |
|---|---|---|
| CommBank | `#FFCC00` yellow | `#fc0`, 40 uses in clientlib-main CSS (CBA) |
| | `#231F20` near black | the most used colour in the same file, 116 uses |
| | `#1175B5` blue | 108 uses; links |
| | `#F2B400` to `#FEF10C` | a yellow gradient on the NetBank log-on page |
| NAB | `#C20000` red | 69 uses in the components CSS (NAB) |
| | `#8F0000`, `#750000` | darker reds, 14 and 4 uses |
| | `#3A4F59`, `#26343A` | slate greys, 17 and 3 uses |
| ANZ | `#004165` navy | 77 uses in the home page HTML (ANZ) |
| | `#0063CD` blue | 250 uses in homepage.css; also on the log-on page |
| | `#0056B2`, `#002B59` | darker blues, 178 and 219 uses |
| | `#8ADAF9`, `#E2F6FD` | light blue and a blue tint, 44 and 43 uses |
| Westpac | `#DA1710` red | `--wbc-primary-500` (WBC). GEL also gives `#DA1710` as the Westpac primary colour (GEL). |
| | `#B9140E`, `#99100B` | `--wbc-primary-600` and `-700` |
| | `#1F1C4F` navy | `--wbc-hero-500`, also `--text-heading` |
| | `#FF3DDB` pink | `--wbc-pop-500` |

Third-party colour sites give other values, for example `#ED1A3B` for Westpac and `#009FE1` for ANZ. I did not use them, because the banks' own stylesheets disagree with them.

### 2.2 Fonts, radius and layout (measured)

- All four banks use a proprietary font: CommBank "CBABeaconSans", ANZ "Aeonik", Westpac "Brown Pro" and a display font "Westpac". The NAB home page uses `source-sans-pro`, which is free. Thus the presets cannot use the real fonts, except for NAB.
- Radius: ANZ uses mostly 3 to 4 px, NAB 4 px. Westpac has a radius scale from 3 to 24 px and 999 px (pill). CommBank uses mostly `50%` (round items) and 0.
- The NetBank log-on page (CBA) has a single form: "Client number", "Password", "Remember my client number". It has a security reminder: "Always type commbank.com.au or netbank.com.au directly into your browser to log on to NetBank. Never click on links in emails or text messages to access your online banking." Its footer has "Terms of use | Security | Privacy", the copyright and the ABN.
- The NAB, ANZ and Westpac log-on pages are JavaScript applications. `curl` got only the shell of each page, so I have no measured layout for them (UNVERIFIED).
- I found no public screenshots or text of the CDR authorisation pages of the four banks. Their CDR policies say only that the consumer goes to the bank, selects accounts and confirms. Thus this note has no evidence of how a real bank designs its CDR consent page (UNVERIFIED).

### 2.3 Trust cues on bank sign-in pages (JUDGEMENT, from the CBA page and general knowledge)

- A clear logo at the top left, on a brand-colour band or a white bar.
- One task for each page, in a card or a narrow column, with a large primary button.
- A security message near the form.
- A short footer with the legal entity.

## 3. CDR rules for these pages

### 3.1 Standards (binding)

- Data holders "MUST state in consumer-facing interactions and communications that third parties do not need consumer passwords to access CDR data" (CDS-A).
- Data holders "MUST NOT include forgotten details links in redirect screens" (CDS-A).
- Data holders "MUST communicate the expiry period of the OTP to the consumer in the authentication flow" (CDS-A).
- Data holders "MUST NOT introduce unwarranted friction". Unwarranted friction includes "Offering additional or alternative services" and "Referencing or including other documents" (CDS-A).
- Account selection, account confirmation, and a visual "pending" status with an explanation (CDS-Z).
- Data holders "MUST seek to comply" with WCAG guidelines 1.4, 2.1, 2.5, 3.1 and 3.3 (CDS-X). [ADR 0016](../adr/0016-design-system-on-shadcn-base-ui-and-dtcg-tokens.md) sets WCAG 2.2 AA as the target of a consumer UI.
- The data language text: "Account name, type and balance" with "Name of account; Type of account; Account balance;", and "Transaction details" with "Incoming and outgoing transactions; Amounts; Dates; Descriptions of transactions; Who you have sent money to and received money from; (e.g. their name)" (CDS-L).

### 3.2 Guidelines (not binding)

- Use the brand name of the data recipient, also on the cancel screens (DSB-2AU, ref 16). Show the name of the brand that receives the data (DSB-2AU, ref 37; DSB-3AU, 3AU.00.30).
- Let the consumer ask for a new OTP (DSB-2AU, ref 19).
- Data holders can include their CDR policy in the flow (DSB-2AU, ref 17; DSB-3AU, 3AU.02.20).
- Use "confirm" or "authorise" for the final action (DSB-3AU, 3AU.02.21). Use plain language such as "stop sharing" (3AU.02.18).
- Neither guideline page says anything about the data holder logo, colours, layout or a progress indicator. The design of these items is free, within the standards above.

### 3.3 Effects on the design

- The footer of a consumer page must not link to the terms and conditions, the privacy policy or other bank services (CDS-A, friction). A real bank footer has these links (CBA), so the preset pages differ from a real sign-in page on purpose. The simulation footer and an optional link to the CDR policy are permitted (JUDGEMENT).
- No "Forgotten your customer number?" link (CDS-A). The real NetBank page has one (CBA).
- The page shows the OTP expiry and the password statement as normal content, not in a tooltip.

## 4. Names

### 4.1 Rules

- Realistic, with no "Mock", "Example" or a similar word.
- Not confusingly similar to a real ADI or bank brand. I compared each name with the APRA register (APRA) and with known brands of the four major banks (Bankwest, St.George, BankSA, Bank of Melbourne, RAMS, ubank).
- The legal entity name fits the Register limit of 36 characters for `legalEntityName` ([Branded data holder instances](2026-11-01-branded-data-holder-instances.md), section 6).
- Invented host names use the `.test` top-level domain, which RFC 6761 reserves for tests (RFC6761).
- The names in this note follow one theme, Australian flora and fauna. No name copies a word of the name of its inspiration.

### 4.2 Candidate names (for review)

| Inspiration | Recommended | Second | Legal entity name (characters) | Host name |
|---|---|---|---|---|
| CommBank | **Wattle Bank** | Goldcrest Bank | Wattle Bank Limited (19) | `wattlebank.test` |
| NAB | **Jarrah Bank** | Redgum Bank | Jarrah Bank Limited (19) | `jarrahbank.test` |
| ANZ | **Harbour Bank** | Tasman Bank | Harbour Bank Limited (20) | `harbourbank.test` |
| Westpac | **Kestrel Bank** | Sandstone Bank | Kestrel Bank Limited (20) | `kestrelbank.test` |

Results of the checks:

- No candidate is on the APRA register (APRA). A light web search found no Australian bank with a candidate name.
- "Waratah" is not a candidate: Waratah Securities Australia Limited is linked to Westpac, and the waratah is the emblem of New South Wales, where Westpac started. A red "Waratah" brand could suggest Westpac. (Search result only, UNVERIFIED.)
- "Redgum": Redgum Financial is a small Australian finance and mortgage-broking business. Thus Jarrah is the recommendation, and Redgum is the second choice.
- "Tasman" suggests "trans-Tasman", and thus ANZ (Australia and New Zealand). Historical Tasmanian banks also used the word. Thus it is the second choice only.
- "Ironbark" is not a candidate: Ironbark Financial Group is a current financial brand.
- "Goldfields" is not a candidate: Goldfields Money was an ADI.
- "Federation" and "Commonwealth" are not candidates, because they have the same meaning as the name of CommBank.
- Kestrel starts with K. A brand with a W (the Westpac logo) is not a candidate for the Westpac palette.
- I did not search the IP Australia trade mark register. Do a search at https://search.ipaustralia.gov.au/trademarks/search/quick before the names merge (UNVERIFIED).

### 4.3 Product names and BSBs in the mock-ups

The project does not copy the names, fees or rates of real products. The first draft of the gallery used real product names ("Goal Saver", "Online Saver", "eSaver", "Choice", "Everyday Access", "Bonus Saver") and the real CommBank BSB 062-000. This draft replaces them.

| Brand | Transaction account | Savings account | BSB |
|---|---|---|---|
| Wattle Bank | Wattle Pocket | Seedling Saver | 372-100 |
| Jarrah Bank | Jarrah Daybook | Heartwood Saver | 373-100 |
| Harbour Bank | Harbour Mooring | Tidewater Saver | 374-100 |
| Kestrel Bank | Kestrel Perch | Updraft Saver | 375-100 |

- Product names: a light web search found no Australian bank product with these names. "Everyday Access" (Bank Australia) and "Bonus Saver" (Bank Australia, Greater Bank, RACQ Bank) are real products, so they are not used. "Updraft Saver" is near "Up Saver" (Up, issued by Bendigo and Adelaide Bank), but it is a different name (JUDGEMENT). The transaction account names start with the brand name, so that they cannot match a product of a different bank.
- BSB structure: a BSB has six digits. AusPayNet gives each institution a two-digit or three-digit code, and the institution gives the other digits. Most two-digit codes of the first digits 0 to 9 are in use, for example 01 ANZ, 03 and 73 Westpac, 06 and 76 CBA, 08 and 78 NAB ([Wikipedia: Bank state branch](https://en.wikipedia.org/wiki/Bank_state_branch)). AusPayNet publishes no list of free codes.
- Measured on 2026-11-01 with the search of the AusPayNet BSB directory (https://bsb.auspaynet.com.au/): the prefix `37` has only `370-010` to `370-100` and `370-370` (Flash Payments, FPX). No BSB in the directory starts with `371` to `379`. The prefix `36` has only `369-369` (BNK Banking Corporation).
- No range is clearly free: AusPayNet can give a code that has no BSB yet, or give one later. `372` to `375` are the safest option that I found, because the directory has no BSB in them now. Do a check again before the product data merges (UNVERIFIED that the codes are not allocated).

## 5. Logos

The logos are hand-written SVG. The wordmarks are outlines of the brand font, made with fontTools, so the files do not need a font. Each brand has four files: logo (light background), logo-dark, mark and mark-dark. Logos are exempt from the WCAG contrast criteria, but each logo and mark has a contrast of 3:1 or more against its background, except where the table below says otherwise.

| Brand | Mark | Different from the real logo because |
|---|---|---|
| Wattle Bank | A black disc with a yellow wattle sprig (three round flower heads on a stem). On dark: a yellow disc with a black sprig. | The CommBank logo is a yellow and black diamond. The mark is round, has no diamond and no split. |
| Jarrah Bank | A red curved gum leaf with a white vein. The small mark is a red rounded square with a white leaf. | The NAB logo is a red star. The mark has no points. |
| Harbour Bank | A navy disc with a light-blue sun and two waves. On dark: a light-blue disc with navy waves. | The ANZ logo is the letters ANZ with a figure symbol. The mark has no letters and no figures. |
| Kestrel Bank | A navy rounded square with a white K and a red upper arm. On dark: a white square with a navy K. | The Westpac logo is a red W with no tile. The mark is a different letter in a tile. |

Contrast of the logos (measured): the red Jarrah tile (`#C20000`) on the dark gallery background (`#16181D`) is 2.79:1. The dark variant of the mark uses `#E03A3A`, and the white leaf in the tile has 6.38:1 on `#C20000`. All other logo colours have 3.07:1 or more on their background.

## 6. Palettes and contrast

To look similar to its inspiration, the palette of each brand uses the colours of section 2.1. I changed a colour only where a WCAG 2.2 AA pair failed, or where a role had no real colour. Each change is in the table. The gallery gives the full contrast table of each brand.

### 6.1 Wattle Bank (CommBank palette)

| Role | Value | Source |
|---|---|---|
| Primary action, header | `#FFCC00` | CBA |
| Text on primary, body text, heading, hero band | `#231F20` | CBA |
| Primary hover / pressed | `#F2B400` / `#E0A400` | CBA log-on gradient / derived |
| Link, focus ring | `#0F6CA8` | CBA `#1175B5`, darker: `#1175B5` on the grey page is 4.51:1, too close to the limit |
| Muted text | `#5C595A` | derived from CBA greys |
| Error | `#C8001A` | CBA `#E1001A`, darker |
| Page / card / divider / input border | `#F4F4F4` / `#FFFFFF` / `#D3D2D2` / `#706D6E` | CBA |

### 6.2 Jarrah Bank (NAB palette)

| Role | Value | Source |
|---|---|---|
| Primary action, link, top rule of the header | `#C20000` | NAB |
| Primary hover / pressed | `#8F0000` / `#750000` | NAB |
| Body text, hero band, focus ring | `#26343A` | NAB |
| Notice text | `#3A4F59` | NAB |
| Heading / muted text | `#1B2429` / `#4D5A61` | derived from the NAB slate |
| Page / card / input border | `#F5F5F5` / `#FFFFFF` / `#6B7682` | NAB |

### 6.3 Harbour Bank (ANZ palette)

| Role | Value | Source |
|---|---|---|
| Header, heading | `#004165` | ANZ |
| Primary action, link, focus ring | `#0063CD` | ANZ |
| Primary hover / pressed | `#0056B2` / `#002B59` | ANZ |
| Accent (logo, on navy only) | `#8ADAF9` | ANZ |
| Hero band, notice | `#E2F6FD` | ANZ |
| Body text / muted text | `#2B2B2B` / `#5C5C5C` | ANZ `#2B2B2B`; muted darker than ANZ `#666666` |
| Error | `#C12D23` | ANZ |

### 6.4 Kestrel Bank (Westpac palette)

| Role | Value | Source |
|---|---|---|
| Primary action, link | `#DA1710` | WBC, GEL |
| Primary hover / pressed | `#B9140E` / `#99100B` | WBC |
| Heading, hero band, focus ring | `#1F1C4F` | WBC |
| Accent (on navy only) | `#FF3DDB` | WBC |
| Body text / muted text | `#161619` / `#5B5B6A` | WBC muted 950 / 600 |
| Page / input border / divider | `#F5F5F6` / `#706F7C` / `#DEDEE1` | WBC muted scale |

### 6.5 Contrast results (measured, WCAG 2.2 formula)

| Pair | Minimum | Wattle | Jarrah | Harbour | Kestrel |
|---|---|---|---|---|---|
| Body text on page | 4.5 | 14.82 | 11.78 | 12.79 | 16.57 |
| Heading on card | 4.5 | 16.30 | 15.78 | 10.80 | 15.70 |
| Muted text on page | 4.5 | 6.30 | 6.53 | 6.04 | 6.12 |
| Button label on primary | 4.5 | 10.78 | 6.38 | 5.73 | 5.11 |
| Button label on hover | 4.5 | 8.77 | 9.69 | 7.06 | 6.63 |
| Button label on pressed | 4.5 | 7.36 | 11.96 | 14.10 | 8.63 |
| Link on page | 4.5 | 5.12 | 5.85 | 5.18 | 4.69 |
| Header text on header | 4.5 | 10.78 | 12.85 | 10.80 | 15.70 |
| Hero text on hero | 4.5 | 16.30 | 12.85 | 9.69 | 15.70 |
| Notice text on notice | 4.5 | 5.01 | 7.58 | 9.69 | 13.90 |
| Error text on card | 4.5 | 6.05 | 6.38 | 5.73 | 6.63 |
| Focus ring on card | 3 | 5.63 | 12.85 | 5.73 | 15.70 |
| Input border on card | 3 | 5.12 | 4.63 | 4.54 | 4.94 |
| Primary button on card | 3 | **1.51** | 6.38 | 5.73 | 5.11 |

- All text pairs pass AA.
- The yellow Wattle button on a white card is 1.51:1. WCAG 1.4.11 does not require a visible boundary for a button that has a text label (WCAG, Understanding 1.4.11), and the label is 10.78:1. Thus it passes, but a person with low vision can see the button shape less well. CommBank uses yellow buttons in the same way (CBA). Option: a 1 px `#231F20` border on the yellow button.
- Kestrel link on the grey page is 4.69:1. It passes, with a small margin. Links are on the white card in the proposed layout (5.11:1).
- Jarrah and Kestrel use the same red for the primary action and for errors. An error then looks like a brand element. The error message needs an icon and text, not colour only (WCAG 1.4.1) (JUDGEMENT).

## 7. Proposed design direction (JUDGEMENT)

### 7.1 Common to all brands

- **Layout:** mobile first. One column, maximum width 440 px for the card on a large screen. From top to bottom: a header with the logo; a hero band with the data recipient (logo and brand name) and one sentence; a raised card with the step, the heading, the form and the actions; the password notice; the simulation footer.
- **Steps:** "Step 1 of 3" and a thin bar: customer number, OTP, consent. The CX guidelines do not require it. It tells the consumer that the flow is short.
- **Footer:** "<Name> is a simulation for the cdr-lab project. It is not a real bank, and it is not affiliated with a bank." No other links, because of the friction standard (section 3.3).
- **Spacing:** the 4 px base of the tokens. Add a `comfortable` context to the `density` modifier for consumer pages: controls 48 px high (WCAG 2.5.8 needs 24 px), card padding 20 to 24 px, 16 to 24 px between groups.
- **Type:** body 15 to 16 px, line height 1.5, heading 24 px (weight 700, letter spacing -0.015em), section heading 16 px, small text 12 to 13 px. Input text 17 px, so that iOS Safari does not zoom. The OTP field is one input with `autocomplete="one-time-code"` and wide letter spacing, not one box for each digit (WCAG 3.3.8).
- **Elevation:** the card uses two shadow layers (1 px and a soft 32 px) on a grey page. The card overlaps the hero band by 24 px. This gives depth, which the flat consoles do not have.
- **Motion:** 120 ms for hover and press, with `cubic-bezier(0.2, 0, 0, 1)`, and a 1 px press movement. No page transitions. `prefers-reduced-motion` removes all movement.
- **Focus:** a 3 px outline with a 2 px offset, in the focus colour of the brand.
- **Theme:** light only, as ADR 0016 says. A dark theme is a later change.
- **Fonts:** one variable font for each brand, self-hosted (for example from Fontsource), SIL Open Font License (GF). JetBrains Mono stays the code font.

### 7.2 Different for each brand

| Item | Wattle | Jarrah | Harbour | Kestrel |
|---|---|---|---|---|
| Font | Figtree | Public Sans | Geist | DM Sans |
| Header | yellow band | white, 4 px red top rule | navy band | white |
| Hero band | near black, yellow accent | slate | light blue tint | navy, pink accent |
| Card radius | 14 px | 4 px | 8 px | 16 px |
| Button radius | 8 px | 4 px | 4 px | pill |
| Text on primary | dark | white | white | white |

The font choices are free fonts with a similar character to the real font: Figtree for the round geometric CBA font, Public Sans for a firm grotesque, Geist for the modern Aeonik, and DM Sans for the geometric Brown. NAB uses the free Source Sans; I did not use it, so that the Jarrah pages look less like the real NAB pages.

### 7.3 Mapping to DTCG tokens

- Add one context for each brand to the `brand` modifier: `brand/wattle.json`, `brand/jarrah.json`, `brand/harbour.json` and `brand/kestrel.json`. The `lab` brand stays for the consoles.
- Put the bank colours in the primitive tier as named scales of each brand (for example `primitive.color.wattle.yellow`), and refer to them from the brand context. A component never uses a primitive token, as now.
- Keep the existing semantic roles. Add these roles: `color.heading`, `color.link`, `color.brand` and `color.fg-on-brand` (header), `color.hero` and `color.fg-on-hero`, `color.accent` (decorative only, never for text), and `font.display` if a brand needs a second font later.
- Add component-tier tokens only where a brand changes the shape: `radius.card` and `radius.control`.
- Add the `comfortable` context to the `density` modifier.
- The brand directory of the project (an ADR records its format) contains the token file of the brand, the logos and the font files. Terrazzo builds the CSS custom properties of one brand. The Go templates load the same CSS, as ADR 0016 permits.

## 8. Open questions

| Question | Where it is resolved |
|---|---|
| Are the names and logos far enough from the real banks? Is a trade mark search at IP Australia necessary before the merge? | The review of the gallery |
| A yellow primary button with no border (as CommBank), or with a 1 px dark border for low vision? | The review of the design direction |
| Does the step indicator count as friction? The standards do not say. | The review of the design direction |
| Must the brand directory contain the font files, or only a font name from a fixed set in the repository? A user who adds a fifth brand then needs no code change in both cases, but font files add size and a licence check. | The ADR of the brand directory |
| The consumer pages are Go templates. Do they use the Tailwind CSS of `web/packages/ui`, or a small plain CSS file that uses only the token custom properties? | The ADR or DESIGN.md of milestone 1 |
