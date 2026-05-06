package ast

import (
	"fmt"
	"regexp"

	"github.com/goccy/go-yaml/ast"
)

func (p *parserT) parseExtracts(node ast.Node) ([]AstExtractT, error) {

	seq, err := p.nodeToSequence(node)
	if err != nil {
		return nil, err
	}

	var (
		extracts []AstExtractT
	)

	for _, v := range seq.Values {

		extract, err := p.parseExtractNode(v)
		if err != nil {
			return nil, err
		}

		extracts = append(extracts, *extract)
	}

	return extracts, nil
}

func (p *parserT) parseExtractNode(node ast.Node) (*AstExtractT, error) {

	mapping, err := p.nodeToMapping(node)
	if err != nil {
		return nil, err
	}

	var (
		hasValue bool
		extract  AstExtractT
	)

	for _, v := range mapping.Values {

		key, err := p.nodeToString(v.Key)
		if err != nil {
			return nil, err
		}

		switch key {
		case kwExtractName:
			extract.Name, err = p.nodeToString(v.Value)

		case kwExtractJq:
			if hasValue {
				err = fmt.Errorf("%w: multiple value keys in extract: %s", ErrUnexpectedKey, key)
				return nil, p.wrapError(v, err)
			}
			hasValue = true
			extract.JqValue, err = p.nodeToString(v.Value)

		case kwExtractRegex:
			if hasValue {
				err = fmt.Errorf("%w: multiple value keys in extract: %s", ErrUnexpectedKey, key)
				return nil, p.wrapError(v, err)
			}
			hasValue = true
			var exp *regexp.Regexp
			if exp, err = p.nodeToRegex(v.Value); err == nil {
				extract.RegexValue = exp.String()
			}

		default:
			if p.strict {
				err = p.wrapError(v, fmt.Errorf("%w: unexpected key in extract: %s", ErrUnexpectedKey, key))
			}
		}

		if err != nil {
			return nil, err
		}
	}

	return &extract, nil
}
