# Mock CDR PKI

The mock PKI issues the certificates of the lab.

## Library specifications

None.

## Goals

### Set-up

- The PKI gives the lab a CDR CA.

## Certificates

### Revocation

- <a id="pki-0001"></a>**PKI-0001** The CDR CA signs each client certificate. ([CP section 6][accc-cp])

```markdown
### A heading in a fenced code block
```

## Judgements

### <a id="pki-0002"></a>PKI-0002 A long root validity

#### Basis

- Kind: choice
- Basis: Step 3.3. The CP gives no limit ([CP section 6][accc-cp]).
- Reason: A long validity is the least complex.

## Withdrawn

None.

[accc-cp]: https://example.com/cp
