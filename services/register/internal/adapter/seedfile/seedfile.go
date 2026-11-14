// Package seedfile reads a full set of participants from a JSON file.
package seedfile

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/lukemundy/cdr-lab/services/register/internal/domain"
)

func Read(r io.Reader) (domain.Participants, error) {
	var p domain.Participants

	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&p); err != nil {
		return domain.Participants{}, err
	}

	if decoder.More() {
		return domain.Participants{}, errors.New("the file has data after the participants")
	}

	return p, nil
}
