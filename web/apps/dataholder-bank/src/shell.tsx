import { useRef, type ReactNode } from "react";

import { useBrand } from "./brand";

// The shell has no "Secure" label, no data recipient and no step, so that it does not look like the CDR authentication.
export function BankShell({
  home,
  navigation,
  footer,
  children,
}: {
  // The link of the logo to the home page, with BankLogo in it.
  home: ReactNode;
  navigation: ReactNode;
  footer: ReactNode;
  children: ReactNode;
}) {
  const { layout } = useBrand();
  const main = useRef<HTMLElement>(null);

  return (
    <div data-density="comfortable" data-layout={layout} className="bank">
      <a
        href="#main-content"
        // The handler moves the focus itself, so that a router does not act on the fragment.
        onClick={(event) => {
          event.preventDefault();
          main.current?.focus();
        }}
        className="bank-skip"
      >
        Skip to content
      </a>
      <header className="bank-header">
        <div className="bank-header-in">
          {home}
          <nav aria-label="Main">
            <ul className="bank-nav">{navigation}</ul>
          </nav>
        </div>
      </header>
      <main id="main-content" ref={main} tabIndex={-1} className="bank-main">
        {children}
      </main>
      {footer}
    </div>
  );
}

export function BankLogo() {
  const { name } = useBrand();
  return <img className="bank-logo" src="/brand/logo.svg" alt={name} />;
}

// WithPhoto shows the hero photo of the brand next to a page, for example the sign-in page.
export function WithPhoto({ children, photo }: { children: ReactNode; photo: ReactNode }) {
  return (
    <div className="with-photo">
      <div className="with-photo-image">{photo}</div>
      <div className="with-photo-page">{children}</div>
    </div>
  );
}
