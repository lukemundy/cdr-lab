//go:build openid

package openid

import "testing"

// acceptedFindings are the conditions of the suite whose failure or warning the data holder accepts, by the name of
// the condition (src in the log). Each has a judgement in services/dataholder/design/.
var acceptedFindings = map[string]string{
	// The suite probes TLS 1.2 on the resource endpoint with no client certificate, and the mTLS listener ends the
	// handshake when the client has no certificate.
	"DisallowInsecureCipher":            "api-standards.md: DH-0497 The TLS findings of the OpenID conformance suite",
	"CheckForBCP195InsecureFAPICiphers": "api-standards.md: DH-0497 The TLS findings of the OpenID conformance suite",
}

// onlyAcceptedFindings logs each finding, and gives true when each finding is accepted.
func onlyAcceptedFindings(t *testing.T, findings []finding) bool {
	t.Helper()

	accepted := len(findings) > 0

	for _, f := range findings {
		reason, ok := acceptedFindings[f.Src]
		if !ok {
			accepted = false
			reason = "not accepted"
		}

		t.Logf("    %s %s: %.200s (%s)", f.Result, f.Src, f.Msg, reason)
	}

	return accepted
}
