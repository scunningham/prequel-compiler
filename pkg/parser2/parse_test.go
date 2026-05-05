package parser2

import (
	"fmt"
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/prequel-dev/prequel-compiler/pkg/ast"
)

func TestParseRules(t *testing.T) {

	rule := `rules:
  - cre:
      id: set-1x1
      tags:
      - test1
      - test2
      - test3
    metadata:

      id: eeJwJiWQa9TyH3qTYYSZM9
      hash: 9GJSdx4smGJeJCdiw6tiK59GJSdx4smGJeJCdiw6tiK5
    rule:
      set:
        event:
          source: nginx.access.log
          origin: true
        match:
          - value: "test"
        negate:
          - value: "fart"
            anchor: 1
`

	r, err := ParseRules([]byte(rule), WithStrict(true))
	if err != nil {
		t.Fatalf("ParseRules failed: %v", err)
	}

	//fmt.Printf("Parsed rules: %+v\n", r)
	fmt.Println(Draw(r[0]))
}

func TestParseRules2(t *testing.T) {

	rule := `rules:
  - cre:
      id: PREQUEL-2024-0006
      severity: 2
      title: Kafka Topic Operator Thread Blocked
      category: message-queue-problem
      author: Prequel
      description: |
        There is a known issue in the Strimzi Kafka Topic Operator where the operator thread can become blocked. This can cause the operator to stop processing events and can lead to a backlog of events. This can cause the operator to become unresponsive and can lead to liveness probe failures and restarts of the Strimzi Kafka Topic Operator.
      cause: |
        The Kafka Topic Operator is using a single thread to process events. This can cause the thread to become blocked if there are a large number of events to process.
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
    metadata:
      kind: prequel
      id: 6KA9oesqvnZu7MDfvpkMXz
      hash: 9GJSdx4smGJeJCdiw6tiK59GJSdx4smGJeJCdiw6tiK5
      gen: 1
    rule:
      set:
        window: 60s
        match:
          - set:
              event:
                source: cre.kubernetes
              match:
                # Matches the unhealthy event for the Kafka Topic Operator pod when the startup probe fails for the topic operator due to this problem
                - jq: '.reason == "Unhealthy" and .involvedObject.fieldPath == "spec.containers{topic-operator}"'
          - sequence:
              window: 10s
              event:
                source: prequel.log.kafka.topic-operator
                # When this detection fires we blame the topic operator for the problem
                origin: true
              order:
                # This regex matches the warning log message that indicates the single topic operator thread is blocked and prevents CRUD operations on Kafka topics
                - regex: "WARN(.+)vertx-blocked-thread-checker(.+)Thread(.+)has been blocked for (.+)ms"
                # This is the substring pattern confirming the thread is blocked
                - "io.vertx.core.VertxException: Thread blocked" 
`

	r, err := ParseRules([]byte(rule), WithStrict(true))
	if err != nil {
		t.Fatalf("ParseRules failed: %v", err)
	}

	//fmt.Printf("Parsed rules: %+v\n", r)
	v := Draw(r[0])
	fmt.Println(v)
}

func TestParseRules4(t *testing.T) {

	rule := `rules:
  - cre:
      id: set-1x1
      tags:
      - test1
      - test2
      - test3
    metadata:

      id: eeJwJiWQa9TyH3qTYYSZM9
      hash: 9GJSdx4smGJeJCdiw6tiK5
    rule:
      set:
        event:
          source: nginx.access.log
          origin: true
        match:
          - value: "test"
        negate:
          - value: "fart"
            anchor: 1
`

	r, err := ParseRules([]byte(rule), WithStrict(true))
	if err != nil {
		t.Fatalf("ParseRules failed: %v", err)
	}

	//fmt.Printf("Parsed rules: %+v\n", r)
	fmt.Println(Draw(r[0]))
}

func TestParseOther(t *testing.T) {

	rule := `
rules:
  - cre:
      id: nested-example
    metadata:
      id: eeJwJiWQa9TyH3qTYYSZM9
      hash: 9GJSdx4smGJeJCdiw6tiK5
    rule:
      sequence:
        window: 30s
        correlations:
          - hostname
        order:
          - sequence:
              window: 10s
              event:
                source: cre.log.rabbitmq
                origin: true
              order:
                - value: Discarding message
                  count: 10
                - Mnesia overloaded
              negate:
                - SIGTERM
          - sequence:
              window: 5s
              correlations:
                - container_id
              order:
                - sequence:
                    window: 1s
                    event:
                      source: cre.log.nginx
                    order:
                      - error message
                      - shutdown
                - set:
                    event:
                      source: cre.log.nginx
                    match:
                      - 90%
                - set:
                    event:
                      source: cre.prequel.k8s
                    match:
                      - field: "reason"
                        value: "Killing"
        negate:
          - set:
              event:
                source: cre.prequel.k8s
              match:
                - field: "reason"
                  value: "NodeShutdown"
`

	tt, err := ast.Build([]byte(rule))
	if err != nil {
		t.Fatalf("ast.Build failed: %v", err)
	}

	ast.DrawTree(tt, "ast_test.dot")

	data, err := os.ReadFile("ast_test.dot")
	if err != nil {
		t.Fatalf("os.ReadFile failed: %v", err)
	}

	fmt.Printf("AST DOT data:\n%s\n", string(data))
}

func TestBadRule(t *testing.T) {
	rule := `
rules:
  - cre:
      id: nested-example
    metadata:
      id: eeJwJiWQa9TyH3qTYYSZM9
      hash: 9GJSdx4smGJeJCdiw6tiK59GJSdx4smGJeJCdiw6tiK5
    rule:
      sequence:
        window: 30s
        correlations:
          - hostname
        order:
          - sequence:
              window: 10s
              event:
                source: cre.log.rabbitmq
                origin: true
              order:
                - value: Discarding message
                  count: 10
                - Mnesia overloaded
              negate:
                - SIGTERM
          - sequence:
              window: 5s
              correlations:
                - container_id
              order:
                - sequence:
                    window: 1s
                    event:
                      source: cre.log.nginx
                    order:
                      - error message
                      - shutdown
                - set:
                    event:
                      source: cre.log.nginx
                    match:
                      - 90%
                - set:
                    event:
                      source: cre.prequel.k8s
                    match:
                      - field: "reason"
                        value: "Killing"
        negate:
          - set:
              event:
                source: cre.prequel.k8s
              match:
                - field: "reason"
                  value: "NodeShutdown"
`

	tt, err := ParseRules([]byte(rule), WithStrict(true))
	if err != nil {
		t.Fatalf("ParseRules failed: %v", err)
	}

	fmt.Println(Draw(tt[0], WithColor()))

}

func TestProm(t *testing.T) {

	rule := `
rules:
  - cre:
      id: TestSuccessSimplePromQL
    metadata:
      id: "J7uRQTGpGMyL1iFpssnBeS"
      hash: "rdJLgqYgkEp8J7uRQTGpGMyL1iFpssnBeSjg8Qks1qiq"
      gen: 1
    rule:
      set:
        window: 50s
        match:
          - promql:
              event:
                source: cre.metrics
                origin: true
              expr: 'sum(rate(http_requests_total[5m])) by (service)'
              interval: 10s
          - set:
              event:
                source: kafka
              match:
                - regex: "io.vertx.core.VertxException: Thread blocked"
`

	tt, err := ParseRules([]byte(rule), WithStrict(true))
	if err != nil {
		t.Fatalf("ParseRules failed: %v", err)
	}

	fmt.Println(Draw(tt[0], WithColor()))
}

func TestPromOld(t *testing.T) {

	rule := `
rules:
  - cre:
      id: TestSuccessSimplePromQL
    metadata:
      id: "J7uRQTGpGMyL1iFpssnBeS"
      hash: "rdJLgqYgkEp8J7uRQTGpGMyL1iFpssnBeSjg8Qks1qiq"
      gen: 1
    rule:
      set:
        window: 50s
        match:
          - promql:
              event:
                source: cre.metrics
                origin: true
              expr: 'sum(rate(http_requests_total[5m])) by (service)'
              interval: 10s
          - set:
              event:
                source: kafka
              match:
                - regex: "io.vertx.core.VertxException: Thread blocked"
`

	tt, err := ast.Build([]byte(rule))
	if err != nil {
		t.Fatalf("ast.Build failed: %v", err)
	}

	ast.DrawTree(tt, "ast_test.dot")

	data, err := os.ReadFile("ast_test.dot")
	if err != nil {
		t.Fatalf("os.ReadFile failed: %v", err)
	}

	fmt.Printf("AST DOT data:\n%s\n", string(data))
}

func TestScript(t *testing.T) {

	rule := `
rules:
  - cre:
      id: TestSuccessChildScript
    metadata:
      id: "J7uRQTGpGMyL1iFpssnBeS"
      hash: "rdJLgqYgkEp8jg8QrdJLgqYgkEp8jg8Qks1qiqks1qiq"
      gen: 1
    rule:
      sequence:
        window: 30s
        order:
          - script:
              code: "function process(ev) print(\"Processing...\") end"
              input:
                sequence:
                  window: 10s
                  event:
                    source: kafka
                    origin: true
                  order:
                    - value: "term1"
                    - value: "term2"
          - set:
              event:
                source: kafka
              match:
                - value: "term2"
`

	tt, err := ParseRules([]byte(rule), WithStrict(true))
	if err != nil {
		t.Fatalf("ParseRules failed: %v", err)
	}

	fmt.Println(Draw(tt[0], WithColor()))
}

func TestScriptOld(t *testing.T) {

	rule := `
rules:
  - cre:
      id: TestSuccessChildScript
    metadata:
      id: "J7uRQTGpGMyL1iFpssnBeS"
      hash: "rdJLgqYgkEp8jg8Qks1qiqrdJLgqYgkEp8jg8Qks1qiq"
      gen: 1
    rule:
      sequence:
        window: 30s
        order:
          - script:
              code: "function process(ev) print(\"Processing...\") end"
              input:
                sequence:
                  window: 10s
                  event:
                    source: kafka
                    origin: true
                  order:
                    - value: "term1"
                    - value: "term2"
          - set:
              event:
                source: kafka
              match:
                - value: "term2"
`

	tt, err := ast.Build([]byte(rule))
	if err != nil {
		t.Fatalf("ast.Build failed: %v", err)
	}

	ast.DrawTree(tt, "ast_test.dot")

	data, err := os.ReadFile("ast_test.dot")
	if err != nil {
		t.Fatalf("os.ReadFile failed: %v", err)
	}

	fmt.Printf("AST DOT data:\n%s\n", string(data))
}

func TestNeg1(t *testing.T) {

	rule, err := rewriteAnchor([]byte(TestSuccessNegateOptions2))
	if err != nil {
		t.Fatalf("rewriteAnchor failed: %v", err)
	}

	fmt.Println(string(rule))

	tt, err := ParseRules([]byte(rule), WithStrict(true))
	if err != nil {
		t.Fatalf("ParseRules failed: %v", err)
	}

	fmt.Println(Draw(tt[0], WithColor()))
}

func TestNeg1Old(t *testing.T) {
	rule := TestSuccessNegateOptions2Old

	tt, err := ast.Build([]byte(rule))
	if err != nil {
		t.Fatalf("ast.Build failed: %v", err)
	}

	ast.DrawTree(tt, "ast_test.dot")

	data, err := os.ReadFile("ast_test.dot")
	if err != nil {
		t.Fatalf("os.ReadFile failed: %v", err)
	}

	fmt.Printf("AST DOT data:\n%s\n", string(data))
}

var TestSuccessNegateOptions2 = `
term1: &ref_0
  sequence:
    window: 10s
    event:
      source: log
      origin: true
    order:
      - value: Discarding message
        count: 10
      - Mnesia overloaded
    negate:
      - value: SIGTERM
        anchor: 1
term2: &ref_1
  set:
    event:
      source: k8s
    match:
      - field: reason
        value: Killing set
term3: &ref_2
  set:
    event:
      source: log
    match:
      - value: Killing neg
rules:
  - cre:
      id: TestSuccessNegateOptions2
    metadata:
      id: J7uRQTGpGMyL1iFpssnBeS
      hash: rdJLgqYgkEp8jg8Qks1qiqrdJLgqYgkEp8jg8Qks1qiq
      gen: 1
    rule:
      sequence:
        window: 30s
        correlations:
          - hostname
        order:
          - *ref_0
          - *ref_1
        negate:
          - <<: *ref_2
            window: 10s
            slide: 1s
            anchor: 0
            absolute: true

`

func rewriteAnchor(data []byte) ([]byte, error) {

	m := make(map[string]any)

	err := yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, fmt.Errorf("yaml.Unmarshal failed: %w", err)
	}

	// Remove any anchor keys
	for key := range m {
		if key != "rules" {
			delete(m, key)
		}
	}

	// Re-marshal the map to YAML
	outData, err := yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("yaml.Marshal failed: %w", err)
	}

	return outData, nil

}

var TestSuccessNegateOptions2Old = `
rules:
  - cre:
      id: TestSuccessNegateOptions2
    metadata:
      id: "J7uRQTGpGMyL1iFpssnBeS"
      hash: "rdJLgqYgkEp8jg8Qks1qiq"
      generation: 1
    rule:
      sequence:
        window: 30s
        correlations:
          - hostname
        negate:
          - value: term3
            window: 10s
            slide: 1s
            anchor: 0
            abs: true
        order:
          - term1
          - term2

terms:        
  term1:
    sequence:
      window: 10s
      event:
        source: log
        origin: true
        image_url: "*rabbitmq*"
      order:
        - value: Discarding message
          count: 10
        - Mnesia overloaded
      negate:
        - SIGTERM
		  anchor: 1
  term2:
    set:
      event:
        source: k8s
      match:
      - field: "reason"
        value: "Killing"
  term3:
    sequence:
      window: 10s
      event:
        source: log
      order:
        - value: "Killing"
        - value: "in the name of the king"	
`
