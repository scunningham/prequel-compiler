package parser2

import (
	"errors"
)

var (
	ErrMissingKey      = errors.New("missing key")
	ErrUnexpectedKey   = errors.New("unexpected key")
	ErrUnexpectedType  = errors.New("unexpected type")
	ErrUndefinedAnchor = errors.New("undefined anchor")
)
