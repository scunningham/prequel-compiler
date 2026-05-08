package ast

import (
	"fmt"
	"time"

	"github.com/goccy/go-yaml/ast"
)

// 	Window       string       `yaml:"window,omitempty"`
// 	Correlations []string     `yaml:"correlations,omitempty"`
// 	Event        *ParseEventT `yaml:"event,omitempty"`
// 	Match        []ParseTermT `yaml:"match,omitempty"`
// 	Order        []ParseTermT `yaml:"order,omitempty"`
// 	Negate       []ParseTermT `yaml:"negate,omitempty"`

func (p *parserT) parseInnerNode(state ruleState, ty AstNodeType, node ast.Node) (AstNode, error) {

	mapping, err := p.nodeToMapping(node)
	if err != nil {
		return nil, err
	}

	var (
		proto      = protoNode{ty: ty, window: -1} // Default to -1 to indicate no window specified; a window of 0 is valid and means "match events that occur at the same time".}
		child      = state.pushNode(ty).setRank(0) // Reset the rank for the child node; the parent rank should not affect the rank of terms within a set or sequence.
		negateNode ast.Node
	)

	// Sanity check on child address depth; this should be after pushing the child node
	// since that is when the depth is incremented.
	// Note: maxDepth is one based, whereas addr.Depth is zero based, so we check if Depth+1 exceeds maxDepth.
	if child.addr.Depth >= p.maxDepth {
		err := fmt.Errorf("%w: %d", ErrMaxDepthExceeded, p.maxDepth)
		return nil, p.wrapErrorParent(node, err)
	}

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
			if proto.window, err = p.parseWindow(v.Value); err != nil {
				return nil, err
			}

		case kwMatch:
			if proto.ty != AstNodeTypeSet {
				err := fmt.Errorf("%w: '%s' key is not allowed in this context", ErrUnexpectedKey, key)
				return nil, p.wrapError(v.Key, err)
			}
			if proto.terms, err = p.parseTerms(child, v.Value, 0); err != nil {
				return nil, err
			}

		case kwOrder:
			if proto.ty != AstNodeTypeSeq {
				err := fmt.Errorf("%w: '%s' key is not allowed in this context", ErrUnexpectedKey, key)
				return nil, p.wrapError(v.Key, err)
			}
			if proto.terms, err = p.parseTerms(child, v.Value, 0); err != nil {
				return nil, err
			}

		case kwNegate:
			// Defer parsing the negate node until the end in case it occurs before the match/order node.
			// This is necessary to get the addressing consistent; ie. negative terms successively addressed
			// after positive terms.
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

	// Sanity checks
	switch {
	case len(proto.terms) == 0:
		// at least something has to match.
		return nil, p.wrapErrorParent(mapping, ErrMissingTerm)
	case ty == AstNodeTypeSeq && len(proto.terms) == 1 && proto.terms[0].count() <= 1:
		return nil, p.wrapError(findKey(mapping, kwOrder), ErrShortSequence)
	case proto.window < 0 && len(proto.terms) > 1:
		return nil, p.wrapErrorParent(mapping, ErrMissingWindow)
	}

	if negateNode != nil {
		// Fix up the rank on the state to include the already parsed match/order terms,
		// so that negate terms are ranked after them.
		var (
			negateOffset = len(proto.terms)
			negateState  = child.setRank(uint32(negateOffset))
		)
		if proto.negate, err = p.parseTerms(negateState, negateNode, negateOffset); err != nil {
			return nil, err
		}
	}

	p.maybeFixupEventOrigin(proto, state)

	return p.constructNode(state, child, mapping, proto)
}

// If not strict and no origin specified, determine if there is exactly
// one leaf term in the rule, and if so, assign origin to that term.
// This is deprecated behavior; new rules should explicitly specify an origin.
func (p *parserT) maybeFixupEventOrigin(proto protoNode, state ruleState) {

	switch {
	case p.strict:
		// In strict mode, origin must be explicitly specified; do not attempt to infer or assign it.
	case state.getOrigin() > 0:
		// If origin is already set, do not attempt to infer or assign it.
	case state.addr != nil:
		// If this is not the root node, do not attempt to infer or assign origin; it must be set at the root level if applicable.
	case len(proto.terms) == 0:
		// If there are no match/order terms, do not attempt to infer or assign origin.
	case proto.terms[0].leaf == nil:
		// Either all leaves or no leaves; if the first term is not a leaf, do not attempt to infer or assign origin.
	case proto.event == nil:
		// If there is no event, do not attempt to infer or assign origin; origin only applies if there is an event.
	default:
		// Force origin to be true and increment the origin count in the state to reflect this assignment.
		proto.event.Origin = true
		state.incOrigin()
	}
}

func (p *parserT) parseWindow(node ast.Node) (time.Duration, error) {
	window, err := p.nodeToDuration(node)
	if err != nil {
		return 0, err
	}
	if window < 0 {
		return 0, p.wrapError(node, ErrWindowNegative)
	}
	return window, nil
}

// Interpret the prototype and construct the appropriate AST node (SetNode, SequenceNode, etc).

func (p *parserT) constructNode(parent, child ruleState, mapping *ast.MappingNode, proto protoNode) (AstNode, error) {

	// If there are negate terms, they must either all be leaf terms or all inner node terms,
	// and they must match the type of the match/order terms.
	allLeaves := proto.terms[0].leaf != nil

	if len(proto.negate) > 0 {
		if negateAllLeaves := proto.negate[0].leaf != nil; allLeaves != negateAllLeaves {
			err := fmt.Errorf("%w: match terms and negate terms must both be either leaves or inner nodes", ErrUnexpectedType)
			return nil, p.wrapErrorParent(mapping, err)
		}
	}

	// Confirm that the event key is set if required, and not set otherwise.
	switch {
	case allLeaves && proto.event == nil:
		return nil, p.wrapErrorParent(mapping, ErrMissingEvent)

	case !allLeaves && proto.event != nil:
		err := fmt.Errorf("%w: an event is not allowed when using inner node terms", ErrUnexpectedKey)
		return nil, p.wrapError(findKey(mapping, kwEvent), err)
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
		baseAst: baseAst{address: *child.addr, parent: nil, scope: AstScopeCluster},
	}

	grandChild := child.pushNode(AstNodeTypeSet) // Only sets support 1 term.

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
		translatedType = AstNodeTypeLogSet
	case AstNodeTypeSeq:
		translatedType = AstNodeTypeLogSeq
	}

	child.addr.Type = translatedType

	return &AstMatchLeafT{
		baseAst:      baseAst{address: *child.addr, parent: parent.addr, scope: AstScopeNode},
		Window:       proto.window,
		Correlations: proto.correlations,
		Terms:        protoTermsToAstFields(proto.terms),
		Negate:       protoTermsToAstFields(proto.negate),
		Event:        *proto.event,
	}
}

func (p *parserT) constructInnerNode(parent, child ruleState, proto protoNode) AstNode {

	return &AstInnerNodeT{
		baseAst:      baseAst{address: *child.addr, parent: parent.addr, scope: AstScopeCluster},
		Window:       proto.window,
		Correlations: proto.correlations,
		Terms:        protoTermsToAstTerms(proto.terms),
		Negate:       protoTermsToAstTerms(proto.negate),
	}
}
