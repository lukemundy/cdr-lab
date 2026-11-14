import fontList from "../fonts.json" with { type: "json" };

export interface Font {
  package: string;
  family: string;
}

// The fixed font set (services/dataholder/internal/adapter/branddir/tokens.go has the same families).
export const fonts: Font[] = fontList;
