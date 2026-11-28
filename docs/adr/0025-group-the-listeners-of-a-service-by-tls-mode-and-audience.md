# 0025. Group the listeners of a service by TLS mode and audience, with a security policy for each group of routes

- Date: 2026-11-18

## Context

The Register and each data holder instance have six listeners each. The Register has the health probes, the public APIs, the operator interface, the secure APIs, the operator console and the CRL. A data holder has the online bank site in place of the CRL. In the lab, each instance publishes a block of six ports, and an engineer must know which port is which listener. In Kubernetes, each listener is one more container port, one more Service port and one more ingress or route backend. Each listener is also one more setting in each service, in its tests and in the compose file.

Most of these listeners exist because each one has a different security policy: the TLS mode, CORS, the `Host` allowlist, the format of a 404, and the response headers. The protocol does not need a different port for these. Only these differences need a different socket:

- The TLS mode: none, TLS, or mTLS.
- The origin. An origin is the scheme, the host and the port ([RFC 6454, section 4][rfc6454-origin]). The online bank site must have an origin that is different from the origin of the consumer pages ([data holder: the site and the consumer pages on different origins](../../services/dataholder/design/online-bank-site-sign-in.md#dh-2058)).

ADR 0014 gives each operator console a listener of its own ([ADR 0014](0014-web-uis-as-embedded-react-single-page-applications.md)). ADR 0022 gives the CRL a plain HTTP listener of the Register ([ADR 0022](0022-run-the-mock-cdr-intermediate-ca-in-the-register.md)). This ADR amends both. In ADR 0014, it replaces the statement "An operator console has its own listener in its service" and the consequence "Each service with a console has one more port". In ADR 0022, it replaces the statement that the Register has one more listener for the CRL.

These facts limit the options:

- A Go TLS server requests a client certificate only in the handshake. `ClientAuth` is a policy of the TLS configuration, not of a path ([Go: ClientAuthType][go-clientauth]). A Go server cannot start a renegotiation ([Go: RenegotiationSupport][go-renegotiation]), and the configuration has no setting for TLS 1.3 post-handshake authentication ([Go: Config][go-tls-config], [RFC 8446, section 4.6.2][rfc8446-pha]). HTTP/2 does not permit post-handshake authentication ([RFC 8740, section 3][rfc8740]).
- The CDS gives a public base URI and a resource base URI for each data holder brand ([CDS: Participant Endpoints][cds-participant-endpoints]). RFC 8705 gives a separate set of endpoint URLs for mTLS clients ([RFC 8705, section 5][rfc8705-aliases]).
- The kubelet sends an HTTP probe to the IP address of the pod, and the `Host` header is then that address ([Kubernetes: HTTP probes][k8s-http-probes]).
- Cookies are not isolated by port ([RFC 6265, section 8.5][rfc6265-ports]).

The options were:

1. **A listener for each TLS mode and audience, with a security policy for each group of routes on it.** One plain HTTP listener for the deployment and the lab operator, one TLS listener and one mTLS listener for the CDR APIs, and one TLS listener for the online bank site.
   - Advantages: four listeners or fewer for each service. Each listener has one TLS mode. The CDR listeners do not change.
   - Disadvantages: the policies of different interfaces are on one listener, so a defect in the routes can apply the wrong policy to a path. A catch-all route of a single-page application must not catch the paths of the other groups.
2. **One listener for TLS and mTLS.** The listener requests a client certificate and does not require it. Each mTLS path then checks that the client gave a certificate.
   - Advantages: one listener fewer.
   - Disadvantages: the client certificate becomes optional on all paths, because the server requests it only in the handshake. A browser shows a certificate prompt on the consumer pages when it has a client certificate. The CDS and RFC 8705 already give separate URLs for the mTLS endpoints.
3. **One listener that finds the protocol from the first bytes of each connection, for TLS and plain HTTP**, for example with [cmux][cmux].
   - Advantages: one port for each service.
   - Disadvantages: it is a workaround for a separation that the protocols already give. Each connection goes through more code before TLS, and the TLS mode still changes for each connection, not for each path.
4. **The online bank site on the public listener, with a host name of its own.** A different host gives a different origin.
   - Advantages: one listener fewer. A different host also isolates the cookies of the site, which a different port does not do.
   - Disadvantages: the lab PKI must issue a certificate with the new host name. The lab DNS needs an alias for each instance, and the tests change. A name under `.localhost` does not resolve on each host for each client, for example for HTTPie.
5. **The operator interface and the operator console on the public TLS listener.**
   - Advantages: no plain HTTP listener for the lab operator.
   - Disadvantages: the console has no login ([ADR 0014](0014-web-uis-as-embedded-react-single-page-applications.md)), so a deployment then puts it on the port that participants reach from the internet. The CDR rules of the public listener, for example CORS and the CDS 404, also do not apply to these interfaces.

## Decision

We use option 1. This decision applies to the HTTP listeners of the services. It does not apply to the HTTP clients of a service.

- Each service groups its listeners by TLS mode and audience. A service has these listeners, and only the listeners that its functions need:
  - The **internal listener**: plain HTTP, for the deployment and the lab operator. It serves the health probes, the operator interface and the operator console. On the Register, it also serves the CRL.
  - The **public listener**: TLS with no client authentication, for the public CDR APIs and the consumer pages.
  - The **mTLS listener**: TLS with a required client certificate, for the CDR endpoints that need mTLS.
  - The **online bank site listener**: TLS, on a data holder only.
  - The **web app listener**: TLS, on a data recipient only, for the web app of the data recipient and its API. It has the last digit of the online bank site listener, because both serve a consumer web UI with a session cookie.
- The mTLS endpoints stay on a listener of their own, because a Go server requests a client certificate only in the handshake. Option 2 makes the certificate optional on all paths.
- The online bank site keeps a listener of its own, because it must have an origin that is different from the consumer pages. The web app of a data recipient has a listener of its own, because its session cookie and the catch-all route of the web app must not be on the listener of the CDR endpoints of the data recipient. Option 4 needs changes to the PKI, the lab DNS and the tests to remove one port. A later decision can change this when a deployment gives each site its own host name.
- The operator interface and the console stay off the public listener, because the console has no login.
- The security policy of a listener becomes the policy of a group of routes on the listener. Each group has its own policy, and the policy of one group does not apply to the paths of a different group. For example, the `Host` allowlist of the operator interface does not apply to the health probes. A request for a path of no group gets 404.
- The health probes and the CRL do not check the `Host` header and need no authentication. The kubelet sends the IP address of the pod as the `Host`, and the CRL is signed public data.
- The operator interface and the console share one `Host` allowlist setting, because they are on one listener with one host and port. The setting is required when the internal listener serves the operator interface, and it has no default.
- The catch-all route of a web UI does not catch the paths of the other groups on its listener, for example the probes, the operator interface and the CRL.
- A process that runs only the background jobs serves only the health probes on its internal listener.
- In Kubernetes, the participants reach only the CRL path of the internal listener of the Register, for example through an ingress rule for that path only. The other routes of the internal listener stay inside the cluster.
- Each listener keeps the last digit of its port in the lab, so that each listener has the same last digit in each service.

## Consequences

- The Register has three listeners and each data holder instance has four. The compose file, the charts and the configuration of each service have fewer ports and fewer settings.
- A defect in the routes of the internal listener can give a path the policy of a different group. A separate listener for each policy cannot have this defect.
- A process that serves the operator interface needs the allowlist setting also when the console is off.
- Each CDR certificate that the Register issues gives a CRL URL on the internal listener. A lab made before the change needs `mise run reset`.
- The probes, the operator interface, the console and the CRL of the Register have one port. A deployment that exposes the CRL needs an ingress rule for one path, not for one port. The chart does not control the ingress of each deployment, so it cannot enforce this rule.
- The online bank site and the consumer pages keep separate ports on one host. The browser still sends the cookies of one to the other, as ADR 0014 accepts.
- The ports 8x82, 8x84 and 8085 go. Each script, test and document that uses them changes.

[cds-participant-endpoints]: https://consumerdatastandardsaustralia.github.io/standards-archives/standards-1.36.0/#participant-endpoints
[cmux]: https://github.com/soheilhy/cmux
[go-clientauth]: https://pkg.go.dev/crypto/tls#ClientAuthType
[go-renegotiation]: https://pkg.go.dev/crypto/tls#RenegotiationSupport
[go-tls-config]: https://pkg.go.dev/crypto/tls#Config
[k8s-http-probes]: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#http-probes
[rfc6265-ports]: https://www.rfc-editor.org/rfc/rfc6265#section-8.5
[rfc6454-origin]: https://www.rfc-editor.org/rfc/rfc6454#section-4
[rfc8446-pha]: https://www.rfc-editor.org/rfc/rfc8446#section-4.6.2
[rfc8705-aliases]: https://www.rfc-editor.org/rfc/rfc8705#section-5
[rfc8740]: https://www.rfc-editor.org/rfc/rfc8740#section-3
