# cdrlab: Non-functional requirements

This area gives the non-functional requirements of all commands of `cdrlab`.

## Non-functional requirements

- <a id="cli-0060"></a>**CLI-0060** Security: the commands write 0 private keys and 0 operator tokens to their output, their error output or their logs. ([Judgement: CDR certificates from the Register (ADR 0032)][adr-0032])
- <a id="cli-0061"></a>**CLI-0061** Telemetry: the commands send no telemetry. ([Judgement: the commands send no telemetry](#cli-0103))
- <a id="cli-0062"></a>**CLI-0062** Integrity: a run again of a command that completed, with the same inputs, changes nothing. ([Judgement: the CA set-up changes no CA that exists](certificates.md#cli-0100); [Judgement: a renewal window of 30 days](certificates.md#cli-0101))
- <a id="cli-0158"></a>**CLI-0158** Portability: in the release mode, `cdrlab up`, `cdrlab down` and `cdrlab reset` run 0 programs on the host other than Docker Compose. For example, no Go, Node, mise, git or openssl. ([Judgement: the compose lab from a compose project in the binary (ADR 0038)][adr-0038])

## Judgements

### <a id="cli-0103"></a>CLI-0103 The commands send no telemetry

- Kind: choice
- Basis: Step 3.3. A Job shows its result as its status and its logs ([Kubernetes: Jobs][k8s-job]).
- Reason: A command runs for seconds. The Register records each server certificate issue, and the trace of the Register shows each request.

[adr-0032]: ../../../docs/adr/0032-issue-each-cdr-certificate-of-a-deployment-through-the-register-and-renew-it-with-no-manual-step.md
[adr-0038]: ../../../docs/adr/0038-run-the-compose-lab-with-cdrlab-from-a-compose-project-in-the-binary-and-issue-its-cdr-certificates-from-the-host.md
[k8s-job]: https://kubernetes.io/docs/concepts/workloads/controllers/job/
