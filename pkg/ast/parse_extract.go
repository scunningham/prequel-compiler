package ast

import (
	"fmt"
	"regexp"

	"github.com/goccy/go-yaml/ast"
)

var (
	validateExtractName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
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
			extract.Name, err = p.parseExtractName(v.Value)

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

func (p *parserT) parseExtractName(v ast.Node) (string, error) {

	s, err := p.nodeToString(v)
	if err != nil {
		return "", err
	}

	// Ignore strict here; a valid extract name is required for correct operation.
	if !validateExtractName.MatchString(s) {
		err := fmt.Errorf("%w: extract name must start with a letter and contain only letters, numbers, or underscores", ErrBadExtractName)
		return "", p.wrapError(v, err)
	}
	return s, nil

}
