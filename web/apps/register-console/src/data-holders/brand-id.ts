import type { DataHolderBrand } from "../api/client";

// services/register/design/operator-console-api.md: a data holder brand is in a path with its dataHolderBrandId, or
// with its interimId when it has no dataHolderBrandId.
export function brandIdOf(brand: DataHolderBrand): string {
  return brand.dataHolderBrandId ?? brand.interimId ?? "";
}
