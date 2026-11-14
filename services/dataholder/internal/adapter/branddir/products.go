package branddir

import (
	"encoding/json"
	"errors"
	"io"
)

type productFile struct {
	Products []json.RawMessage `json:"products"`
}

// ReadProducts reads a file of the format of products.json, which is also the format of a product load file. It gives
// each product as JSON text, and does not validate the products.
func ReadProducts(r io.Reader) ([][]byte, error) {
	var f productFile

	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&f); err != nil {
		return nil, err
	}

	if decoder.More() {
		return nil, errors.New("the file has data after the products")
	}

	products := make([][]byte, 0, len(f.Products))
	for _, p := range f.Products {
		products = append(products, p)
	}

	return products, nil
}
