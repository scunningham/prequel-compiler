package parser2

import (
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

// type ParseScriptT struct {
// 	Code     string      `yaml:"code"`
// 	Language string      `yaml:"language,omitempty"` // Assumes 'lua' if empty
// 	Timeout  string      `yaml:"timeout,omitempty"`  // Uses default if empty; expects duration string
// 	Input    *ParseTermT `yaml:"input"`              // Required input
// }

func (p *parserT) parseScriptNode(state ruleState, node ast.Node) (*AstScriptT, error) {

	mapping, err := p.nodeToMapping(node)
	if err != nil {
		return nil, err
	}

	child := state.pushChild(AstNodeTypeScript)

	var (
		script = AstScriptT{
			baseAst: baseAst{
				ty:      AstNodeTypeScript,
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

		case kwScriptCode:
			script.Code, err = p.nodeToString(v.Value)
			// TODO: validate script code here

		case kwScriptLang:
			script.Language, err = p.nodeToString(v.Value)
			// TODO: validate supported languages here

		case kwScriptTimeout:
			script.Timeout, err = p.nodeToDuration(v.Value)

		case kwScriptInput:
			script.Input, err = p.parseScriptInput(child, v.Value)

		default:
			if p.strict {
				err = p.wrapError(v, fmt.Errorf("%w: %s", ErrUnexpectedKey, key))
			}
		}

		if err != nil {
			return nil, err
		}
	}

	return &script, nil
}

func (p *parserT) parseScriptInput(state ruleState, node ast.Node) (AstNode, error) {

	return p.parseRootNode(state, node)

	// if err != nil {
	// 	return
	// }

	// // Expect exactly one term in the input definition
	// if len(protoTerms) != 1 {
	// 	tErr := fmt.Errorf("%w: script input must contain exactly one term, found %d", ErrUnexpectedKey, len(terms))
	// 	err = p.wrapError(node, tErr)
	// 	return
	// }

	// terms := protoTermsToAstTerms(protoTerms)

	// return terms[0], nil

}
