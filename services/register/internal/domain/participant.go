package domain

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"unicode/utf8"
)

// The maxLength values of the CDS Register schemas.
const (
	MaxIDLength   = 36
	MaxNameLength = 200
	// The LegalEntityDetail schema of Get Data Holder Brands gives a shorter limit than the other name fields.
	MaxDataHolderLegalEntityNameLength = 36
	MaxAccreditationNumberLength       = 100
	MaxURILength                       = 1000
	MaxDescriptionLength               = 4000
)

// Status is the status of a brand or a software product.
type Status string

const (
	Active   Status = "ACTIVE"
	Inactive Status = "INACTIVE"
	Removed  Status = "REMOVED"
)

var statuses = []Status{Active, Inactive, Removed}

type Industry string

const (
	Banking        Industry = "banking"
	Energy         Industry = "energy"
	NonBankLending Industry = "non-bank-lending"
	Telco          Industry = "telco"
)

var industries = []Industry{Banking, Energy, NonBankLending, Telco}

func ParseIndustry(s string) (Industry, bool) {
	i := Industry(s)
	return i, slices.Contains(industries, i)
}

type record struct {
	kind string
	id   string
}

func (r record) String() string {
	return fmt.Sprintf("%s %q", r.kind, r.id)
}

// problems collects all validation errors, so that the operator can correct a file in one pass.
type problems struct {
	errs []error
	// rules counts the errors in errs that are broken status rules.
	rules int
	ids   map[string]bool
}

func (p *problems) add(r record, field, format string, args ...any) {
	detail := fmt.Sprintf(format, args...)
	p.errs = append(p.errs, FieldError{
		RecordID: r.id, Field: field, Detail: detail, text: fmt.Sprintf("%s: %s", r, detail),
	})
}

// The format starts with the verb, for example "is %s", because the text puts the record or the field before it.
func (p *problems) statusRuleBroken(r record, format string, args ...any) {
	rule := fmt.Sprintf(format, args...)
	p.errs = append(p.errs, FieldError{
		RecordID: r.id, Field: "status", Detail: "status " + rule, text: fmt.Sprintf("%s %s", r, rule),
	})
	p.rules++
}

func (p *problems) required(r record, field, value string) {
	if value == "" {
		p.add(r, field, "%s is required", field)
	}
}

// maxLength counts characters, not bytes, as in the CDS Register schemas.
func (p *problems) requiredText(r record, field, value string, maxLength int) {
	p.required(r, field, value)

	if utf8.RuneCountInString(value) > maxLength {
		p.add(r, field, "%s is longer than %d characters", field, maxLength)
	}
}

func (p *problems) requiredList(r record, field string, values []string) {
	if len(values) == 0 {
		p.add(r, field, "%s is required", field)
	}
}

func (p *problems) requiredURI(r record, field, value string) {
	p.required(r, field, value)
	p.optionalURI(r, field, value)
}

func (p *problems) optionalURI(r record, field, value string) {
	if problem := uriProblem(value); problem != "" {
		p.add(r, field, "%s %s", field, problem)
	}
}

func (p *problems) requiredURIList(r record, field string, values []string) {
	p.requiredList(r, field, values)

	for _, value := range values {
		if problem := uriProblem(value); problem != "" {
			p.add(r, field, "%s has a value that %s", field, problem)
		}
	}
}

// uriProblem gives no problem for an empty value, so that the caller decides if the field is required.
func uriProblem(value string) string {
	switch {
	case value == "":
		return ""
	case utf8.RuneCountInString(value) > MaxURILength:
		return fmt.Sprintf("is longer than %d characters", MaxURILength)
	case !isHTTPSURI(value):
		return "is not an https URI with a host"
	default:
		return ""
	}
}

// uriCharacters are the characters of RFC 3986, section 2. url.Parse also accepts some characters that are not in
// this set, for example a space in the path.
const uriCharacters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-._~:/?#[]@!$&'()*+,;=%"

func isHTTPSURI(value string) bool {
	if strings.ContainsFunc(value, func(c rune) bool { return !strings.ContainsRune(uriCharacters, c) }) {
		return false
	}

	u, err := url.Parse(value)

	return err == nil && u.Scheme == "https" && u.Host != ""
}

func (p *problems) id(r record, field, id string) {
	switch {
	case id == "":
		p.add(r, field, "%s is required", field)
	case len(id) > MaxIDLength:
		p.add(r, field, "%s is longer than %d characters", field, MaxIDLength)
	case p.ids[id]:
		p.add(r, field, "the ID is not unique")
	}

	if p.ids == nil {
		p.ids = map[string]bool{}
	}

	p.ids[id] = true
}

func (p *problems) status(r record, s Status) bool {
	if !slices.Contains(statuses, s) {
		p.add(r, "status", "status %q is not ACTIVE, INACTIVE or REMOVED", s)
		return false
	}

	return true
}

// err gives a StatusRuleError only when each problem is a broken status rule: the records are then valid alone.
func (p *problems) err() error {
	switch {
	case len(p.errs) == 0:
		return nil
	case p.rules == len(p.errs):
		return StatusRuleError{problemList(p.errs)}
	default:
		return notValidError{problemList(p.errs)}
	}
}
