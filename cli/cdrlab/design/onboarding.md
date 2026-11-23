# cdrlab: Onboarding

This area gives the commands that add participants to the Register and customers to a data holder brand, in a compose lab or a cluster deployment. They send a file to the add-only load of an operator interface ([ADR 0040][adr-0040]). The umbrella chart runs them with the seed data of its values, and a lab user runs them with the same files.

## Participants and customers

- <a id="cli-0179"></a>**CLI-0179** The participant add takes the URL of the Register operator interface and the path of a load file. It sends the file to [the participant load of the Register](../../../services/register/design/operator-interface.md#reg-1531). ([Judgement: add-only loads of the operator interfaces (ADR 0040)][adr-0040])
- <a id="cli-0180"></a>**CLI-0180** The customer add takes the URL of the operator interface of a data holder brand and the path of a customer load file. It sends the file to [the customer load of the data holder](../../../services/dataholder/design/operator-interface.md#dh-2649). ([Judgement: add-only loads of the operator interfaces (ADR 0040)][adr-0040])
- <a id="cli-0181"></a>**CLI-0181** The participant add reads the operator token from the environment variable `REGISTER_OPERATOR_TOKEN`. The customer add reads it from `DATAHOLDER_OPERATOR_TOKEN`. ([Judgement: the operator token only from the environment](#cli-0187))
- <a id="cli-0182"></a>**CLI-0182** No input of the two commands has a default. When an input is absent, the command fails, and the error names the input. ([Judgement: no default for a value of one deployment](../../../services/dataholder/design/deployment.md#dh-1375))
- <a id="cli-0183"></a>**CLI-0183** The command sends the file in one request, with no change. The service checks the records. ([Judgement: add-only loads of the operator interfaces (ADR 0040)][adr-0040])
- <a id="cli-0184"></a>**CLI-0184** When the service refuses the file, the command fails. The error gives the status and the `detail` of the response. ([RFC 9457][rfc9457])
- <a id="cli-0185"></a>**CLI-0185** After a load, the participant add gives each identifier of [the result of the load](../../../services/register/design/operator-interface.md#reg-1537), with the word `added` or `existing`. ([Judgement: add-only loads of the operator interfaces (ADR 0040)][adr-0040])
- <a id="cli-0186"></a>**CLI-0186** After a load, the customer add gives the number of customers that the load created and the number that existed. It gives no value of a customer. For example: `2 customers created, 1 existed`. ([Judgement: no value of a customer in the output](#cli-0188))
- <a id="cli-0189"></a>**CLI-0189** The participant add and the customer add make 0 calls to the Kubernetes API. ([Judgement: add-only loads of the operator interfaces (ADR 0040)][adr-0040])

## Onboarding

- <a id="cli-0190"></a>**CLI-0190** The onboarding of a software product is two commands: [the participant add](#cli-0179) of its data recipient, brand and software product, and then [the certificate command](certificates.md#cli-0030) for its client certificate. ([Judgement: add-only loads of the operator interfaces (ADR 0040)][adr-0040]; [Register: the issue registers the certificate](../../../services/register/design/certificate-management.md#reg-0610))
- <a id="cli-0191"></a>**CLI-0191** The onboarding of a data holder brand is two commands: [the participant add](#cli-0179) of its legal entity and brand, and then [the certificate command](certificates.md#cli-0030) for its server certificate. ([Judgement: add-only loads of the operator interfaces (ADR 0040)][adr-0040])
- <a id="cli-0192"></a>**CLI-0192** `cdrlab` has no command that does the two steps of an onboarding together. ([Judgement: add-only loads of the operator interfaces (ADR 0040)][adr-0040])

## Judgements

### <a id="cli-0187"></a>CLI-0187 The operator token only from the environment

- Kind: choice
- Basis: Step 3.2. [CWE-214][cwe-214] lists a secret in the arguments of a process as a weakness, because other users of the host can see the arguments. The certificate command reads the token of the Register from `REGISTER_OPERATOR_TOKEN`.
- Reason: A Job gives the token from a Secret as an environment variable. One variable for each service lets a deployment give a different token to each service.

### <a id="cli-0188"></a>CLI-0188 No value of a customer in the output

- Kind: choice
- Basis: Step 3.2. The output of a container of a Job is its log ([Kubernetes: logging architecture][k8s-logging]). [No personal information in telemetry, also when it is mocked](../../../pkg/platform/telemetry/DESIGN.md#tele-0004).
- Reason: The lab user wrote the file, so the counts are sufficient to see the result. A consumer signs in with the email address of the file, so the customer number is not necessary.

[adr-0040]: ../../../docs/adr/0040-add-participants-and-customers-with-add-only-loads-of-the-operator-interfaces-from-cdrlab-and-from-the-umbrella-values.md
[cwe-214]: https://cwe.mitre.org/data/definitions/214.html
[k8s-logging]: https://kubernetes.io/docs/concepts/cluster-administration/logging/
[rfc9457]: https://www.rfc-editor.org/rfc/rfc9457
