package domain

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// FieldError has no value of the record, because the text can go to telemetry.
type FieldError struct {
	// A dot separates the names of a nested field, for example extendedData.payee.
	Field string
	// Detail names the field, so that it is also correct in a message with all the errors.
	Detail string
}

type ValidationError struct {
	Errors []FieldError
}

func (e ValidationError) Error() string {
	details := make([]string, 0, len(e.Errors))
	for _, f := range e.Errors {
		details = append(details, f.Detail)
	}

	return strings.Join(details, "; ")
}

// fieldProblems collects all the errors of a record, so that the operator can correct each value in one pass.
type fieldProblems struct {
	errs []FieldError
}

func (p *fieldProblems) check(ok bool, field, format string, args ...any) {
	if !ok {
		p.errs = append(p.errs, FieldError{Field: field, Detail: fmt.Sprintf(format, args...)})
	}
}

func (p *fieldProblems) name(field, value string) {
	length := utf8.RuneCountInString(value)
	p.check(
		length >= 1 && length <= MaxNameCharacters,
		field, "%s must have 1 to %d characters", field, MaxNameCharacters,
	)
}

func (p *fieldProblems) validationError() error {
	if len(p.errs) == 0 {
		return nil
	}

	return ValidationError{Errors: p.errs}
}
