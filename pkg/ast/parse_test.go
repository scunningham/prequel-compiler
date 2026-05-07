package ast

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/prequel-dev/prequel-compiler/pkg/testdata"
)

func TestAstSuccess(t *testing.T) {

	tests := []struct {
		name      string
		yaml      string
		wantOrder []string
	}{
		{
			name: "Success_Simple1",
			yaml: testdata.TestSuccessSimpleRule1,
			wantOrder: []string{
				"v1.machine_seq.rdJLgqYgkEp8jg8Qks1qiq.d0.n0.t0",
				"v1.log_seq.rdJLgqYgkEp8jg8Qks1qiq.d1.n1.t0",
			},
		},
		{
			name: "Success_Complex2",
			yaml: testdata.TestSuccessComplexRule2,
			wantOrder: []string{
				"v1.machine_seq.rdJLgqYgkEp8jg8Qks1qiq.d0.n0.t0",
				"v1.log_seq.rdJLgqYgkEp8jg8Qks1qiq.d1.n1.t0",
				"v1.log_set.rdJLgqYgkEp8jg8Qks1qiq.d1.n2.t1",
				"v1.machine_seq.rdJLgqYgkEp8jg8Qks1qiq.d1.n3.t2",
				"v1.log_seq.rdJLgqYgkEp8jg8Qks1qiq.d2.n4.t0",
				"v1.log_set.rdJLgqYgkEp8jg8Qks1qiq.d2.n5.t1",
				"v1.log_set.rdJLgqYgkEp8jg8Qks1qiq.d2.n6.t2",
			},
		},
		{
			name: "Success_Complex3",
			yaml: testdata.TestSuccessComplexRule3,
			wantOrder: []string{
				"v1.machine_seq.rdJLgqYgkEp8jg8Qks1qiq.d0.n0.t0",
				"v1.log_seq.rdJLgqYgkEp8jg8Qks1qiq.d1.n1.t0",
				"v1.log_set.rdJLgqYgkEp8jg8Qks1qiq.d1.n2.t1",
			},
		},
		{
			name: "Success_Complex4",
			yaml: testdata.TestSuccessComplexRule4,
			wantOrder: []string{
				"v1.machine_seq.2KdXQZDAfRbYcH9FBDteBS.d0.n0.t0",
				"v1.log_seq.2KdXQZDAfRbYcH9FBDteBS.d1.n1.t0",
				"v1.machine_seq.2KdXQZDAfRbYcH9FBDteBS.d1.n2.t1",
				"v1.log_seq.2KdXQZDAfRbYcH9FBDteBS.d2.n3.t0",
				"v1.log_set.2KdXQZDAfRbYcH9FBDteBS.d2.n4.t1",
				"v1.log_set.2KdXQZDAfRbYcH9FBDteBS.d2.n5.t2",
				"v1.machine_seq.2KdXQZDAfRbYcH9FBDteBS.d1.n6.t2",
				"v1.log_seq.2KdXQZDAfRbYcH9FBDteBS.d2.n7.t0",
				"v1.log_set.2KdXQZDAfRbYcH9FBDteBS.d2.n8.t1",
				"v1.log_set.2KdXQZDAfRbYcH9FBDteBS.d2.n9.t2",
				"v1.log_set.2KdXQZDAfRbYcH9FBDteBS.d1.n10.t3",
			},
		},
		{
			name: "Success_NegateOptions1",
			yaml: testdata.TestSuccessNegateOptions1,
			wantOrder: []string{
				"v1.machine_seq.rdJLgqYgkEp8jg8Qks1qiq.d0.n0.t0",
				"v1.log_seq.rdJLgqYgkEp8jg8Qks1qiq.d1.n1.t0",
			},
		},
		{
			name: "Success_NegateOptions2",
			yaml: testdata.TestSuccessNegateOptions2,
			wantOrder: []string{
				"v1.machine_seq.rdJLgqYgkEp8jg8Qks1qiq.d0.n0.t0",
				"v1.log_seq.rdJLgqYgkEp8jg8Qks1qiq.d1.n1.t0",
				"v1.log_set.rdJLgqYgkEp8jg8Qks1qiq.d1.n2.t1",
				"v1.log_set.rdJLgqYgkEp8jg8Qks1qiq.d1.n3.t2",
			},
		},
		{
			name: "Success_Extract1",
			yaml: testdata.TestSuccessSimpleExtraction,
			wantOrder: []string{
				"v1.machine_seq.rdJLgqYgkEp8jg8Qks1qiq.d0.n0.t0",
				"v1.log_seq.rdJLgqYgkEp8jg8Qks1qiq.d1.n1.t0",
			},
		},
		{
			name: "Success_PromQL",
			yaml: testdata.TestSuccessSimplePromQL,
			wantOrder: []string{
				"v1.machine_set.rdJLgqYgkEp8jg8Qks1qiq.d0.n0.t0",
				"v1.promql.rdJLgqYgkEp8jg8Qks1qiq.d1.n1.t0",
				"v1.log_set.rdJLgqYgkEp8jg8Qks1qiq.d1.n2.t1",
			},
		},
		{
			name: "Success_ChildScript",
			yaml: testdata.TestSuccessChildScript,
			wantOrder: []string{
				"v1.machine_seq.rdJLgqYgkEp8jg8Qks1qiq.d0.n0.t0",
				"v1.script.rdJLgqYgkEp8jg8Qks1qiq.d1.n1.t0",
				"v1.log_seq.rdJLgqYgkEp8jg8Qks1qiq.d2.n2.t0",
				"v1.log_set.rdJLgqYgkEp8jg8Qks1qiq.d1.n3.t1",
			},
		},
		{
			name: "Success_ChildScriptMultipleInputs",
			yaml: testdata.TestSuccessChildScriptMultipleInputs,
			wantOrder: []string{
				"v1.machine_set.rdJLgqYgkEp8jg8Qks1qiq.d0.n0.t0",
				"v1.script.rdJLgqYgkEp8jg8Qks1qiq.d1.n1.t0",
				"v1.machine_seq.rdJLgqYgkEp8jg8Qks1qiq.d2.n2.t0",
				"v1.log_seq.rdJLgqYgkEp8jg8Qks1qiq.d3.n3.t0",
				"v1.log_set.rdJLgqYgkEp8jg8Qks1qiq.d3.n4.t1",
			},
		},
		{
			name: "Success_ChildScriptPromQLInput",
			yaml: testdata.TestSuccessChildScriptPromQLInput,
			wantOrder: []string{
				"v1.machine_set.rdJLgqYgkEp8jg8Qks1qiq.d0.n0.t0",
				"v1.script.rdJLgqYgkEp8jg8Qks1qiq.d1.n1.t0",
				"v1.promql.rdJLgqYgkEp8jg8Qks1qiq.d2.n2.t0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rules, err := ParseRules([]byte(tt.yaml), WithStrict(true))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(rules) != 1 {
				t.Fatalf("expected 1 rule, got %d", len(rules))
			}

			order := extractOrder(t, rules[0])

			minLen := min(len(order), len(tt.wantOrder))

			// Compare shared range.
			for i := range minLen {
				if order[i] != tt.wantOrder[i] {
					fmt.Println(Draw(rules[0], WithColor()))
					t.Fatalf("first mismatch at position %d: expected '%s' != '%s'", i, tt.wantOrder[i], order[i])
				}
			}

			// Handle length mismatch.
			if len(order) != len(tt.wantOrder) {
				t.Fatalf("expected order  sz %v, got %v", len(tt.wantOrder), len(order))
			}
		})
	}
}

func TestAstFail(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		strict  bool
		wantErr error
		wantPos int
	}{
		{
			name:    "Fail_MissingPositiveCondition",
			yaml:    testdata.TestFailMissingPositiveCondition,
			wantErr: ErrMissingTerm,
			wantPos: 607,
		},
		{
			name:    "Fail_NegativeCondition1",
			yaml:    testdata.TestFailNegativeCondition1,
			wantErr: ErrMissingTerm,
			wantPos: 635,
		},
		{
			name:    "Fail_NegativeCondition2",
			yaml:    testdata.TestFailNegativeCondition2,
			wantErr: ErrMissingTerm,
			wantPos: 635,
		},
		{
			name:    "Fail_NegativeCondition2_Strict",
			yaml:    testdata.TestFailNegativeCondition2,
			wantErr: ErrUnexpectedKey, // 'imageUrl'
			strict:  true,
			wantPos: 420,
		},
		{
			name:    "Fail_NegativeCondition3",
			yaml:    testdata.TestFailNegateOptions3,
			wantErr: ErrMissingTerm,
			wantPos: 749,
		},
		{
			name:    "Fail_NegativeCondition4",
			yaml:    testdata.TestFailNegateOptions4,
			wantErr: ErrMissingTerm,
			wantPos: 765,
		},
		{
			name:    "Fail_TermsSyntaxError1",
			yaml:    testdata.TestFailTermsSyntaxError1,
			wantErr: ErrUnexpectedKey,
			wantPos: 670,
		},
		{
			name:    "Fail_TermsSyntaxError2",
			yaml:    testdata.TestFailTermsSyntaxError2,
			wantErr: ErrUnexpectedType,
			wantPos: 672,
		},
		{
			name:    "Fail_TermsSemanticError1",
			yaml:    testdata.TestFailTermsSemanticError1,
			wantErr: ErrShortSequence,
			wantPos: 697,
		},
		{
			name:    "Fail_TermsSemanticError2",
			yaml:    testdata.TestFailTermsSemanticError2,
			wantErr: ErrMissingEvent,
			wantPos: 199,
		},
		{
			name:    "Fail_TermsSemanticError3",
			yaml:    testdata.TestFailTermsSemanticError3,
			wantErr: ErrMissingOrigin,
			wantPos: 183,
		},
		{
			name:    "Fail_TermsSemanticError4",
			yaml:    testdata.TestFailTermsSemanticError4,
			wantErr: ErrUnexpectedType,
			wantPos: 314,
		},
		{
			name:    "Fail_TermsSemanticError5",
			yaml:    testdata.TestFailTermsSemanticError5,
			wantErr: ErrBadAnchor,
			wantPos: 411,
		},
		{
			name:    "Fail_TermsSemanticError6",
			yaml:    testdata.TestFailTermsSemanticError6,
			strict:  true,
			wantErr: ErrMissingOrigin,
			wantPos: 183,
		},
		{
			name:   "Fail_TermsSemanticError6_NonStrict",
			yaml:   testdata.TestFailTermsSemanticError6,
			strict: false, // Normally fails in strict mode
		},
		{
			name:    "Fail_MultipleOrigin",
			yaml:    testdata.TestFailMultipleOrigin,
			wantErr: ErrMultipleOrigin,
			wantPos: 502,
		},
		{
			name:    "Fail_Typo",
			yaml:    testdata.TestFailTypo,
			wantErr: ErrUnexpectedKey,
			wantPos: 290,
		},
		{
			name:    "Fail_MissingOrder",
			yaml:    testdata.TestFailMissingOrder,
			wantErr: ErrUnexpectedKey,
			wantPos: 279,
		},
		{
			name:    "Fail_MissingMatch",
			yaml:    testdata.TestFailMissingMatch,
			wantErr: ErrUnexpectedKey,
			wantPos: 274,
		},
		{
			name:    "Fail_InvalidWindow",
			yaml:    testdata.TestFailInvalidWindow,
			wantErr: ErrUnexpectedType,
			wantPos: 224,
		},
		{
			name:    "Fail_UnsupportedRule",
			yaml:    testdata.TestFailUnsupportedRule,
			wantErr: ErrUnexpectedKey,
			wantPos: 203,
		},
		{
			name:    "Fail_MissingCreId",
			yaml:    testdata.TestFailMissingCreRule,
			wantErr: ErrMissingKey,
			wantPos: 36,
		},
		{
			name:    "Fail_MissingRuleId",
			yaml:    testdata.TestFailMissingRuleIdRule,
			wantErr: ErrMissingKey,
			wantPos: 100,
		},
		{
			name:    "Fail_MissingRuleHash",
			yaml:    testdata.TestFailMissingRuleHashRule,
			wantErr: ErrMissingKey,
			wantPos: 102,
		},
		{
			name:    "Fail_BadRuleId",
			yaml:    testdata.TestFailBadRuleIdRule,
			wantErr: ErrBadIdentifier,
			wantPos: 108,
		},
		{
			name:    "Fail_BadCreId",
			yaml:    testdata.TestFailBadCreIdRule,
			wantErr: ErrBadIdentifier,
			wantPos: 48,
		},
		{
			name:    "Fail_BadRuleHash",
			yaml:    testdata.TestFailBadRuleHashRule,
			wantErr: ErrBadHash,
			wantPos: 147,
		},
		{
			name:    "Fail_ScriptRoot",
			yaml:    testdata.TestFailScriptRoot,
			wantErr: ErrUnexpectedKey,
			wantPos: 162,
		},
		{
			name:    "Fail_ScriptNoInput",
			yaml:    testdata.TestFailScriptNoInput,
			wantErr: ErrMissingKey,
			wantPos: 203,
		},
		{
			name:    "Fail_MissingWindow",
			yaml:    testdata.TestFailMissingWindow,
			wantErr: ErrMissingWindow,
			wantPos: 151,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rules, err := ParseRules([]byte(tt.yaml), WithStrict(tt.strict))

			ok := checkParserError(t, err, tt.wantErr, tt.wantPos)

			switch {
			case ok && rules == nil:
				t.Errorf("expected rules to be returned, got nil")
			case !ok && rules != nil:
				t.Errorf("expected no rules to be returned, got %v", rules)
			}
		})
	}
}

// Validate the following invariants on the tree:
// 1. No duplicate addresses
// 2. Root node has no parent address
// 3. Node ids are unique
// 4. Depth is consistent with distance from root

func extractOrder(t *testing.T, rule AstRuleT) []string {
	t.Helper()

	var (
		order    []string
		dupeIds  = make(map[uint32]struct{})
		dupeAddr = make(map[string]struct{})
	)

	err := rule.Walk(func(node AstNode, _ *AstNegateOptsT) error {
		var (
			addr    = node.Address()
			addrStr = addr.String()
		)

		order = append(order, addrStr)

		if _, exists := dupeIds[addr.NodeId]; exists {
			t.Fatalf("Duplicate node ID found: %d", addr.NodeId)
		}
		dupeIds[addr.NodeId] = struct{}{}

		if _, exists := dupeAddr[addrStr]; exists {
			t.Fatalf("Duplicate address found: %s", addrStr)
		}
		dupeAddr[addrStr] = struct{}{}

		// Depth should be one larger than parent's depth;
		// or can rewinde if going back up the tree
		if parentAddr := node.Parent(); parentAddr == nil {
			if addr.Depth != 0 {
				t.Fatalf("Root node has non-zero depth: %d", addr.Depth)
			}
		} else if addr.Depth != parentAddr.Depth+1 {
			t.Fatalf("Node depth inconsistent with parent's depth: %d (parent: %d)", addr.Depth, parentAddr.Depth)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	return order
}

func TestSuccessExamples(t *testing.T) {

	rules, err := filepath.Glob(filepath.Join("../testdata", "success_examples", "*.yaml"))
	if err != nil {
		t.Fatalf("Error finding CRE test files: %v", err)
	}

	for _, rule := range rules {

		t.Run(filepath.Base(rule), func(t *testing.T) {
			testData, err := os.ReadFile(rule)
			if err != nil {
				t.Fatalf("Error reading test file %s: %v", rule, err)
			}

			_, err = ParseRules(testData, WithStrict(true))
			if err != nil {
				t.Fatalf("Error building rule %s: %v", rule, err)
			}
		})
	}
}

func TestFailureExamples(t *testing.T) {

	rules, err := filepath.Glob(filepath.Join("../testdata", "failure_examples", "*.yaml"))
	if err != nil {
		t.Fatalf("Error finding CRE test files: %v", err)
	}

	for _, rule := range rules {

		t.Run(filepath.Base(rule), func(t *testing.T) {

			testData, err := os.ReadFile(rule)
			if err != nil {
				t.Fatalf("Error reading test file %s: %v", rule, err)
			}

			_, err = ParseRules(testData, WithStrict(true))

			if err == nil {
				t.Fatalf("expected failure, got nil")
			}

		})
	}
}

// Return true if in a non error state
func checkParserError(t *testing.T, err error, wantErr error, wantPos int) bool {
	t.Helper()

	var (
		perr    ParseError
		hasPerr = errors.As(err, &perr)
	)

	if !errors.Is(err, wantErr) {
		t.Errorf("expected error '%v', got '%v'", wantErr, err)
		if hasPerr {
			t.Log(perr.Format(false, true))
		}
	}

	if wantErr == nil {
		return true
	}

	switch {
	case wantPos <= 0:
	case !hasPerr:
		t.Errorf("expected a parser error, got '%v'", err)
	case perr.Offset() != wantPos:
		t.Errorf("expected error at position %d, got %d", wantPos, perr.Offset())
		t.Log(perr.Format(false, true))
	}

	return false
}

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/goccy/go-yaml"
// )

// func TestParseRules(t *testing.T) {

// 	rule := `rules:
//   - cre:
//       id: set-1x1
//       tags:
//       - test1
//       - test2
//       - test3
//     metadata:

//       id: eeJwJiWQa9TyH3qTYYSZM9
//       hash: 9GJSdx4smGJeJCdiw6tiK59GJSdx4smGJeJCdiw6tiK5
//     rule:
//       set:
//         event:
//           source: nginx.access.log
//           origin: true
//         match:
//           - value: "test"
//         negate:
//           - value: "fart"
//             anchor: 0
// `

// 	r, err := ParseRules([]byte(rule), WithStrict(true))
// 	if err != nil {
// 		t.Fatalf("ParseRules failed: %v", err)
// 	}

// 	//fmt.Printf("Parsed rules: %+v\n", r)
// 	fmt.Println(Draw(r[0]))
// }

// func TestParseRules2(t *testing.T) {

// 	rule := `rules:
//   - cre:
//       id: PREQUEL-2024-0006
//       severity: 2
//       title: Kafka Topic Operator Thread Blocked
//       category: message-queue-problem
//       author: Prequel
//       description: |
//         There is a known issue in the Strimzi Kafka Topic Operator where the operator thread can become blocked. This can cause the operator to stop processing events and can lead to a backlog of events. This can cause the operator to become unresponsive and can lead to liveness probe failures and restarts of the Strimzi Kafka Topic Operator.
//       cause: |
//         The Kafka Topic Operator is using a single thread to process events. This can cause the thread to become blocked if there are a large number of events to process.
//       impact: |
//         No Kafka topics will be created or updated.
//       tags:
//         - known-problem
//         - kafka
//         - strimzi
//       mitigation: |
//         - Add additional CPU resources and restart the Kafka Topic Operator
//         - Use the Zookeeper store instead of the Kafka Streams store for the Strimzi Kafka Topic Operator
//       mitigationScore: 2
//       impactScore: 8
//       references:
//         - https://github.com/strimzi/strimzi-kafka-operator/issues/6046
//       applications:
//         - name: "kafka"
//     metadata:
//       kind: prequel
//       id: 6KA9oesqvnZu7MDfvpkMXz
//       hash: 9GJSdx4smGJeJCdiw6tiK59GJSdx4smGJeJCdiw6tiK5
//       gen: 1
//     rule:
//       set:
//         window: 60s
//         match:
//           - set:
//               event:
//                 source: cre.kubernetes
//               match:
//                 # Matches the unhealthy event for the Kafka Topic Operator pod when the startup probe fails for the topic operator due to this problem
//                 - jq: '.reason == "Unhealthy" and .involvedObject.fieldPath == "spec.containers{topic-operator}"'
//           - sequence:
//               window: 10s
//               event:
//                 source: prequel.log.kafka.topic-operator
//                 # When this detection fires we blame the topic operator for the problem
//                 origin: true
//               order:
//                 # This regex matches the warning log message that indicates the single topic operator thread is blocked and prevents CRUD operations on Kafka topics
//                 - regex: "WARN(.+)vertx-blocked-thread-checker(.+)Thread(.+)has been blocked for (.+)ms"
//                 # This is the substring pattern confirming the thread is blocked
//                 - "io.vertx.core.VertxException: Thread blocked"
// `

// 	r, err := ParseRules([]byte(rule), WithStrict(true))
// 	if err != nil {
// 		t.Fatalf("ParseRules failed: %v", err)
// 	}

// 	//fmt.Printf("Parsed rules: %+v\n", r)
// 	v := Draw(r[0])
// 	fmt.Println(v)
// }

// func TestParseRules4(t *testing.T) {

// 	rule := `rules:
//   - cre:
//       id: set-1x1
//       tags:
//       - test1
//       - test2
//       - test3
//     metadata:

//       id: eeJwJiWQa9TyH3qTYYSZM9
//       hash: 9GJSdx4smGJeJCdiw6tiK59GJSdx4smGJeJCdiw6tiK5
//     rule:
//       set:
//         event:
//           source: nginx.access.log
//           origin: true
//         match:
//           - value: "test"
//         negate:
//           - value: "fart"
//             anchor: 0
// `

// 	r, err := ParseRules([]byte(rule), WithStrict(true))
// 	if err != nil {
// 		t.Fatalf("ParseRules failed: %v", err)
// 	}

// 	//fmt.Printf("Parsed rules: %+v\n", r)
// 	fmt.Println(Draw(r[0]))
// }

// func TestBadRule(t *testing.T) {
// 	rule := `
// rules:
//   - cre:
//       id: nested-example
//     metadata:
//       id: eeJwJiWQa9TyH3qTYYSZM9
//       hash: 9GJSdx4smGJeJCdiw6tiK59GJSdx4smGJeJCdiw6tiK5
//     rule:
//       sequence:
//         window: 30s
//         correlations:
//           - hostname
//         order:
//           - sequence:
//               window: 10s
//               event:
//                 source: cre.log.rabbitmq
//                 origin: true
//               order:
//                 - value: Discarding message
//                   count: 10
//                 - Mnesia overloaded
//               negate:
//                 - SIGTERM
//           - sequence:
//               window: 5s
//               correlations:
//                 - container_id
//               order:
//                 - sequence:
//                     window: 1s
//                     event:
//                       source: cre.log.nginx
//                     order:
//                       - error message
//                       - shutdown
//                 - set:
//                     event:
//                       source: cre.log.nginx
//                     match:
//                       - 90%
//                 - set:
//                     event:
//                       source: cre.prequel.k8s
//                     match:
//                       - field: "reason"
//                         value: "Killing"
//         negate:
//           - set:
//               event:
//                 source: cre.prequel.k8s
//               match:
//                 - field: "reason"
//                   value: "NodeShutdown"
// `

// 	tt, err := ParseRules([]byte(rule), WithStrict(true))
// 	if err != nil {
// 		t.Fatalf("ParseRules failed: %v", err)
// 	}

// 	fmt.Println(Draw(tt[0], WithColor()))

// }

// func TestProm(t *testing.T) {

// 	rule := `
// rules:
//   - cre:
//       id: TestSuccessSimplePromQL
//     metadata:
//       id: "J7uRQTGpGMyL1iFpssnBeS"
//       hash: "rdJLgqYgkEp8J7uRQTGpGMyL1iFpssnBeSjg8Qks1qiq"
//       gen: 1
//     rule:
//       set:
//         window: 50s
//         match:
//           - promql:
//               event:
//                 source: cre.metrics
//                 origin: true
//               expr: 'sum(rate(http_requests_total[5m])) by (service)'
//               interval: 10s
//           - set:
//               event:
//                 source: kafka
//               match:
//                 - regex: "io.vertx.core.VertxException: Thread blocked"
// `

// 	tt, err := ParseRules([]byte(rule), WithStrict(true))
// 	if err != nil {
// 		t.Fatalf("ParseRules failed: %v", err)
// 	}

// 	fmt.Println(Draw(tt[0], WithColor()))
// }

// func TestScript(t *testing.T) {

// 	rule := `
// rules:
//   - cre:
//       id: TestSuccessChildScript
//     metadata:
//       id: "J7uRQTGpGMyL1iFpssnBeS"
//       hash: "rdJLgqYgkEp8jg8QrdJLgqYgkEp8jg8Qks1qiqks1qiq"
//       gen: 1
//     rule:
//       sequence:
//         window: 30s
//         order:
//           - script:
//               code: "function process(ev) print(\"Processing...\") end"
//               input:
//                 sequence:
//                   window: 10s
//                   event:
//                     source: kafka
//                     origin: true
//                   order:
//                     - value: "term1"
//                     - value: "term2"
//           - set:
//               event:
//                 source: kafka
//               match:
//                 - value: "term2"
// `

// 	tt, err := ParseRules([]byte(rule), WithStrict(true))
// 	if err != nil {
// 		t.Fatalf("ParseRules failed: %v", err)
// 	}

// 	fmt.Println(Draw(tt[0], WithColor()))
// }

// // func TestScriptOld(t *testing.T) {

// // 	rule := `
// // rules:
// //   - cre:
// //       id: TestSuccessChildScript
// //     metadata:
// //       id: "J7uRQTGpGMyL1iFpssnBeS"
// //       hash: "rdJLgqYgkEp8jg8Qks1qiqrdJLgqYgkEp8jg8Qks1qiq"
// //       gen: 1
// //     rule:
// //       sequence:
// //         window: 30s
// //         order:
// //           - script:
// //               code: "function process(ev) print(\"Processing...\") end"
// //               input:
// //                 sequence:
// //                   window: 10s
// //                   event:
// //                     source: kafka
// //                     origin: true
// //                   order:
// //                     - value: "term1"
// //                     - value: "term2"
// //           - set:
// //               event:
// //                 source: kafka
// //               match:
// //                 - value: "term2"
// // `

// // 	tt, err := ast.Build([]byte(rule))
// // 	if err != nil {
// // 		t.Fatalf("ast.Build failed: %v", err)
// // 	}

// // 	ast.DrawTree(tt, "ast_test.dot")

// // 	data, err := os.ReadFile("ast_test.dot")
// // 	if err != nil {
// // 		t.Fatalf("os.ReadFile failed: %v", err)
// // 	}

// // 	fmt.Printf("AST DOT data:\n%s\n", string(data))
// // }

// func TestNeg1(t *testing.T) {

// 	rule, err := rewriteAnchor([]byte(TestSuccessNegateOptions2))
// 	if err != nil {
// 		t.Fatalf("rewriteAnchor failed: %v", err)
// 	}

// 	fmt.Println(string(rule))

// 	tt, err := ParseRules([]byte(rule), WithStrict(true))
// 	if err != nil {
// 		t.Fatalf("ParseRules failed: %v", err)
// 	}

// 	fmt.Println(Draw(tt[0], WithColor()))
// }

// func TestSuccessLogSetSingle3_(t *testing.T) {

// 	rule, err := rewriteAnchor([]byte(TestSuccessLogSetSingle3))
// 	if err != nil {
// 		t.Fatalf("rewriteAnchor failed: %v", err)
// 	}

// 	fmt.Println(string(rule))

// 	tt, err := ParseRules([]byte(rule), WithStrict(true))
// 	if err != nil {
// 		t.Fatalf("ParseRules failed: %v", err)
// 	}

// 	fmt.Println(Draw(tt[0], WithColor()))
// }

// // func TestNeg1Old(t *testing.T) {
// // 	rule := TestSuccessNegateOptions2Old

// // 	tt, err := ast.Build([]byte(rule))
// // 	if err != nil {
// // 		t.Fatalf("ast.Build failed: %v", err)
// // 	}

// // 	ast.DrawTree(tt, "ast_test.dot")

// // 	data, err := os.ReadFile("ast_test.dot")
// // 	if err != nil {
// // 		t.Fatalf("os.ReadFile failed: %v", err)
// // 	}

// // 	fmt.Printf("AST DOT data:\n%s\n", string(data))
// // }

// var TestSuccessNegateOptions2 = `
// term1: &ref_0
//   sequence:
//     window: 10s
//     event:
//       source: log
//       origin: true
//     order:
//       - value: Discarding message
//         count: 10
//       - Mnesia overloaded
//     negate:
//       - value: SIGTERM
//         anchor: 1
// term2: &ref_1
//   set:
//     event:
//       source: k8s
//     match:
//       - field: reason
//         value: Killing set
// term3: &ref_2
//   set:
//     event:
//       source: log
//     match:
//       - value: Killing neg
// rules:
//   - cre:
//       id: TestSuccessNegateOptions2
//     metadata:
//       id: J7uRQTGpGMyL1iFpssnBeS
//       hash: rdJLgqYgkEp8jg8Qks1qiqrdJLgqYgkEp8jg8Qks1qiq
//       gen: 1
//     rule:
//       sequence:
//         window: 30s
//         correlations:
//           - hostname
//         order:
//           - *ref_0
//           - *ref_1
//         negate:
//           - <<: *ref_2
//             window: 10s
//             slide: 1s
//             anchor: 0
//             absolute: true

// `

// func rewriteAnchor(data []byte) ([]byte, error) {

// 	m := make(map[string]any)

// 	err := yaml.Unmarshal(data, &m)
// 	if err != nil {
// 		return nil, fmt.Errorf("yaml.Unmarshal failed: %w", err)
// 	}

// 	// Remove any anchor keys
// 	for key := range m {
// 		if key != "rules" {
// 			delete(m, key)
// 		}
// 	}

// 	// Re-marshal the map to YAML
// 	outData, err := yaml.Marshal(m)
// 	if err != nil {
// 		return nil, fmt.Errorf("yaml.Marshal failed: %w", err)
// 	}

// 	return outData, nil

// }

// var TestSuccessNegateOptions2Old = `
// rules:
//   - cre:
//       id: TestSuccessNegateOptions2
//     metadata:
//       id: "J7uRQTGpGMyL1iFpssnBeS"
//       hash: "rdJLgqYgkEp8jg8Qks1qiq"
//       generation: 1
//     rule:
//       sequence:
//         window: 30s
//         correlations:
//           - hostname
//         negate:
//           - value: term3
//             window: 10s
//             slide: 1s
//             anchor: 0
//             abs: true
//         order:
//           - term1
//           - term2

// terms:
//   term1:
//     sequence:
//       window: 10s
//       event:
//         source: log
//         origin: true
//         image_url: "*rabbitmq*"
//       order:
//         - value: Discarding message
//           count: 10
//         - Mnesia overloaded
//       negate:
//         - SIGTERM
// 		  anchor: 1
//   term2:
//     set:
//       event:
//         source: k8s
//       match:
//       - field: "reason"
//         value: "Killing"
//   term3:
//     sequence:
//       window: 10s
//       event:
//         source: log
//       order:
//         - value: "Killing"
//         - value: "in the name of the king"
// `

// var TestSuccessLogSetSingle3 = `
// rules:
//   - cre:
//       id: cre-2024-006
//     metadata:
//       id: "J7uRQTGpGMyL1iFpssnBeS"
//       hash: rdJLgqYgkEp8jg8Qks1qiqrdJLgqYgkEp8jg8Qks1qiq
//       gen: 1
//     rule:
//       set:
//         event:
//           source: rabbitmq
//           origin: true
//         match:
//           - Discarding message
// `

// var TestSuccessNestedSequence5 = `
// rules:
//   - cre:
//       id: cre-2024-006
//     metadata:
//       id: "J7uRQTGpGMyL1iFpssnBeS"
//       hash: "rdJLgqYgkEp8jg8Qks1qiqrdJLgqYgkEp8jg8Qks1qiq"
//       gen: 1
//     rule:
//       sequence:
//         window: 30s
//         correlations:
//           - hostname
//         order:
//           - sequence:
//               window: 10s
//               event:
//                 source: rabbitmq
//                 origin: true
//               order:
//                 - value: Discarding message
//                   count: 10
//                 - Mnesia overloaded
//               negate:
//                 - SIGTERM
//           - sequence:
//               window: 5s
//               correlations:
//                 - containerId
//               order:
//                 - sequence:
//                     window: 1s
//                     event:
//                       source: nginx
//                     order:
//                       - error message
//                       - shutdown
//                 - set:
//                     event:
//                       source: nginx
//                     match:
//                       - 90%
//                 - set:
//                     event:
//                       source: k8s
//                     match:
//                       - field: "reason"
//                         value: "Killing"
//         negate:
//           - set:
//               event:
//                 source: k8s
//               match:
//                 - field: "reason"
//                   value: "NodeShutdown"

// `

// func TestSuccessNestedSequence5_(t *testing.T) {

// 	rule, err := rewriteAnchor([]byte(TestSuccessNestedSequence5))
// 	if err != nil {
// 		t.Fatalf("rewriteAnchor failed: %v", err)
// 	}

// 	tt, err := ParseRules([]byte(rule), WithStrict(true))
// 	if err != nil {
// 		t.Fatalf("ParseRules failed: %v", err)
// 	}

// 	fmt.Println(Draw(tt[0], WithColor()))
// }

// var TestSuccessNestedSequence6 = `
// rules:
//   - cre:
//       id: nested-example
//     metadata:
//       id: "J7uRQTGpGMyL1iFpssnBeS"
//       hash: "rdJLgqYgkEp8jg8Qks1qiqrdJLgqYgkEp8jg8Qks1qiq"
//       gen: 1
//     rule:
//       sequence:
//         window: 30s
//         correlations:
//           - hostname
//         order:
//           - sequence:
//               window: 10s
//               event:
//                 source: rabbitmq
//                 origin: true
//               order:
//                 - value: Discarding message
//                   count: 10
//                 - Mnesia overloaded
//               negate:
//                 - SIGTERM
//           - sequence:
//               window: 5s
//               correlations:
//                 - container_id
//               order:
//                 - sequence:
//                     window: 1s
//                     event:
//                       source: nginx
//                     order:
//                       - error message
//                       - shutdown
//                 - set:
//                     event:
//                       source: nginx
//                     match:
//                       - 90%
//                 - set:
//                     event:
//                       source: k8s
//                     match:
//                       - field: "reason"
//                         value: "Killing"
//           - sequence:
//               window: 5s
//               correlations:
//                 - container_id
//               order:
//                 - sequence:
//                     window: 1s
//                     event:
//                       source: nginx
//                     order:
//                       - error message
//                       - shutdown
//                 - set:
//                     event:
//                       source: nginx
//                     match:
//                       - 90%
//                 - set:
//                     event:
//                       source: k8s
//                     match:
//                       - field: "reason"
//                         value: "Killing"
//         negate:
//           - set:
//               event:
//                 source: k8s
//               match:
//                 - field: "reason"
//                   value: "NodeShutdown"
// `

// func BenchmarkParseRules(b *testing.B) {

// 	rule := TestSuccessNestedSequence6

// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_, err := ParseRules([]byte(rule), WithStrict(true))
// 		if err != nil {
// 			b.Fatalf("ParseRules failed: %v", err)
// 		}
// 	}
// }
