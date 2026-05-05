package compiler

import (
	"fmt"

	"github.com/prequel-dev/prequel-compiler/pkg/ast"
	"github.com/prequel-dev/prequel-compiler/pkg/schema"
	"github.com/rs/zerolog/log"
)

type DefaultPlugin struct{}

func NewDefaultPlugin() *DefaultPlugin {
	return &DefaultPlugin{}
}

func (p *DefaultPlugin) Compile(runtime RuntimeI, node *ast.AstNodeT) (ObjsT, error) {

	var (
		objs = make(ObjsT, 0)
		obj  *ObjT
		err  error
	)

	switch node.Metadata.Type {
	case schema.NodeTypeLogSeq, schema.NodeTypeLogSet:
		if obj, err = ObjLogMatcher(runtime, node); err != nil {
			log.Error().Err(err).Str("scope", node.Metadata.Scope).Msg("Failed to compile matchers")
			return nil, err
		}
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedNodeType, node.Metadata.Type)
	}

	objs = append(objs, obj)

	return objs, nil
}
