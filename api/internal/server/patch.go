package server

import (
	"bytes"
	"encoding/json"
	"errors"
)

// PatchField is a generic type for tracking whether a field was present in a JSON payload.
// It explicitly rejects explicit null values.
type PatchField[T any] struct {
	Value   T
	Present bool
}

func (p *PatchField[T]) UnmarshalJSON(data []byte) error {
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

// OptionalValue converts a PatchField to a pointer, returning nil if the field was not present.
func OptionalValue[T any](field PatchField[T]) *T {
	if !field.Present {
		return nil
	}
	return &field.Value
}
