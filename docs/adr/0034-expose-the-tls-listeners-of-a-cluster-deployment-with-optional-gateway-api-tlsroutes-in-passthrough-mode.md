# 0034. Expose the TLS listeners of a cluster deployment with optional Gateway API TLSRoutes in passthrough mode

- Date: 2026-11-23

## Context

Each service has an internal listener with plain HTTP, a public listener with TLS and an mTLS listener ([ADR 0025](0025-group-the-listeners-of-a-service-by-tls-mode-and-audience.md)). A data holder brand can also have the listener of its online bank site, with TLS. The services terminate TLS themselves. The mTLS listener needs the client certificate from end to end, so a proxy that terminates TLS cannot be in front of it.

The charts of the services make ClusterIP Services only. A team that deploys the lab in a shared cluster needs clients outside the cluster, for example the software of an ADR developer, to reach the public and the mTLS listeners. These facts constrain the choice:

- The Ingress API supports only rules for HTTP traffic, and has no standard field for TLS passthrough ([Kubernetes: Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)). ingress-nginx is retired and gets no security fixes ([Kubernetes blog: Ingress NGINX retirement](https://kubernetes.io/blog/2025/11/11/ingress-nginx-retirement/)).
- TLSRoute is in the Standard channel of the Gateway API since v1.5.0. In `Passthrough` mode, the Gateway does not decrypt, and it routes by the SNI of the TLS handshake ([Gateway API: TLSRoute](https://gateway-api.sigs.k8s.io/reference/api-types/tlsroute/), [Kubernetes blog: Gateway API v1.5](https://kubernetes.io/blog/2026/04/21/gateway-api-v1-5/)).
- In a shared cluster, a platform team usually owns the Gateway and its ports.
- The internal listener has the operator interface and the operator console. The console has no login. The Register serves its CRL on the same listener, and each relying party must get the CRL.

The real options were:

1. **Documentation only.** The chart has no route. Each team writes its own routes.
2. **Optional TLSRoute templates, and an HTTPRoute for the CRL path.** The values give the parent Gateway and the host names. The chart makes no Gateway.
3. **LoadBalancer Services.** No Gateway controller is necessary, but each brand needs an external address, and the cluster needs a load balancer implementation.
4. **Ingress templates with a passthrough annotation of one controller.** They work only with that controller.

## Decision

The umbrella chart gives optional Gateway API routes (option 2). They are off by default.

- When the routes are on, the chart makes one TLSRoute in `Passthrough` mode for each TLS listener of the release. Each route has its own host name, and refers to the existing Gateway that the values give. The chart makes no Gateway, because the Gateway belongs to the platform of the cluster.
- The public listener and the mTLS listener of a service have different host names. Thus one Gateway listener on port 443 can route all of them by SNI. The values can also give a port for each host name.
- The chart makes one HTTPRoute, for the CRL path of the Register only. It makes no route, and no Service of a different type, for any other path of an internal listener.
- In-cluster callers use the Service names. Each certificate of a listener has the Service name and the external host name, so that the two names work.
- The operator reaches the operator interface and the console with `kubectl port-forward`. Kubernetes RBAC then controls who can reach them. A guide gives oauth2-proxy as an option for a browser URL to the console only. oauth2-proxy cannot pass the static operator token, so the operator interface stays on port-forward.
- The chart has no Ingress template. A cluster with no TLSRoute support can use a LoadBalancer Service from the values of the common library chart.

## Consequences

- A team with a Gateway that supports TLSRoute exposes the lab with values only.
- Each brand needs two host names, and three with its online bank site. A team must make DNS records and public certificates for them.
- With passthrough, the services see the address of the Gateway as the client address.
- The base URIs that the Register gives use the external host names. A pod that calls a base URI must resolve the external name and reach the Gateway from the cluster.
- The routes can have a defect that CI does not find, because CI renders and checks them but runs no Gateway.
- Operator access needs a Kubernetes identity with the permission to port-forward in the namespace. A person with that permission can use the operator interface and see all the invented consumer data.
