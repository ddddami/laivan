package server

import (
	"bytes"
	"encoding/json"
	"errors"
)

// patchField tracks whether a field was present in a JSON payload.
// It explicitly rejects explicit null values.
type patchField[T any] struct {
	Value   T
	Present bool
}

func (p *patchField[T]) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return errors.New("patch fields must not be null")
	}

	var value T
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	p.Value = value
	p.Present = true
	return nil
}

// patchValue converts a patch field to a pointer, returning nil if it was absent.
func patchValue[T any](field patchField[T]) *T {
	if !field.Present {
		return nil
	}
	return &field.Value
}
