# cdrlab: Certificates

This area gives the set-up of the mock CDR CA and the mock public CA, the CDR certificates that `cdrlab` gets through the Register, and the public TLS certificates that it issues.

## CA set-up

- <a id="cli-0010"></a>**CLI-0010** With Secrets as the output, the CA set-up makes the mock CDR CA and writes two Secrets: the intermediate Secret and the CDR CA bundle Secret. The command takes the name of each Secret. ([Judgement: CDR certificates from the Register (ADR 0032)][adr-0032])
- <a id="cli-0011"></a>**CLI-0011** The CA set-up makes each CA with the profiles of the CA library, and each CA key is an RSA 4096 key. ([CA library: one root CA and one issuing intermediate CA](../../../pkg/ca/DESIGN.md#ca-0001); [CA library: a public CA that is separate from the CDR CA](../../../pkg/ca/DESIGN.md#ca-0009); [PKI: the CA keys](../../../pki/DESIGN.md#pki-0003))
- <a id="cli-0012"></a>**CLI-0012** The intermediate Secret has the type `kubernetes.io/tls`. `tls.crt` has the intermediate certificate, and `tls.key` has the intermediate key. ([Kubernetes: TLS Secrets][k8s-tls-secret])
- <a id="cli-0013"></a>**CLI-0013** The CDR CA bundle Secret has the CDR CA bundle in the key `ca.crt`: the intermediate certificate and then the root certificate, in PEM. ([relying parties trust the CDR CA bundle](../../../pki/DESIGN.md#pki-0013))
- <a id="cli-0014"></a>**CLI-0014** When the two Secrets exist, the CA set-up changes nothing, and it exits with the status 0. ([Judgement: the CA set-up changes no CA that exists](#cli-0100))
- <a id="cli-0015"></a>**CLI-0015** When only one of the two Secrets exists, the CA set-up changes nothing, and it fails. The error names the Secret that is absent. ([Judgement: the CA set-up changes no CA that exists](#cli-0100))
- <a id="cli-0016"></a>**CLI-0016** Security: the CA set-up writes the root key of the mock CDR CA to 0 Secrets and 0 files. ([Judgement: CDR certificates from the Register (ADR 0032)][adr-0032])
- <a id="cli-0017"></a>**CLI-0017** When the CA set-up writes the first Secret and the write of the second Secret fails, it deletes the first Secret. Thus a failed run leaves no Secret of the two. ([Judgement: the CA set-up changes no CA that exists](#cli-0100))
- <a id="cli-0119"></a>**CLI-0119** With a directory as the output, the CA set-up makes the mock CDR CA and the mock public CA, and writes these files in the directory. ([PKI: only the owner reads a root key](../../../pki/DESIGN.md#pki-0073); [PKI: the deployment sets who reads the intermediate key](../../../pki/DESIGN.md#pki-0074))
  - `cdr/root.crt`, `cdr/intermediate.crt`, `cdr/ca-bundle.pem` and `public/root.crt`, which any user can read
  - `cdr/intermediate.key`, which any user can read
  - `public/root.key`, which only the owner can read
- <a id="cli-0120"></a>**CLI-0120** `cdr/ca-bundle.pem` is the CDR CA bundle: the intermediate certificate and then the root certificate, in PEM. ([relying parties trust the CDR CA bundle](../../../pki/DESIGN.md#pki-0013))
- <a id="cli-0121"></a>**CLI-0121** When each file of the directory exists, the CA set-up changes nothing, and it exits with the status 0. ([Judgement: the CA set-up changes no CA that exists](#cli-0100))
- <a id="cli-0122"></a>**CLI-0122** When only some files of the directory exist, the CA set-up changes nothing, and it fails. The error names a file that is absent. ([Judgement: the CA set-up changes no CA that exists](#cli-0100))
- <a id="cli-0123"></a>**CLI-0123** When a write of a file fails, the CA set-up deletes each file that it wrote in that run. ([Judgement: the CA set-up changes no CA that exists](#cli-0100))

## CDR certificate

- <a id="cli-0030"></a>**CLI-0030** The certificate command gets a CDR certificate through the Register issue path, and writes it to the output: a Secret or files. ([Judgement: CDR certificates from the Register (ADR 0032)][adr-0032])
  - a server certificate for a data holder brand, [through its issue path](../../../services/register/design/certificate-management.md#reg-0634)
  - a client certificate for a software product, [through its issue path](../../../services/register/design/certificate-management.md#reg-0632)
- <a id="cli-0031"></a>**CLI-0031** The command takes these inputs. ([Register: certificate management](../../../services/register/design/certificate-management.md#reg-0600))
  - the URL of the Register operator interface, and the operator token
  - the kind of the certificate, and the identifiers of the legal entity and of the brand or the software product
  - the CN and the O of the subject
  - for a server certificate, the DNS names
  - the output: the name of the Secret, or the certificate file and the key file
- <a id="cli-0032"></a>**CLI-0032** The command makes a new RSA 2048 key and a CSR with [the CSR profile of the CA library](../../../pkg/ca/DESIGN.md#ca-0010). ([CDS: Certificate Signing Request Profile][cds-csr-profile])
- <a id="cli-0033"></a>**CLI-0033** Security: the command sends the CSR only, never the private key. ([Register: no private key](../../../services/register/design/certificate-management.md#reg-0603))
- <a id="cli-0034"></a>**CLI-0034** With a Secret as the output, the Secret has the type `kubernetes.io/tls`. ([Kubernetes: TLS Secrets][k8s-tls-secret])
  - `tls.crt`: the certificate chain of the issue response
  - `tls.key`: the private key
  - `ca.crt`: the CDR CA bundle from the Register
- <a id="cli-0035"></a>**CLI-0035** When the output has a certificate that meets each of these conditions, the command changes nothing, and it exits with the status 0. ([Judgement: a renewal window of 30 days](#cli-0101))
  - It expires more than the renewal window after the time of the run.
  - It chains to the CDR CA bundle of the Register.
  - It has the subject and the DNS names of the inputs.
  - Its key agrees with the key of the output.
- <a id="cli-0036"></a>**CLI-0036** The renewal window is 30 days. An input can give a different window. ([Judgement: a renewal window of 30 days](#cli-0101))
- <a id="cli-0037"></a>**CLI-0037** Integrity: when the issue fails, the command does not change the output, and it fails. For example, the Register is not ready, or it refuses the CSR. ([Judgement: a renewal window of 30 days](#cli-0101))
- <a id="cli-0038"></a>**CLI-0038** When the Register refuses the CSR, the error gives the status and the `detail` of the response. ([RFC 9457][rfc9457])
- <a id="cli-0039"></a>**CLI-0039** The command does not revoke the certificate that it replaces. ([Judgement: no revocation of a replaced certificate](#cli-0102))
- <a id="cli-0124"></a>**CLI-0124** With files as the output, the command writes the certificate chain of the issue response to the certificate file, and the private key to the key file, in PEM. ([RFC 8555, section 9.1][rfc8555-chain]; [RFC 7468][rfc7468])
- <a id="cli-0125"></a>**CLI-0125** Only the owner can read a key file that the command writes. When an input permits it, any user can read the key file. ([Judgement: a key file for the owner only, by default](#cli-0171))
- <a id="cli-0126"></a>**CLI-0126** Integrity: the command replaces each file as one unit. A reader of the file gets the earlier file or the new file, never a part of a file. ([the reload of a certificate file](../../../pkg/cds/DESIGN.md#cdsl-0089))
- <a id="cli-0127"></a>**CLI-0127** When an input gives a CSR file, the command sends that CSR, and it writes only the certificate file. It takes the subject and the DNS names from the CSR. ([Judgement: an issue from the CSR of a lab user](#cli-0172))
- <a id="cli-0128"></a>**CLI-0128** When an input gives a CSR file and also a subject, a DNS name or a key file, the command fails. ([Judgement: an issue from the CSR of a lab user](#cli-0172))
- <a id="cli-0129"></a>**CLI-0129** With a CSR file, the key of [the renewal check](#cli-0035) is the public key of the CSR. ([Judgement: an issue from the CSR of a lab user](#cli-0172))

## Public TLS certificate

- <a id="cli-0130"></a>**CLI-0130** The public certificate command issues a TLS server certificate from the mock public CA in the directory of the input, for the DNS names of the input. The first DNS name is the CN. ([CA library: the profile of a public TLS server certificate](../../../pkg/ca/DESIGN.md#ca-0038); [CA library: the subject of a public TLS server certificate](../../../pkg/ca/DESIGN.md#ca-0039))
- <a id="cli-0131"></a>**CLI-0131** The command makes a new RSA 2048 key, and writes the certificate file and the key file with the rules of the [CDR certificate files](#cli-0124). ([Judgement: public TLS certificates from the files of the public CA](#cli-0170))
- <a id="cli-0132"></a>**CLI-0132** When the certificate file has a certificate that meets each of these conditions, the command changes nothing, and it exits with the status 0. ([Judgement: a renewal window of 30 days](#cli-0101))
  - It expires more than the renewal window after the time of the run.
  - It chains to the public root certificate of the directory.
  - It has the DNS names of the inputs.
  - Its key agrees with the key file.
- <a id="cli-0133"></a>**CLI-0133** The command makes 0 network calls. ([Judgement: public TLS certificates from the files of the public CA](#cli-0170))

## Judgements

### <a id="cli-0100"></a>CLI-0100 The CA set-up changes no CA that exists

- Kind: choice
- Basis: Step 3.3. Each certificate that the Register issued chains to the intermediate, and the Register keeps the revocations of that intermediate ([PKI: the Register keeps the only list of revocations](../../../pki/DESIGN.md#pki-0050)).
- Reason: A new CA under a running Register makes each issued certificate fail. One Secret or file with no other is a broken state, and a fresh deployment is the simplest correction.

### <a id="cli-0101"></a>CLI-0101 A renewal window of 30 days

- Kind: choice
- Basis: Step 3.3. A CDR certificate is valid for 13 months ([CA library: the validity of a CDR certificate](../../../pkg/ca/DESIGN.md#ca-0026)). The CDS gives no renewal time.
- Reason: A daily run then has 30 tries before a certificate expires, and the old certificate stays in use after a failed try. A renewal also occurs when the CA or the names change, so a change of the values takes effect at the next run.

### <a id="cli-0102"></a>CLI-0102 No revocation of a replaced certificate

- Kind: choice
- Basis: Step 3.3. The CDS permits a renewal with the same subject ([CA library: no check of a unique subject DN](../../../pkg/ca/DESIGN.md#ca-0037)). A replaced lab certificate has no compromise.
- Reason: Its key is only in the output, which the command writes again. A revocation adds a call that can fail after a successful renewal.

### <a id="cli-0170"></a>CLI-0170 Public TLS certificates from the files of the public CA

- Kind: choice
- Basis: Step 3.3. The Register is not a public CA ([Judgement: no server certificate for a data recipient](../../../services/register/design/certificate-management.md#reg-0647)), and the public CA is separate from the CDR CA ([CA library: a public CA that is separate from the CDR CA](../../../pkg/ca/DESIGN.md#ca-0009)).
- Reason: The public root key is only in the lab directory, so the host issues the public TLS certificates of the compose lab. A cluster deployment gets them from cert-manager.

### <a id="cli-0171"></a>CLI-0171 A key file for the owner only, by default

- Kind: choice
- Basis: Step 3.2. [CWE-732][cwe-732] lists a permission that is too wide on a critical resource as a weakness. The containers of the compose lab read the keys as a different user ([Judgement: lab keys that any user can read](../../../pki/DESIGN.md#pki-0095)).
- Reason: The default must suit each deployment, so it is the narrow permission. `cdrlab up` gives the wider permission for the compose lab only.

### <a id="cli-0172"></a>CLI-0172 An issue from the CSR of a lab user

- Kind: choice
- Basis: Step 3.1. A participant gives a CSR, and the CA never gets its private key ([Register: no private key](../../../services/register/design/certificate-management.md#reg-0603)).
- Reason: A lab user can test software that makes its own key. The Register checks the names, so the command sends the CSR as it is.

[adr-0032]: ../../../docs/adr/0032-issue-each-cdr-certificate-of-a-deployment-through-the-register-and-renew-it-with-no-manual-step.md
[cds-csr-profile]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#certificate-signing-request-profile
[cwe-732]: https://cwe.mitre.org/data/definitions/732.html
[k8s-tls-secret]: https://kubernetes.io/docs/concepts/configuration/secret/#tls-secrets
[rfc7468]: https://www.rfc-editor.org/rfc/rfc7468
[rfc8555-chain]: https://www.rfc-editor.org/rfc/rfc8555#section-9.1
[rfc9457]: https://www.rfc-editor.org/rfc/rfc9457
