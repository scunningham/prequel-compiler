package ast

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/token"
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
	ErrMissingOrigin   = errors.New("missing origin")
	ErrUnknownNodeType = errors.New("unknown node type")
	ErrBadExtractName  = errors.New("extract name is not valid")
	ErrBadScriptLang   = errors.New("script language is not valid")
	ErrBadScriptCode   = errors.New("script code is not valid")
	ErrBadAnchor       = errors.New("anchor value is out of range")
	ErrMissingTerm     = errors.New("at least one term is required")
	ErrShortSequence   = errors.New("sequence must have at least 2 terms")
	ErrMissingEvent    = errors.New("event is required when using leaf terms")
	ErrMissingWindow   = errors.New("window is required when using multiple terms")
	ErrWindowNegative  = errors.New("window duration cannot be negative")
	ErrMissingSource   = errors.New("source is required in event")
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
	token *token.Token
	err   error
}

func (e ParseError) Offset() int {
	if e.token == nil {
		return 0
	}
	return e.token.Position.Offset
}

func (e ParseError) Line() int {
	if e.token == nil {
		return 0
	}
	return e.token.Position.Line
}

func (e ParseError) Column() int {
	if e.token == nil {
		return 0
	}
	return e.token.Position.Column
}

func (e ParseError) Unwrap() error {
	return e.err
}

func (e ParseError) Error() string {
	return e.err.Error()
}

func (e ParseError) Format(colored, inclSource bool) string {
	return yaml.FormatErrorWithToken(e.err.Error(), e.token, colored, inclSource)
}
