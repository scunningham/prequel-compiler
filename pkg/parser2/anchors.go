package parser2

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

func LoadAnchorsFromDir(dir string) (map[string]ast.Node, error) {
	files, err := collectYamlFilesFromDir(dir)
	if err != nil {
		return nil, err
	}

	anchors := make(map[string]ast.Node)

	for _, file := range files {
		fileAnchors, err := loadAnchorsFromFile(file)
		if err != nil {
			return nil, err
		}

		var dupe string
		anchors, dupe = mergeAnchors(anchors, fileAnchors)
		if dupe != "" {
			return nil, fmt.Errorf("duplicate anchor name '%s' in file %s", dupe, file)
		}
	}

	return anchors, nil
}

func collectYamlFilesFromDir(dir string) (files []string, err error) {

	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {

		switch {
		case err != nil:
			return err
		case d.IsDir():
			return nil
		case !isYAMLFile(path):
			return nil
		}

		files = append(files, path)
		return nil
	})

	return
}

func mergeAnchors(a, b map[string]ast.Node) (map[string]ast.Node, string) {
	if a == nil {
		return b, ""
	}

	for k := range a {
		if _, exists := b[k]; exists {
			return nil, k
		}
	}

	for k, v := range b {
		a[k] = v
	}
	return a, ""
}

func loadAnchorsFromFile(path string) (map[string]ast.Node, error) {
	rdr, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer rdr.Close()
	return collectAnchors(rdr)
}

func collectAnchors(rdr io.Reader) (map[string]ast.Node, error) {
	data, err := io.ReadAll(rdr)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}

	file, err := parser.ParseBytes(data, 0)
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	anchors := make(map[string]ast.Node)
	for _, doc := range file.Docs {
		if err := _collectAnchors(doc, anchors); err != nil {
			return nil, err
		}
	}

	return anchors, nil
}

func isYAMLFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".yaml" || ext == ".yml"
}

// Recursively walk the AST and collect anchor nodes.
func _collectAnchors(node ast.Node, anchors map[string]ast.Node) error {
	if node == nil {
		return nil
	}

	// If this node has an anchor, record it.
	if anchorNode, ok := node.(*ast.AnchorNode); ok {
		// Dedupe anchor name

		name := anchorNode.Name.String()
		if _, exists := anchors[name]; exists {
			return fmt.Errorf("duplicate anchor name '%s'", name)
		}

		anchors[name] = anchorNode.Value
		return _collectAnchors(anchorNode.Value, anchors)
	}

	switch n := node.(type) {

	case *ast.MappingNode:
		for _, value := range n.Values {
			if err := _collectAnchors(value.Key, anchors); err != nil {
				return err
			}
			if err := _collectAnchors(value.Value, anchors); err != nil {
				return err
			}
		}

	case *ast.SequenceNode:
		for _, value := range n.Values {
			if err := _collectAnchors(value, anchors); err != nil {
				return err
			}
		}

	case *ast.DocumentNode:
		if err := _collectAnchors(n.Body, anchors); err != nil {
			return err
		}
	}
	return nil
}
