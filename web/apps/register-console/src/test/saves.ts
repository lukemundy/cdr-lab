import { vi } from "vitest";

export type Saved = { name: string; text: Promise<string> };

// The spy stops the click, so that the browser of the test saves no file. The test restores the mocks.
export function captureSaves() {
  const saves: Saved[] = [];
  vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(function (
    this: HTMLAnchorElement,
  ) {
    saves.push({ name: this.download, text: fetch(this.href).then((r) => r.text()) });
  });
  return saves;
}
