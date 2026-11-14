package domain

import "errors"

type FieldError struct {
	RecordID string
	// Field has the name of the field in the load file. A dot separates the names of a nested field, for example
	// endpointDetail.publicBaseUri.
	Field string
	// Detail does not name the record, so that a UI can show it at the field.
	Detail string
	// text names the record, for an interface that shows all errors in one message.
	text string
}

func (e FieldError) Error() string { return e.text }

type problemList []error

func (l problemList) Error() string { return errors.Join(l...).Error() }

func (l problemList) Unwrap() []error { return l }

// FieldErrors gives an error that is not a validation error as one other problem.
func FieldErrors(err error) (fields []FieldError, others []error) {
	problems, ok := errors.AsType[problemList](err)
	if !ok {
		if err != nil {
			others = append(others, err)
		}

		return nil, others
	}

	for _, problem := range problems {
		if field, ok := errors.AsType[FieldError](problem); ok {
			fields = append(fields, field)
		} else {
			others = append(others, problem)
		}
	}

	return fields, others
}
