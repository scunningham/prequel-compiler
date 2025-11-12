package ast

import (
	"time"

	"github.com/prequel-dev/prequel-compiler/pkg/parser"
	"github.com/prequel-dev/prequel-compiler/pkg/schema"
	"github.com/rs/zerolog/log"
)

type AstPromQL struct {
	Query    string
	External string
	Interval time.Duration
}

func (b *builderT) buildPromQLNode(parserNode *parser.NodeT, machineAddress *AstNodeAddressT, termIdx *uint32) (*AstNodeT, error) {

	// Expects on child of type ParsePromQL

	if len(parserNode.Children) != 1 {
		log.Error().Int("child_count", len(parserNode.Children)).Msg("PromQL node must have exactly one child")
		return nil, parserNode.WrapError(ErrInvalidNodeType)
	}

	promNode, ok := parserNode.Children[0].(*parser.PromQLT)

	if !ok {
		log.Error().Interface("promql", parserNode.Children[0]).Msg("Failed to build PromQL node")
		return nil, parserNode.WrapError(ErrMissingScalar)
	}

	if promNode.Query == "" {
		log.Error().Msg("PromQL query string is empty")
		return nil, parserNode.WrapError(ErrMissingScalar)
	}

	if parserNode.Metadata.Event != nil && parserNode.Metadata.Event.Origin {
		b.HasOrigin = true
	}

	pn := &AstPromQL{
		Query: promNode.Query,
	}

	if promNode.Interval != nil {
		pn.Interval = *promNode.Interval
	}

	var (
		address = b.newAstNodeAddress(parserNode.Metadata.RuleHash, parserNode.Metadata.Type.String(), termIdx)
		node    = newAstNode(parserNode, parserNode.Metadata.Type, schema.ScopeNode, machineAddress, address)
	)

	node.Object = pn
	return node, nil

}
