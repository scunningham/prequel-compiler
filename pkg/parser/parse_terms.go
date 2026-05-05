package parser

import (
	"fmt"
	"math"
	"regexp"

	"github.com/goccy/go-yaml/ast"
)

// Terms node expects a sequence of terms,

func (p *parserT) parseTerms(state ruleState, v ast.Node, allowNegate bool) ([]*protoTerm, error) {

	seq, err := p.nodeToSequence(v)
	if err != nil {
		return nil, err
	}

	var (
		allLeaves bool
		terms     []*protoTerm
	)

	for i, termNode := range seq.Values {

		term, err := p.parseTerm(state, termNode, allowNegate)

		switch {
		case err != nil:
			return nil, err

		case i == 0:
			// First term; determine if this is a leaf term or an inner node term, and set the allLeaves flag accordingly.
			allLeaves = term.leaf != nil

		case allLeaves && term.leaf == nil:
			err := fmt.Errorf("%w: all terms must be leaves", ErrUnexpectedType)
			return nil, p.wrapError(termNode, err)

		case !allLeaves && term.leaf != nil:
			err := fmt.Errorf("%w: all terms must be inner nodes", ErrUnexpectedType)
			return nil, p.wrapError(termNode, err)

		}

		terms = append(terms, term)

		state = state.incRank()
	}

	return terms, nil
}

// A term which can be either a simple string or a mapping.

func (p *parserT) parseTerm(state ruleState, node ast.Node, allowNegate bool) (*protoTerm, error) {

	if node.Type() != ast.StringType {
		return p.parseTermAsMap(state, node, allowNegate)
	}

	v, ok := node.(*ast.StringNode)
	if !ok {
		err := fmt.Errorf("%w: expected term to be a string, got %s", ErrUnexpectedType, node.Type())
		return nil, p.wrapError(node, err)
	}

	return &protoTerm{
		leaf: &protoField{
			StrValue: v.Value,
		},
	}, nil
}

// This is where it gets interesting.
//
//	type ParseTermT struct {
// 		// LineMatch
//		Field      string            `yaml:"field,omitempty"`
//		StrValue   string            `yaml:"value,omitempty"`
//		JqValue    string            `yaml:"jq,omitempty"`
//		RegexValue string            `yaml:"regex,omitempty"`
//		Count      int               `yaml:"count,omitempty"`
//		Extract    []ParseExtractT   `yaml:"extract,omitempty"`
//
//		// Recursive terms
//		Set        *ParseSetT        `yaml:"set,omitempty"`
//		Sequence   *ParseSequenceT   `yaml:"sequence,omitempty"`

// 		// Applies to linematch and seq/seq
//		NegateOpts *ParseNegateOptsT `yaml:",inline,omitempty"`
//
// 		// Other types of terms
//		PromQL     *ParsePromQL      `yaml:"promql,omitempty"`
//		Script     *ParseScriptT     `yaml:"script,omitempty"`
//	}
//
//	A term is either a line match, or a recursive struct.

func (p *parserT) parseTermAsMap(state ruleState, node ast.Node, allowNegate bool) (*protoTerm, error) {

	mapping, err := p.nodeToMapping(node)
	if err != nil {
		return nil, err
	}

	var (
		child AstNode
		leaf  *protoField
		nOpts *AstNegateOptsT
	)

	for _, v := range mapping.Values {

		key, ok := v.Key.(*ast.StringNode)
		if !ok {
			err := fmt.Errorf("%w: %s", ErrUnexpectedType, v.Key.Type())
			return nil, p.wrapError(v.Key, err)
		}

		switch key.Value {

		case kwField, kwValue, kwJq, kwRegex, kwCount, kwExtract:
			if child != nil {
				err := fmt.Errorf("%w: multiple term keys found in term definition", ErrUnexpectedKey)
				return nil, p.wrapError(v.Key, err)
			}
			if leaf == nil {
				leaf = &protoField{Count: 1}
			}
			if err := p.parseTermField(key, v.Value, leaf, allowNegate); err != nil {
				return nil, err
			}

		case kwSet:
			if leaf != nil || child != nil {
				err := fmt.Errorf("%w: multiple term keys found in term definition", ErrUnexpectedKey)
				return nil, p.wrapError(v.Key, err)
			}
			if child, err = p.parseInnerNode(state, AstNodeTypeSet, v.Value); err != nil {
				return nil, err
			}

		case kwSequence:
			if leaf != nil || child != nil {
				err := fmt.Errorf("%w: multiple term keys found in term definition", ErrUnexpectedKey)
				return nil, p.wrapError(v.Key, err)
			}
			if child, err = p.parseInnerNode(state, AstNodeTypeSeq, v.Value); err != nil {
				return nil, err
			}

		case kwWindow, kwSlide, kwAnchor, kwAbsolute:
			if !allowNegate {
				err := fmt.Errorf("%w: negate options are not allowed in this context", ErrUnexpectedKey)
				return nil, p.wrapError(key, err)
			}
			if nOpts == nil {
				nOpts = &AstNegateOptsT{}
			}
			if err = p.parseNegateOpts(key.Value, v.Value, nOpts); err != nil {
				return nil, err
			}

		case kwPromQL:
			if leaf != nil || child != nil {
				err := fmt.Errorf("%w: multiple term keys found in term definition", ErrUnexpectedKey)
				return nil, p.wrapError(v.Key, err)
			}
			if nOpts != nil {
				err := fmt.Errorf("%w: negate options cannot be used with %s terms", ErrUnexpectedKey, key.Value)
				return nil, p.wrapError(key, err)
			}
			if child, err = p.parsePromQLNode(state, v.Value); err != nil {
				return nil, err
			}

		case kwScript:
			if leaf != nil || child != nil {
				err := fmt.Errorf("%w: multiple term keys found in term definition", ErrUnexpectedKey)
				return nil, p.wrapError(v.Key, err)
			}
			if nOpts != nil {
				err := fmt.Errorf("%w: negate options cannot be used with %s terms", ErrUnexpectedKey, key.Value)
				return nil, p.wrapError(key, err)
			}
			if child, err = p.parseScriptNode(state, v.Value); err != nil {
				return nil, err
			}

		default:
			err := fmt.Errorf("%w: unexpected key '%s' in term definition", ErrUnexpectedKey, key.Value)
			return nil, p.wrapError(v.Key, err)
		}
	}

	return &protoTerm{
		negateOpts: nOpts,
		child:      child,
		leaf:       leaf,
	}, nil
}

func (p *parserT) parseTermField(key *ast.StringNode, v ast.Node, match *protoField, allowNegate bool) error {

	var (
		err      error
		hasValue bool
	)

	if match.RegexValue != "" || match.JqValue != "" || match.StrValue != "" {
		hasValue = true
	}

	mkMatchKeyError := func() error {
		kerr := fmt.Errorf("%w: '%s' key cannot be used together with other line match keys", ErrUnexpectedKey, key.Value)
		return p.wrapError(key, kerr)
	}

	switch key.Value {
	case kwField:
		match.Field, err = p.nodeToString(v)

	case kwCount:
		match.Count, err = p.nodeToUint64(v)
		switch {
		case err != nil:
			// fall through
		case match.Count == 0:
			err = ErrZeroCount
		case allowNegate && match.Count > 1:
			err = ErrNegateCount
		}

	case kwExtract:
		match.Extract, err = p.parseExtracts(v)

	case kwJq:
		if hasValue {
			err = mkMatchKeyError()
		} else {
			match.JqValue, err = p.nodeToString(v)
		}
		// TODO: validate jq if hook available.

	case kwRegex:
		if hasValue {
			err = mkMatchKeyError()
		} else {
			var exp *regexp.Regexp
			if exp, err = p.nodeToRegex(v); err == nil {
				match.RegexValue = exp.String()
			}
		}

	case kwValue:
		if hasValue {
			err = mkMatchKeyError()
		} else {
			match.StrValue, err = p.nodeToString(v)
		}

	default:
		kerr := fmt.Errorf("%w: expected line matching key, got %s", ErrUnexpectedKey, key)
		err = p.wrapError(key, kerr)
	}

	return err
}

func (p *parserT) parseNegateOpts(key string, node ast.Node, opts *AstNegateOptsT) error {
	var err error

	switch key {
	case kwWindow:
		opts.Window, err = p.nodeToDurationPositive(node)

	case kwSlide:
		opts.Slide, err = p.nodeToDuration(node)

	case kwAnchor:
		var anchorInt uint64
		anchorInt, err = p.nodeToUint64(node)
		switch {
		case err != nil:
			// fall through
		case anchorInt > math.MaxUint32:
			nerr := fmt.Errorf("%w: anchor value must be a non-negative integer, got %d", ErrUnexpectedType, anchorInt)
			err = p.wrapError(node, nerr)

		default:
			opts.Anchor = uint32(anchorInt)
		}

	case kwAbsolute:
		if absNode, ok := node.(*ast.BoolNode); !ok {
			err = fmt.Errorf("%w: expected absolute value to be a boolean, got %s", ErrUnexpectedType, node.Type())
			err = p.wrapError(node, err)
		} else {
			opts.Absolute = absNode.Value
		}

	default:
		// Should not happen; parseTermAsMap should only call this on expected keys.
		err = fmt.Errorf("%w: %s", ErrUnexpectedKey, key)
	}

	if err != nil {
		ferr := fmt.Errorf("failed to parse negate option '%s': %w", key, err)
		return p.wrapError(node, ferr)
	}
	return nil
}
