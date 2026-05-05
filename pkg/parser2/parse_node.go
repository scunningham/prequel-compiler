package parser2

import (
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

// type ParseSetT struct {
// 	Window       string       `yaml:"window,omitempty"`
// 	Correlations []string     `yaml:"correlations,omitempty"`
// 	Event        *ParseEventT `yaml:"event,omitempty"`
// 	Match        []ParseTermT `yaml:"match,omitempty"`
// 	Negate       []ParseTermT `yaml:"negate,omitempty"`
// }

func (p *parserT) parseInnerNode(state ruleState, ty AstNodeType, node ast.Node) (AstNode, error) {

	mapping, err := p.nodeToMapping(node)
	if err != nil {
		return nil, err
	}

	// Sanity checks
	switch {
	case state.parent != nil && state.parent.Depth > maxDepth:
		err := fmt.Errorf("%w: maximum depth exceeded: %d/%d", ErrUnexpectedType, state.parent.Depth, maxDepth)
		return nil, p.wrapError(node, err)

	case state.rank > maxRank:
		err := fmt.Errorf("%w: maximum rank exceeded: %d/%d", ErrUnexpectedType, state.rank, maxRank)
		return nil, p.wrapError(node, err)
	}

	var (
		proto      = protoNode{ty: ty}
		child      = state.pushChild(ty)
		negateNode ast.Node
	)

	for _, v := range mapping.Values {

		key, err := p.nodeToString(v.Key)
		if err != nil {
			return nil, err
		}
		switch key {

		case kwEvent:
			if proto.event, err = p.parseEventNode(child, v.Value); err != nil {
				return nil, err
			}

		case kwWindow:
			if proto.window, err = p.nodeToDurationPositive(v.Value); err != nil {
				return nil, err
			}

		case kwMatch:
			if proto.ty != AstNodeTypeSet {
				err := fmt.Errorf("%w: '%s' key is not allowed in this context", ErrUnexpectedKey, key)
				return nil, p.wrapError(v.Key, err)
			}
			if proto.terms, err = p.parseTerms(child, v.Value, false); err != nil {
				return nil, err
			}

		case kwOrder:
			if proto.ty != AstNodeTypeSeq {
				err := fmt.Errorf("%w: '%s' key is not allowed in this context", ErrUnexpectedKey, key)
				return nil, p.wrapError(v.Key, err)
			}
			if proto.terms, err = p.parseTerms(child, v.Value, false); err != nil {
				return nil, err
			}

		case kwNegate:
			// Defer parsing the negate node until the end in case it occurrs before the matc/order node.
			// This is necessary to get the addressing consistent; ie. negative terms successively addressed after
			// normal terms.
			negateNode = v.Value

		case kwCorrelations:
			if proto.correlations, err = p.nodeToStrs(v.Value); err != nil {
				return nil, err
			}

		default:
			err := fmt.Errorf("%w: %s", ErrUnexpectedKey, v.Key)
			return nil, p.wrapError(v, err)
		}
	}

	if negateNode != nil {
		// Fix up the rank on the state to include the already parsed match/order terms,
		// so that negate terms are ranked after them.
		negateState := child.setRank(uint32(len(proto.terms)))
		if proto.negate, err = p.parseTerms(negateState, negateNode, true); err != nil {
			return nil, err
		}
	}

	return p.constructNode(child.parent, state, mapping, proto)
}

// Interpret the prototype and construct the appropriate AST node (SetNode, SequenceNode, etc).

func (p *parserT) constructNode(addr *AstNodeAddressT, state ruleState, mapping *ast.MappingNode, proto protoNode) (AstNode, error) {

	// Sanity check; at least something has to match.
	if len(proto.terms) == 0 {
		err := fmt.Errorf("%w: a %s node must contain at least one term", ErrMissingKey, proto.ty.String())
		return nil, p.wrapError(mapping, err)
	}

	// If there are negate terms, they must either all be leaf terms or all inner node terms,
	// and they must match the type of the match/order terms.
	allLeaves := proto.terms[0].leaf != nil

	if len(proto.negate) > 0 {
		if negateAllLeaves := proto.negate[0].leaf != nil; allLeaves != negateAllLeaves {
			err := fmt.Errorf("%w: match terms and negate terms must both be either leaves or inner nodes", ErrUnexpectedType)
			return nil, p.wrapError(mapping, err)
		}
	}

	// Confirm that the event key is set if required, and not set otherwise.
	switch {
	case allLeaves && proto.event == nil:
		err := fmt.Errorf("%w: an event is required when using leaf terms", ErrMissingKey)
		return nil, p.wrapError(mapping, err)

	case !allLeaves && proto.event != nil:
		err := fmt.Errorf("%w: an event is not allowed when using inner node terms", ErrUnexpectedKey)
		return nil, p.wrapError(mapping, err)
	}

	// TODO: Validate anchors is negate terms; should be in range of [1, len(terms)-1]

	var node AstNode

	if allLeaves {
		if proto.event == nil {
			err := fmt.Errorf("%w: an event is required when using leaf terms", ErrMissingKey)
			return nil, p.wrapError(mapping, err)
		}

		node = &AstMatchLeafT{
			baseAst:      baseAst{ty: proto.ty, address: *addr, parent: state.parent, scope: AstScopeNode},
			Window:       proto.window,
			Correlations: proto.correlations,
			Terms:        protoTermsToAstFields(proto.terms),
			Negate:       protoTermsToAstFields(proto.negate),
			Event:        *proto.event,
		}
	} else {

		node = &AstInnerNodeT{
			baseAst:      baseAst{ty: proto.ty, address: *addr, parent: state.parent, scope: AstScopeCluster},
			Window:       proto.window,
			Correlations: proto.correlations,
			Terms:        protoTermsToAstTerms(proto.terms),
			Negate:       protoTermsToAstTerms(proto.negate),
		}
	}

	return node, nil
}
