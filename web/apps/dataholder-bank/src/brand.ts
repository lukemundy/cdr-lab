import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";

import { getBrand } from "./api/client";

// The pages say "Bank" until the data holder gives the brand name, and when it cannot
// (services/dataholder/design/online-bank-site.md, Presentation).
const fallbackName = "Bank";
const noPhotos: readonly string[] = [];

export function useBrand() {
  const brand = useQuery({
    queryKey: ["brand"],
    queryFn: ({ signal }) => getBrand(signal),
    staleTime: Infinity,
  });
  return {
    name: brand.data?.brandName ?? fallbackName,
    demonstrationNotice: brand.data?.demonstrationNotice,
    // Until the brand loads, the page has no layout and no photos, and shows a shape in place of each photo.
    layout: brand.data?.layout,
    photos: brand.data?.photos ?? noPhotos,
  };
}

export function usePageTitle(page: string) {
  const { name } = useBrand();
  useEffect(() => {
    document.title = `${page} - ${name} online banking`;
  }, [page, name]);
}
