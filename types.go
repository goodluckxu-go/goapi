package goapi

import (
	"encoding"
)

// TextInterface Custom text type interface
type TextInterface interface {
	encoding.TextUnmarshaler
	encoding.TextMarshaler
}
