package domain

import "strings"

// maxCertificateNameLength is the upper bound of a CN and of an O in RFC 5280, Appendix A.
const maxCertificateNameLength = 64

type CSRRulesError struct{ Rules []string }

func (e CSRRulesError) Error() string {
	return "The CSR breaks these rules: " + strings.Join(e.Rules, "; ") + "."
}

func (CSRRulesError) Is(target error) bool { return target == ErrNotValid }
