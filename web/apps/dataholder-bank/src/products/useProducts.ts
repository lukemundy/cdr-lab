import { useQuery } from "@tanstack/react-query";

import { listProducts } from "../api/client";

export function useProducts() {
  return useQuery({ queryKey: ["products"], queryFn: ({ signal }) => listProducts(signal) });
}
