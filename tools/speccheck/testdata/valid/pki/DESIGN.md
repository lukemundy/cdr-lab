# Mock CDR PKI

The mock PKI issues the certificates of the lab. The PKI writes dates as ISO-8601 text.

## Context

The context diagram.

## Library specifications

| Library | Specification |
|---|---|
| `pkg/ca` | None. |

## Goals

- The PKI gives the lab a CDR CA.

## Non-goals

- The PKI does not run an OCSP responder.

## Certificates

One paragraph that summarises the group, with the code `PKI-0001` in a code span.

- <a id="pki-0001"></a>**PKI-0001** The CDR CA signs each client certificate. ([CP section 6][accc-cp])
- <a id="pki-0002"></a>**PKI-0002** The lab operator can revoke a certificate with one of these reasons. ([RFC 5280][rfc5280]; [Judgement: a long root validity](#pki-0004))
  - `unspecified`
  - `keyCompromise`

| ID | Certificate | Validity | Source |
|---|---|---|---|
| <a id="pki-0003"></a>**PKI-0003** | Root CA | 20 years | [Judgement: a long root validity](#pki-0004); observed on 2026-10-02 |

```text
A code block is not checked: PKI-0099 [a rule](#pki-0099)
```

## Judgements

### <a id="pki-0004"></a>PKI-0004 A long root validity

- Kind: choice
- Basis: Step 3.3. The CP gives no limit ([CP section 6][accc-cp]).
- Reason: A long validity is the least complex. The lab does not renew [the CA that signs client certificates](#pki-0001).

## Withdrawn

- <a id="pki-0005"></a>**PKI-0005** Removed.

[accc-cp]: https://example.com/cp
[rfc5280]: https://www.rfc-editor.org/rfc/rfc5280
