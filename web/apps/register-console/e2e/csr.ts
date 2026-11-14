import { execFileSync } from "node:child_process";
import { createPublicKey, createHash, X509Certificate } from "node:crypto";
import { readFileSync } from "node:fs";

export type CSR = { path: string; text: string; keyPath: string };

// The path is in the output directory of the test, so the repository never gets the key. A name has no slash, which -subj uses.
export function makeCSR(
  path: string,
  subject: { cn: string; o: string },
  dnsNames: string[] = [],
): CSR {
  const keyPath = `${path}.key`;
  execFileSync(
    "openssl",
    [
      "req",
      "-new",
      "-newkey",
      "rsa:2048",
      "-nodes",
      "-keyout",
      keyPath,
      "-out",
      path,
      "-subj",
      `/C=AU/O=${subject.o}/OU=Consumer Data Right/CN=${subject.cn}`,
      ...(dnsNames.length > 0
        ? ["-addext", `subjectAltName=${dnsNames.map((name) => `DNS:${name}`).join(",")}`]
        : []),
    ],
    { stdio: "pipe" },
  );
  return { path, text: readFileSync(path, "utf8"), keyPath };
}

export type IssuedCertificate = { thumbprint: string; publicKey: string; issuer: string };

// The first certificate of a PEM chain, with its RFC 8705 x5t#S256 thumbprint.
export function firstCertificateOf(chain: string): IssuedCertificate {
  const certificate = new X509Certificate(chain);
  return {
    thumbprint: createHash("sha256").update(certificate.raw).digest("base64url"),
    publicKey: certificate.publicKey.export({ type: "spki", format: "pem" }),
    issuer: certificate.issuer,
  };
}

export function publicKeyOf(csr: CSR): string {
  return createPublicKey(readFileSync(csr.keyPath, "utf8")).export({ type: "spki", format: "pem" });
}
