package ast

import (
	"errors"
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

func (p *parserT) parseRulesNode(node ast.Node) ([]AstRuleT, error) {

	seq, err := p.nodeToSequence(node)
	if err != nil {
		return nil, err
	}

	var (
		rules   []AstRuleT
		errList []error
	)

	for _, ruleNode := range seq.Values {

		rule, err := p.parseRuleNode(ruleNode)

		switch {
		case err != nil:
			errList = append(errList, err)
		default:
			rules = append(rules, *rule)
		}
	}

	return rules, errors.Join(errList...)
}

// Expects layout of a single mapping node with keys 'metadata', 'cre', and 'rule'
// Optionally there is a version filter; with a required version with an optional "<" prefix.

func (p *parserT) parseRuleNode(node ast.Node) (*AstRuleT, error) {

	mapping, err := p.nodeToMapping(node)
	if err != nil {
		return nil, err
	}

	var (
		meta    *AstMetadataT
		cre     *AstCreT
		ruleDom ast.Node
	)

	maybeMeta := func(err error) error {
		if meta == nil {
			return err
		}
		return ErrRule{
			Meta: *meta,
			Err:  err,
		}
	}

	for _, v := range mapping.Values {

		key, err := p.nodeToString(v.Key)
		if err != nil {
			return nil, maybeMeta(err)
		}

		switch key {

		case kwMetadata:
			meta, err = p.parseMetadataNode(v.Value)

		case kwCre:
			cre, err = p.parseCreNode(v.Value)

		case kwRule:
			ruleDom = v.Value

		case kwCompiler:
			// TODO: Handle compiler filter

		default:
			err = p.wrapError(v, fmt.Errorf("%w: %s", ErrUnexpectedKey, key))
		}

		if err != nil {
			return nil, maybeMeta(err)
		}
	}

	switch {
	case meta == nil:
		err := fmt.Errorf("%w: %s", ErrMissingKey, kwMetadata)
		return nil, p.wrapError(mapping, err)
	case ruleDom == nil:
		kerr := fmt.Errorf("%w: %s", ErrMissingKey, kwRule)
		return nil, maybeMeta(p.wrapError(mapping, kerr))
	}

	// Parse the root
	var (
		rootAst   AstNode
		ruleState = newRuleState(meta)
	)

	if rootAst, err = p.parseRootNode(ruleState, ruleDom); err != nil {
		return nil, maybeMeta(err)
	}

	rule := &AstRuleT{
		Cre:      cre,
		Metadata: *meta,
		Root:     rootAst,
	}

	return rule, nil
}

// At the root, expecting:
// type ParseRuleDataT struct {
// 	Sequence *ParseSequenceT `yaml:"sequence,omitempty"`
// 	Set      *ParseSetT      `yaml:"set,omitempty"`
// }

func (p *parserT) parseRootNode(state ruleState, node ast.Node) (AstNode, error) {

	mapping, err := p.nodeToMapping(node)
	if err != nil {
		return nil, err
	}

	var (
		rootNode AstNode
	)

	alreadySet := func(v ast.Node) error {
		if rootNode != nil {
			err := fmt.Errorf("%w: multiple root keys found in rule definition", ErrUnexpectedKey)
			return p.wrapError(v, err)
		}
		return nil
	}

	for _, v := range mapping.Values {

		key, err := p.nodeToString(v.Key)
		if err != nil {
			return nil, err
		}
		switch key {

		case kwSequence:
			if err := alreadySet(v); err != nil {
				return nil, err
			}
			if rootNode, err = p.parseInnerNode(state, AstNodeTypeSeq, v.Value); err != nil {
				return nil, err
			}

		case kwSet:
			if err := alreadySet(v); err != nil {
				return nil, err
			}
			if rootNode, err = p.parseInnerNode(state, AstNodeTypeSet, v.Value); err != nil {
				return nil, err
			}

		default:
			err := fmt.Errorf("%w: only '%s' or '%s' expected in rule root, not %s", ErrUnexpectedKey, kwSequence, kwSet, key)
			return nil, p.wrapError(v.Key, err)

		}
	}

	if rootNode == nil {
		err := fmt.Errorf("%w: expected rule root to contain either '%s' or '%s' key", ErrMissingKey, kwSequence, kwSet)
		return nil, p.wrapError(mapping, err)
	}

	return rootNode, nil
}
