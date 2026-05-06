package ast

import (
	"errors"
	"reflect"
	"testing"

	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

func mustParseYAMLNode(t *testing.T, src string) ast.Node {
	t.Helper()

	doc, err := parser.ParseBytes([]byte(src), 0)
	if err != nil {
		t.Fatalf("failed to parse yaml: %v", err)
	}

	if len(doc.Docs) != 1 {
		t.Fatalf("expected exactly one document, got %d", len(doc.Docs))
	}

	return doc.Docs[0].Body
}

func TestParseCreNode_TableDriven(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		strict  bool
		wantErr error
		wants   AstCreT
	}{
		{
			name: "valid CRE minimal",
			yaml: `
id: CRE-1234
title: Example CRE
`,
			strict: true,
			wants: AstCreT{
				Id:    "CRE-1234",
				Title: "Example CRE",
			},
		},
		{
			name: "valid CRE with everything",
			yaml: `
id: PREQUEL-2024-0006
severity: 2
title: Kafka Topic Operator Thread Blocked
category: message-queue-problem
author: Prequel
description: |
  There is a known issue in the Strimzi Kafka.
cause: |
  The Kafka Topic Operator is using a single thread to process events.
impact: |
  No Kafka topics will be created or updated.
tags:
- known-problem
- kafka
- strimzi
mitigation: |
  - Add additional CPU resources and restart the Kafka Topic Operator
  - Use the Zookeeper store instead of the Kafka Streams store for the Strimzi Kafka Topic Operator
mitigationScore: 2
impactScore: 8
references:
- https://github.com/strimzi/strimzi-kafka-operator/issues/6046
applications:
- name: "kafka"
`,
			strict: true,
			wants: AstCreT{
				Id:              "PREQUEL-2024-0006",
				Title:           "Kafka Topic Operator Thread Blocked",
				Severity:        2,
				Category:        "message-queue-problem",
				Author:          "Prequel",
				Description:     "There is a known issue in the Strimzi Kafka.\n",
				Cause:           "The Kafka Topic Operator is using a single thread to process events.\n",
				Impact:          "No Kafka topics will be created or updated.\n",
				Tags:            []string{"known-problem", "kafka", "strimzi"},
				Mitigation:      "- Add additional CPU resources and restart the Kafka Topic Operator\n- Use the Zookeeper store instead of the Kafka Streams store for the Strimzi Kafka Topic Operator\n",
				MitigationScore: 2,
				ImpactScore:     8,
				References:      []string{"https://github.com/strimzi/strimzi-kafka-operator/issues/6046"},
				Applications:    []AstAppT{{Name: "kafka"}},
			},
		},
		{
			name: "invalid id (too short)",
			yaml: `
id: ab
title: Bad CRE
`,
			strict:  true,
			wantErr: ErrBadIdentifier,
		},
		{
			name: "unexpected key in strict mode",
			yaml: `
id: CRE-9999
title: Strict CRE
unexpected: value
`,
			strict:  true,
			wantErr: ErrUnexpectedKey,
		},
		{
			name: "unexpected key in non-strict mode",
			yaml: `
id: CRE-9999
title: NonStrict CRE
unexpected: value
`,
			strict:  false,
			wantErr: nil,
			wants: AstCreT{
				Id:    "CRE-9999",
				Title: "NonStrict CRE",
			},
		},
		{
			name: "valid CRE with applications",
			yaml: `
id: CRE-8888
title: App CRE
applications:
  - name: app1
    version: v1
  - name: app2
    version: v2
`,
			strict: true,
			wants: AstCreT{
				Id:    "CRE-8888",
				Title: "App CRE",
				Applications: []AstAppT{
					{Name: "app1", Version: "v1"},
					{Name: "app2", Version: "v2"},
				},
			},
		},
		{
			name:   "valid severity value",
			yaml:   `severity: 3`,
			strict: true,
			wants: AstCreT{
				Severity: SeverityLow,
			},
		},
		{
			name:    "negative severity value",
			yaml:    `severity: -1`,
			strict:  true,
			wantErr: ErrUnexpectedType,
		},
		{
			name:    "invalid severity value",
			yaml:    `severity: 11`,
			strict:  true,
			wantErr: ErrBadSeverity,
		},
		{
			name:    "invalid node type",
			yaml:    `shrubbery`,
			wantErr: ErrUnexpectedType,
		},
		{
			name:    "invalid mapping key type",
			yaml:    `11: invalid key type`,
			wantErr: ErrUnexpectedType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := mustParseYAMLNode(t, tt.yaml)
			p := &parserT{strict: tt.strict}
			cre, err := p.parseCreNode(node)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(*cre, tt.wants) {
				t.Errorf("expected CRE %+v, got %+v", tt.wants, *cre)
			}
		})
	}
}
