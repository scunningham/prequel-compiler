package ast

import (
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

func (p *parserT) parseMetadataNode(node ast.Node) (*AstMetadataT, error) {

	mapping, err := p.nodeToMapping(node)
	if err != nil {
		return nil, err
	}

	var (
		meta AstMetadataT
	)

	for _, v := range mapping.Values {

		key, err := p.nodeToString(v.Key)
		if err != nil {
			return nil, err
		}

		switch key {

		case kwName:
			meta.Name, err = p.nodeToString(v.Value)

		case kwId:
			meta.Id, err = p.parseIdNode(v.Value)

		case kwHash:
			meta.Hash, err = p.parseHash(v.Value)

		case kwGen:
			meta.Gen, err = p.parseGen(v.Value)

		case kwKind:
			meta.Kind, err = p.parseKind(v.Value)

		default:
			if p.strict {
				err = p.wrapError(v, fmt.Errorf("%w: %s", ErrUnexpectedKey, key))
			}
		}

		if err != nil {
			return nil, err
		}
	}

	return &meta, nil
}

func (p *parserT) parseIdNode(v ast.Node) (string, error) {
	s, err := p.nodeToString(v)
	if err != nil {
		return "", err
	}

	// Expect string to be a randomized 16 byte wide base58 encoded string.
	// This is not a hard requirement, but implies a wide enough value to avoid collisions.
	// For this valdiation, we will just check the size.
	if p.strict && len(s) < 16 {
		err := fmt.Errorf("%w: id value must at least 16 characters", ErrBadIdentifier)
		return "", p.wrapError(v, err)
	}
	return s, nil
}

func (p *parserT) parseHash(v ast.Node) (string, error) {
	s, err := p.nodeToString(v)
	if err != nil {
		return "", err
	}
	// Expect base58 encoded sha256 hash.
	if p.strict && (len(s) < 32 || len(s) > 44) {
		err := fmt.Errorf("%w: hash value must be at least 32 characters", ErrBadHash)
		return "", p.wrapError(v, err)
	}
	return s, nil
}

func (p *parserT) parseGen(v ast.Node) (uint, error) {
	gen, err := p.nodeToUint64(v)
	if err != nil {
		return 0, err
	}
	// Sanity check generation number; should be a positive integer, and not unreasonably high.
	if gen > maxGen {
		err := fmt.Errorf("%w: generation value must be a positive integer less or equal to %d", ErrBadGen, maxGen)
		return 0, p.wrapError(v, err)
	}

	return uint(gen), nil
}

func (p *parserT) parseKind(v ast.Node) (string, error) {
	s, err := p.nodeToString(v)
	if err != nil {
		return "", err
	}
	switch s {
	case KindPrequel, KindCustom:
	default:
		if p.strict {
			err := fmt.Errorf("%w: kind value must be either 'prequel' or 'custom'", ErrBadKind)
			return "", p.wrapError(v, err)
		}
	}
	return s, nil
}
