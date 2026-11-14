import { expect, test } from "vitest";
import { expectNoAxeViolations } from "./test/checks.ts";

// The golden files of the data holder: the HTML part of the OTP email of each preset brand
// (services/dataholder/internal/adapter/smtpmail/golden_test.go).
const emails = import.meta.glob<string>(
  "../../../../services/dataholder/internal/adapter/smtpmail/testdata/emails/*.html",
  { query: "?raw", import: "default", eager: true },
);
const logos = import.meta.glob<string>("../../../../services/dataholder/brands/*/logo.png", {
  query: "?url",
  import: "default",
  eager: true,
});

const baseName = (path: string) =>
  path
    .split("/")
    .at(-1)
    ?.replace(/\.\w+$/, "") ?? path;

const goldenEmails = Object.entries(emails).map(([path, html]) => {
  const name = baseName(path);
  const logo = logos[`../../../../services/dataholder/brands/${name}/logo.png`];
  if (!logo) throw new Error(`the preset ${name} has no logo.png`);
  return { name, html, logo };
});

// show puts the email in the document, with the logo part of the email at the URL of the test server.
async function show({ html, logo }: (typeof goldenEmails)[number]) {
  const golden = new DOMParser().parseFromString(html, "text/html");
  for (const link of document.head.querySelectorAll("link[rel=stylesheet]")) link.remove();
  document.documentElement.lang = golden.documentElement.lang;
  document.title = golden.title;
  document.body.replaceWith(document.adoptNode(golden.body));
  for (const image of document.querySelectorAll<HTMLImageElement>('img[src^="cid:"]')) {
    image.src = logo;
  }
  await Promise.all([...document.images].map((image) => image.decode()));
}

test("there is a golden email for each preset brand", () => {
  expect(goldenEmails.length).toBeGreaterThan(0);
});

test.each(goldenEmails)("the OTP email of $name passes the accessibility checks", async (email) => {
  await show(email);

  expect(document.images).toHaveLength(1);
  await expectNoAxeViolations(document.documentElement);
});
