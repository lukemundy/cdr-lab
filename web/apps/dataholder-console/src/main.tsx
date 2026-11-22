import "@cdr-lab/ui/styles.css";

import { ConsoleApp } from "@cdr-lab/app/app";
import { createQueryClient } from "@cdr-lab/app/query";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";

import { createAppRouter } from "./router";

const container = document.querySelector("#root");
if (!container) {
  throw new Error("The page has no #root element.");
}

createRoot(container).render(
  <StrictMode>
    <ConsoleApp router={createAppRouter()} queryClient={createQueryClient()} />
  </StrictMode>,
);
