package parser2

import (
	"bytes"
	"fmt"

	"github.com/goccy/go-yaml/ast"
)

type parseOpts struct {
	strict     bool
	colorize   bool
	anchorDir  string
	anchorData []byte
}

type ParseOpt func(*parseOpts)

func WithStrict(strict bool) ParseOpt {
	return func(opts *parseOpts) {
		opts.strict = strict
	}
}

func WithColorize(colorize bool) ParseOpt {
	return func(opts *parseOpts) {
		opts.colorize = colorize
	}
}

func WithAnchorDir(dir string) ParseOpt {
	return func(opts *parseOpts) {
		opts.anchorDir = dir
	}
}

func WithAnchorYAML(data []byte) ParseOpt {
	return func(opts *parseOpts) {
		opts.anchorData = data
	}
}

func _parseOpts(opts []ParseOpt) parseOpts {
	var o parseOpts

	for _, opt := range opts {
		opt(&o)
	}

	return o
}

func maybeAnchors(o parseOpts) (map[string][]byte, error) {

	var anchors map[string]ast.Node

	if len(o.anchorDir) > 0 {
		var err error
		anchors, err = LoadAnchorsFromDir(o.anchorDir)
		if err != nil {
			return nil, err
		}
	}

	if len(o.anchorData) > 0 {
		m, err := collectAnchors(bytes.NewReader(o.anchorData))
		if err != nil {
			return nil, err
		}

		if len(m) > 0 {
			var dupe string
			anchors, dupe = mergeAnchors(anchors, m)
			if dupe != "" {
				return nil, fmt.Errorf("duplicate anchor name '%s' in anchor YAML", dupe)
			}
		}
	}

	if len(anchors) == 0 {
		return nil, nil
	}

	// Convert the anchor nodes to raw YAML bytes so they can be injected into the parser's anchor map.
	anchorBytes := make(map[string][]byte, len(anchors))
	for k, v := range anchors {
		b, err := v.MarshalYAML()
		if err != nil {
			return nil, err
		}
		anchorBytes[k] = b
	}

	// Convert the anchor nodes to raw YAML bytes so they can be injected into the parser's anchor map.

	return anchorBytes, nil
}
