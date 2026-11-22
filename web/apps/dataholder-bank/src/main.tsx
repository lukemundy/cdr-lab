import "@cdr-lab/ui/consumer.css";
import "./site.css";

import { AppShell } from "@cdr-lab/web-platform/shell";
import { createQueryClient } from "@cdr-lab/web-platform/query";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import { createAppRouter } from "./router";

const container = document.querySelector("#root");
if (!container) {
  throw new Error("The page has no #root element.");
}

createRoot(container).render(
  <StrictMode>
    <AppShell router={createAppRouter()} queryClient={createQueryClient()} />
  </StrictMode>,
);
