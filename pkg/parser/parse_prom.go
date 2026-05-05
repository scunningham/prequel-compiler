package parser

import (
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

func (p *parserT) parsePromQLNode(state ruleState, node ast.Node) (*AstPromT, error) {

	mapping, err := p.nodeToMapping(node)
	if err != nil {
		return nil, err
	}

	child := state.pushChild(AstNodeTypePromQL)

	var (
		prom = AstPromT{
			baseAst: baseAst{
				ty:      AstNodeTypePromQL,
				scope:   AstScopeCluster,
				address: *child.parent,
				parent:  state.parent,
			},
		}
	)

	for _, v := range mapping.Values {

		key, err := p.nodeToString(v.Key)
		if err != nil {
			return nil, err
		}

		switch key {

		case kwPromExpr:
			prom.Expr, err = p.nodeToString(v.Value)
			// TODO: validate promql expression here

		case kwPromInterval:
			prom.Interval, err = p.nodeToDurationPositive(v.Value)

		case kwPromFor:
			prom.For, err = p.nodeToDurationPositive(v.Value)

		case kwPromEvent:
			prom.Event, err = p.parseEventNode(child, v.Value)

		default:
			if p.strict {
				err = p.wrapError(v, fmt.Errorf("%w: %s", ErrUnexpectedKey, key))
			}
		}

		if err != nil {
			return nil, err
		}
	}

	return &prom, nil
}
