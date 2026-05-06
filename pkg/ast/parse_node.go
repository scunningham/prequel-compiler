package ast

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
	case state.addr != nil && state.addr.Depth > maxDepth:
		err := fmt.Errorf("%w: maximum depth exceeded: %d/%d", ErrUnexpectedType, state.addr.Depth, maxDepth)
		return nil, p.wrapError(node, err)

	case state.rank > maxRank:
		err := fmt.Errorf("%w: maximum rank exceeded: %d/%d", ErrUnexpectedType, state.rank, maxRank)
		return nil, p.wrapError(node, err)
	}

	var (
		proto      = protoNode{ty: ty}
		child      = state.pushNode(ty).setRank(0) // Reset the rank for the child node; the parent rank should not affect the rank of terms within a set or sequence.
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

	return p.constructNode(state, child, mapping, proto)
}

// Interpret the prototype and construct the appropriate AST node (SetNode, SequenceNode, etc).

func (p *parserT) constructNode(parent, child ruleState, mapping *ast.MappingNode, proto protoNode) (AstNode, error) {

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

	switch {
	case !allLeaves:
		node = p.constructInnerNode(parent, child, proto)
	default:
		node = p.constructLeafNode(parent, child, proto)
	}

	return node, nil
}

func (p *parserT) constructLeafNode(parent, child ruleState, proto protoNode) AstNode {

	// A non root leaft node is constructed as a normal leaf node with the appropriate parent and address.
	if child.addr.Depth > 0 {
		return p._constructLeafNode(parent, child, proto)
	}

	// Leaf nodes are not allowed at the root level;
	// they must be contained within an Cluster scoped inner node.
	// This allows the engine to evaluate the match at the cluster level where the publish logic is executing.

	// Construct a placeholder inner node to hold the leaf terms, with the appropriate parent and address.
	// This will assume the original child's address, and the leaf node will be addressed as a child of this placeholder node.
	root := &AstInnerNodeT{
		baseAst: baseAst{ty: proto.ty, address: *child.addr, parent: nil, scope: AstScopeCluster},
	}

	grandChild := child.pushNode(proto.ty)

	leaf := p._constructLeafNode(child, grandChild, proto)

	term := AstTermT{
		Term: leaf,
	}

	root.Terms = []AstTermT{term}

	return root
}

func (p *parserT) _constructLeafNode(parent, child ruleState, proto protoNode) *AstMatchLeafT {

	// Translate node type from proto to the appropriate match type; ie. Set -> MatchSet, Seq -> MatchSeq.
	translatedType := proto.ty
	switch proto.ty {
	case AstNodeTypeSet:
		translatedType = AstNodeTypeMatchSet
	case AstNodeTypeSeq:
		translatedType = AstNodeTypeMatchSeq
	}

	child.addr.Type = translatedType

	return &AstMatchLeafT{
		baseAst:      baseAst{ty: translatedType, address: *child.addr, parent: parent.addr, scope: AstScopeNode},
		Window:       proto.window,
		Correlations: proto.correlations,
		Terms:        protoTermsToAstFields(proto.terms),
		Negate:       protoTermsToAstFields(proto.negate),
		Event:        *proto.event,
	}
}

func (p *parserT) constructInnerNode(parent, child ruleState, proto protoNode) AstNode {

	return &AstInnerNodeT{
		baseAst:      baseAst{ty: proto.ty, address: *child.addr, parent: parent.addr, scope: AstScopeCluster},
		Window:       proto.window,
		Correlations: proto.correlations,
		Terms:        protoTermsToAstTerms(proto.terms),
		Negate:       protoTermsToAstTerms(proto.negate),
	}
}
