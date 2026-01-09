package goapi

import (
	"encoding"
)

// TextInterface Custom text type interface
// Pointer inheritance must be defined for interface 'encoding.TextUnmarshaler'
// It is necessary to inherit both interfaces 'encoding.TextMarshaler' and 'encoding.TextUnmarshaler' simultaneously
// When in use, determine whether the pointer type of the original type inherits from 'encoding.TextUnmarshaler' and
// whether the original type inherits from 'encoding.TextMarshaler'
type TextInterface interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}
