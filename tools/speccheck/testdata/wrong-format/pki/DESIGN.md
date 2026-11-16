# Mock CDR PKI

## Library specifications

None.

## Certificates

- <a id="PKI-0001"></a>**PKI-0001** The CDR CA signs each client certificate. ([CP section 6][accc-cp])
- <a id="pki-002"></a>**PKI-002** The CDR CA signs each server certificate. ([CP section 6][accc-cp])
- <a id="reg-0003"></a>**REG-0003** The CDR CA signs each CRL. ([CP section 6][accc-cp])
- The CDR CA publishes the CRL each day. ([CP section 6][accc-cp])

| Name | Validity | Source |
|---|---|---|
| Root CA | 20 years | [CP section 6][accc-cp] |

## Judgements

### PKI-0005 A long root validity

- Kind: choice
- Basis: Step 3.3. The CP gives no limit ([CP section 6][accc-cp]).
- Reason: A long validity is the least complex.

### <a id="pki-0006"></a>PKI-0006 One CRL for each CA

- Kind: opinion
- Basis: Step 3.3. The CP gives no rule ([CP section 6][accc-cp]).
- Reason: One CRL is the least complex.

## Withdrawn

- <a id="pki-0007"></a>**PKI-0007** Deleted.

[accc-cp]: https://example.com/cp
