import { useBrand } from "./brand";

// With no such photo in the brand, a shape in the brand colour is in its place, so the site never asks for a missing file.
export function BrandPhoto({
  file,
  className,
  eager = false,
}: {
  file: string;
  className?: string;
  eager?: boolean;
}) {
  const { photos } = useBrand();

  if (!photos.includes(file)) {
    return (
      <span data-photo-placeholder className={`photo photo-placeholder ${className ?? ""}`}>
        <img src="/brand/mark.svg" alt="" />
      </span>
    );
  }

  return (
    <img
      className={`photo ${className ?? ""}`}
      src={`/brand/${file}`}
      alt=""
      loading={eager ? "eager" : "lazy"}
      decoding="async"
    />
  );
}
