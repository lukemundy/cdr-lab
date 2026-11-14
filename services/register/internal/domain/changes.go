package domain

import (
	"errors"
	"fmt"
	"slices"
)

var ErrExists = errors.New("already exists")

// IDNotUniqueError is for an ID that a different record has, also a REMOVED record or a record of a different
// aggregate.
type IDNotUniqueError struct{ ID string }

func (e IDNotUniqueError) Error() string {
	return fmt.Sprintf("the ID %q is not unique: a different record has it", e.ID)
}

func firstIDInUse(existing, added []string) error {
	for _, id := range added {
		if slices.Contains(existing, id) {
			return IDNotUniqueError{ID: id}
		}
	}

	return nil
}
