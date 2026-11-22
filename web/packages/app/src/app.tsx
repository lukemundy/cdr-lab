import type { QueryClient } from "@tanstack/react-query";
import { ToastProvider } from "@cdr-lab/ui/toast";
import { QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider, type AnyRouter } from "@tanstack/react-router";

export type ConsoleAppProps = { router: AnyRouter; queryClient: QueryClient };

export function ConsoleApp({ router, queryClient }: ConsoleAppProps) {
  return (
    <QueryClientProvider client={queryClient}>
      <ToastProvider>
        <RouterProvider router={router} />
      </ToastProvider>
    </QueryClientProvider>
  );
}
