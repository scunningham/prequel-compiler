package ast

import (
	"testing"

	"github.com/goccy/go-yaml/ast"
)

func TestParseTerms(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
		negateOff int
		strict    bool
		wantErr   error
		wantPos   int
	}{
		{
			name:   "valid term with one field",
			strict: true,
			yamlInput: `
terms:
  - "simple string match"
`,
		},
		{
			name:   "valid term with one child",
			strict: true,
			yamlInput: `
terms:
  - set:
      event:
        source: "syslog"
        origin: true
      match:
      - "child node match"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := mustParseYAMLNode(t, tt.yamlInput)
			// Should be mapping Node with single key "terms"
			mapping, ok := node.(*ast.MappingNode)
			if !ok || len(mapping.Values) != 1 {
				t.Fatalf("expected a mapping node with one key, got %T with %d keys", node, len(mapping.Values))
			}
			v := mapping.Values[0]
			p := &parserT{strict: tt.strict, root: node, maxRank: 10, maxDepth: 5}
			state := newRuleState(&AstMetadataT{Id: "test", Hash: "hash"})
			_, err := p.parseTerms(state, v.Value, tt.negateOff)

			if !checkParserError(t, err, tt.wantErr, tt.wantPos) {
				return
			}

			// if !reflect.DeepEqual(got, tt.want) {
			// 	t.Errorf("parseTerms() = %+v, want %+v", got, tt.want)
			// }
		})
	}
}
