# 0014. Build web UIs as React single-page applications that the service that owns their data embeds and serves

- Date: 2026-10-22

## Context

The lab needs web UIs: operator consoles for the services (desktop), and later consumer UIs (mobile), for example a data holder consumer dashboard and data recipient apps. The users are the maintainer, people who see demonstrations, and engineers who run the lab to test their own CDR software. The maintainer and coding agents write the UIs. The consoles are data-dense: tables with sort, filter and pagination, forms, and status changes with a confirmation.

Each operator interface is now a JSON API on a separate listener, with one static bearer token from the configuration, and it refuses cross-origin requests. All lab ports bind to `127.0.0.1`. The data holder authorisation pages (login, one-time password, account selection, consent) are Go `html/template` pages with no JavaScript, and they send `Content-Security-Policy: default-src 'none'; frame-ancestors 'none'; base-uri 'none'`.

The research is in [Web front ends](../research/2026-10-22-web-front-ends.md). We compared three rendering architectures:

1. **A client single-page application (SPA), built to static files and served by the Go service.** No JavaScript runtime in the deployment. Prometheus embeds a React and Vite build in its Go binary ([`web/ui`](https://github.com/prometheus/prometheus/tree/main/web/ui)), and PocketBase embeds its build with `go:embed` ([`ui/embed.go`](https://github.com/pocketbase/pocketbase/blob/master/ui/embed.go)). Grafana, Jaeger and Argo CD also serve a React build from Go. A page does not work without JavaScript.
2. **A JavaScript meta-framework with server rendering (Next.js, React Router framework mode, TanStack Start, SvelteKit).** A Node server for each UI, beside the Go service. Most of the critical advisories of these frameworks in 2025 and 2026 are in their server code, for example a Next.js middleware authorisation bypass ([CVE-2025-29927](https://github.com/vercel/next.js/security/advisories/GHSA-f82v-jwr5-mffw)) and two remote code executions in React Server Components. The React advisory says: "If your app's React code does not use a server, your app is not affected" ([GHSA-xqc2-xf4c-x3h3](https://github.com/facebook/react/security/advisories/GHSA-xqc2-xf4c-x3h3)).
3. **Server-rendered hypermedia from Go (templ with htmx 4, or Datastar).** One language and no npm. Gitea removed htmx in 2026, because injected markup with `hx-*` attributes can send requests even under a nonce-based CSP, and because the attribute code is difficult to lint ([go-gitea/gitea#35059](https://github.com/go-gitea/gitea/issues/35059)). Datastar needs `unsafe-eval` in the CSP unless its CSP mode is on ([Datastar: Security](https://data-star.dev/reference/security)). We found no accessible component library for this model that is equivalent to those for React.

For a SPA, React has the largest set of accessible headless component libraries (React Aria, Base UI, Radix) and the most use (approximately 116 million npm downloads each week, against 10 million for Vue and 4 million for Svelte, from `api.npmjs.org` on 2026-10-22). Editor support for Vue and Svelte still needs TypeScript 6 ([Announcing TypeScript 7.0](https://devblogs.microsoft.com/typescript/announcing-typescript-7-0/)). React governance is in the React Foundation of the Linux Foundation ([The React Foundation](https://react.dev/blog/2026/02/24/the-react-foundation)).

A browser app that holds a token gives it to each script on the page: an XSS defect or a compromised dependency can take it. RFC 10017 "strongly recommends" a Backend for Frontend (BFF), which keeps the tokens on the server, for "applications that handle personal data" ([RFC 10017, section 6.1](https://www.rfc-editor.org/rfc/rfc10017#section-6.1)). "Cookies do not provide isolation by port" ([RFC 6265, section 8.5](https://www.rfc-editor.org/rfc/rfc6265#section-8.5)), so the lab UIs on `localhost` get the cookies of each other. Go `http.CrossOriginProtection` refuses a request that is not safe when its `Sec-Fetch-Site` is `same-site` or `cross-site` ([net/http](https://pkg.go.dev/net/http#CrossOriginProtection)).

Browser telemetry is a risk for ADR 0006: the CDR redirect URLs carry `code`, `state` and a JARM `response`, and the default URL sanitisers of the browser SDKs do not remove them. OpenTelemetry JS 3.0.0 removed `sdk-trace-web`, and its browser instrumentations are experimental ([migration guide](https://github.com/open-telemetry/opentelemetry-js/blob/main/doc/3.x/migration-guide.md)).

## Decision

This decision applies to the web UIs of the repository. It does not apply to the data holder authorisation pages. These stay server-rendered Go pages with no JavaScript, because they are the part of the authorisation server where the consumer authenticates, and a SPA gives the consumer nothing there. A later ADR can change this.

- Each web UI is a React SPA with client rendering only. It does not use server rendering, React Server Components or server functions. Node is a build tool only. No deployed component runs JavaScript on a server. The reason is the attack surface of option 2, and the cost of a second server runtime beside Go.
- Each UI uses React with the React Compiler, TanStack Router for routes and TanStack Query for server data. TanStack Router checks the types of routes and search parameters, and the filter and sort state of the console tables is in the search parameters.
- The service that owns the data of a UI embeds the build in its binary and serves it. The UI and its API have one origin, so they need no CORS. The UI and the service change and ship together.
- An operator console has its own listener in its service. That listener serves the UI and a same-origin JSON API for it. The console API is a driving adapter on the use cases of the service ([ADR 0013](0013-hexagonal-architecture-with-an-internal-composition-root.md)). It is separate from the operator interface, which stays as it is for scripts and tests. Thus the console API can give the views that the UI needs, and the operator interface does not change for the UI.
- The browser gets no credential, with one exception: a consumer UI that has a sign-in, for example the online bank site of a data holder, gets one session cookie. The cookie has only a random session ID, and the session stays on the server. The cookie is `HttpOnly`, `Secure` and `SameSite=Strict`, with the `__Host-` prefix, so no script on the page can read it, and a page of a different site cannot make the browser send it. Cookies are not isolated by port, so its name is different for each instance of the UI on one host. The reason for no token in the browser stays true: an XSS defect or a compromised dependency cannot take the cookie, and a session ends on the server at a sign-out. A server-side session cookie is the pattern that RFC 10017 permits for an application on one domain with no OAuth ([RFC 10017, section 7.1](https://www.rfc-editor.org/rfc/rfc10017#section-7.1)). A request that changes data still goes through `http.CrossOriginProtection`.
- A console gets no credential. In the lab, a console listener has no login. It binds to `127.0.0.1`, as all lab ports do. The operator token is public in the repository (`compose.yaml`), so a login gives no protection in the lab. A console listener uses `http.CrossOriginProtection`, so a page on a different lab port cannot send a change to it.
- A UI that is an OAuth client, for example a data recipient app, uses its service as a BFF: the tokens stay in the service ([RFC 10017](https://www.rfc-editor.org/rfc/rfc10017)).
- The UIs send no browser telemetry. The server spans of the API of each UI show each request. A later ADR can add browser telemetry, with a server-side point that removes data.

## Consequences

- No deployed component runs Node, so the JavaScript server advisories do not apply. The UI build needs Node and npm packages, which bring supply-chain risk to the build and to the browser.
- The image of a service with a UI needs a build stage for the UI before the Go build.
- A UI and its API always have the same version, so the console API needs no versioning. A UI cannot ship without its service.
- Each service with a console has one more port.
- The console API and the operator interface give some of the same functions. A change to a use case can need a change to both adapters.
- Any process on the lab host can use a console with no credential. A production service needs an identity for each operator, with a session.
- A consumer UI with a sign-in keeps its sessions in its service, so the service has state for each signed-in consumer. On a host with more than one UI, the browser also sends the session cookie to the other UIs of the host, which do not use it.
- A console does not work without JavaScript. Search engines and links that need server-rendered pages are not supported.
- If the authorisation pages get the design system later, there are two ways to render UI: React components, and Go templates with the same CSS.
- A JavaScript error in the browser is visible only in the browser. The telemetry backend shows only the requests to the API.
