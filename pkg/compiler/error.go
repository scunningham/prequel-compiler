package compiler

import "errors"

var (
	ErrNoFields            = errors.New("no fields")
	ErrUnsupportedNodeType = errors.New("unsupported node type")
	ErrUnsupportedAstType  = errors.New("unsupported AST node type")
	ErrSequenceSingleMatch = errors.New("sequence with single match (use set instead)")
	ErrUnsupportedScope    = errors.New("unsupported scope")
	ErrObjectTypeAssertion = errors.New("object type assertion failed")
)
