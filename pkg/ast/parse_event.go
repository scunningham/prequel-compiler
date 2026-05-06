package ast

import (
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

func (p *parserT) parseEventNode(state ruleState, node ast.Node) (*AstEventT, error) {

	mapping, err := p.nodeToMapping(node)
	if err != nil {
		return nil, err
	}

	var (
		event AstEventT
	)

	for _, v := range mapping.Values {
		key, err := p.nodeToString(v.Key)
		if err != nil {
			return nil, err
		}

		switch key {
		case kwOrigin:
			event.Origin, err = p.parseOrigin(state, v.Value)
		case kwSource:
			event.Source, err = p.nodeToString(v.Value)
		default:
			if p.strict {
				err = p.wrapError(v, fmt.Errorf("%w: unexpected key in event: %s", ErrUnexpectedKey, key))
			}
		}

		if err != nil {
			return nil, err
		}
	}

	return &event, nil
}

func (p *parserT) parseOrigin(state ruleState, node ast.Node) (bool, error) {
	b, err := p.nodeToBool(node)
	if err != nil {
		return false, err
	}

	if b {
		*state.origin++
		if *state.origin > 1 {
			return false, p.wrapError(node, ErrMultipleOrigin)
		}
	}

	return b, nil
}
