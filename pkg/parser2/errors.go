package parser2

import (
	"errors"
	"fmt"
)

var (
	ErrMissingKey      = errors.New("missing key")
	ErrUnexpectedKey   = errors.New("unexpected key")
	ErrUnexpectedType  = errors.New("unexpected type")
	ErrUndefinedAnchor = errors.New("undefined anchor")
	ErrNegateCount     = errors.New("negate fields cannot have count > 1")
	ErrZeroCount       = errors.New("count value must be a positive integer")
	ErrBadIdentifier   = errors.New("identifier is not valid")
	ErrBadHash         = errors.New("hash is not valid")
	ErrBadGen          = errors.New("gen is not valid")
	ErrBadKind         = errors.New("kind is not valid")
	ErrBadSeverity     = errors.New("severity is not valid")
	ErrOverflow        = errors.New("value overflow")
	ErrMultipleOrigin  = errors.New("multiple origin events are not allowed")
	ErrUnknownNodeType = errors.New("unknown node type")
)

type ErrRule struct {
	Meta AstMetadataT
	Err  error
}

func (e ErrRule) Error() string {
	return fmt.Sprintf("rule error id=%s hash=%s: %v", e.Meta.Id, e.Meta.Hash, e.Err)
}

func (e ErrRule) Unwrap() error {
	return e.Err
}

type ParseError struct {
	Line   int
	Column int
	Offset int
	Msg    string
}

func (e ParseError) Error() string {
	return e.Msg
}
