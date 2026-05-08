package ast

import (
	"reflect"
	"testing"

	"github.com/goccy/go-yaml/ast"
	"github.com/prequel-dev/prequel-logmatch/pkg/match"
)

const (
	stubRuleId   = "test"
	stubRuleHash = "hash"
	stubSource   = "stubSource"
)

func TestParseTerms(t *testing.T) {
	tests := []struct {
		name      string
		yamlInput string
		negateOff int
		strict    bool
		want      []*protoTerm
		wantErr   error
		wantPos   int
	}{
		{
			name: "valid term with one field",
			yamlInput: `
terms:
  - "simple string match"
`,
			want: []*protoTerm{
				{
					field: &protoField{
						StrValue: "simple string match",
					},
				},
			},
		},
		{
			name: "valid term with two fields",
			yamlInput: `
terms:
  - "simple string match 1"
  - "simple string match 2"
`,
			want: []*protoTerm{
				{
					field: &protoField{
						StrValue: "simple string match 1",
					},
				},
				{
					field: &protoField{
						StrValue: "simple string match 2",
					},
				},
			},
		},
		{
			name: "valid term with one child",
			yamlInput: `
terms:
  - set:
      event:
        source: stubSource
        origin: true
      match:
      - "child node match"
`,
			want: []*protoTerm{
				{
					child: genStubChild([]string{"child node match"}),
				},
			},
		},
		{
			name:   "valid term with two children",
			strict: true,
			yamlInput: `
terms:
  - set:
      event:
        source: stubSource
        origin: true
      match:
      - "child node match 1"
  - set:
      event:
        source: stubSource
      match:
      - "child node match 2"
`,
			want: genStubChildren([]string{"child node match 1", "child node match 2"}),
		},
		{
			name: "invalid term with mixed field and child nodes",
			yamlInput: `
terms:
  - "simple string match"
  - set:
      event:
        source: stubSource
      match:
      - "child node match 2"
`,
			wantErr: ErrTermTypeConflict,
			wantPos: 42, // Position of the "set" key in the second term, which is where the type inconsistency is detected.
		},
		{
			name: "invalid term with mixed field and child nodes reversed",
			yamlInput: `
terms:
  - set:
      event:
        source: stubSource
      match:
      - "child node match 1"
  - "simple string match"
`,
			wantErr: ErrTermTypeConflict,
			wantPos: 104, // Position of the second term node (the "simple string match" string), which is where the type inconsistency is detected.
		},
		{
			name: "test bad terms type",
			yamlInput: `
terms: "not a sequence"
`,
			wantErr: ErrUnexpectedType,
			wantPos: 9, // Position of the "not a sequence" value node (string instead of sequence)
		},
		{
			name: "empty terms strict mode",
			yamlInput: `
terms: []
`,
			wantErr: ErrMissingTerm,
			strict:  true,
			wantPos: 9, // Position of the empty sequence node, which is where the missing term error is detected in strict mode.
		},
		{
			name: "empty terms tolerant mode",
			yamlInput: `
terms: []
`,
			wantErr: nil,
			strict:  false,
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
			state := newRuleState(&AstMetadataT{Id: stubRuleId, Hash: stubRuleHash})
			got, err := p.parseTerms(state, v.Value, tt.negateOff)

			if !checkParserError(t, err, tt.wantErr, tt.wantPos) {
				return
			}

			if tt.want != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseTerms() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func genStubChildren(rawMatches []string) []*protoTerm {
	children := make([]*protoTerm, 0, len(rawMatches))
	for i, v := range rawMatches {
		u := uint32(i)
		stub := _genStubChild(2*u, u, i == 0, []string{v})
		children = append(children, &protoTerm{child: stub})
	}
	return children
}

func genStubChild(rawMatches []string) *AstInnerNodeT {
	return _genStubChild(0, 0, true, rawMatches)
}

// genStubChild is a helper function to generate a stub child node for testing purposes,
// given a list of raw match strings.
func _genStubChild(nodeId, rank uint32, origin bool, rawMatches []string) *AstInnerNodeT {

	parentAddr := &AstNodeAddressT{
		Type:     AstNodeTypeSet,
		RuleId:   stubRuleId,
		RuleHash: stubRuleHash,
		NodeId:   nodeId,
		Rank:     rank,
	}

	leaf := &AstMatchLeafT{
		baseAst: baseAst{
			scope: AstScopeNode,
			address: AstNodeAddressT{
				Type:     AstNodeTypeLogSet,
				RuleId:   stubRuleId,
				RuleHash: stubRuleHash,
				Depth:    1,
				NodeId:   nodeId + 1,
			},
			parent: parentAddr,
		},
		Window: 0,
		Terms:  []AstFieldT{},
		Event: AstEventT{
			Source: stubSource,
			Origin: origin,
		},
	}
	for _, raw := range rawMatches {
		leaf.Terms = append(leaf.Terms, AstFieldT{
			Count: 1,
			TermValue: match.TermT{
				Type:  match.TermRaw,
				Value: raw,
			},
		})
	}

	return &AstInnerNodeT{
		baseAst: baseAst{
			scope:   AstScopeCluster,
			address: *parentAddr,
		},
		Terms: []AstTermT{{Term: leaf}},
	}
}
