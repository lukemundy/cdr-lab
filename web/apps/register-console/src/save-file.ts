// The save makes no request, so it needs no change to the CSP: no directive applies to a download from a link.
export function saveFile(name: string, text: string) {
  const url = URL.createObjectURL(new Blob([text], { type: "application/pem-certificate-chain" }));
  const link = document.createElement("a");
  link.href = url;
  link.download = name;
  link.click();
  // A revocation before the download starts can stop the download in some browsers.
  setTimeout(() => URL.revokeObjectURL(url));
}
