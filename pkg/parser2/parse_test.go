package parser2

import (
	"testing"
)

const anchorYaml string = `
event: &nginx_event
  source: nginx.access.log
  origin: true
`

func TestParseRules(t *testing.T) {

	rule := `
rules:
  - cre:
      id: set-1x1
    metadata:
      id: eeJwJiWQa9TyH3qTYYSZM9
      hash: 9GJSdx4smGJeJCdiw6tiK5
    rule:
      set:
        event: *nginx_event
        match:
          - value: "test"
        negate:
          - value: "fart"
`

	_, err := ParseRules([]byte(rule), WithStrict(true), WithAnchorYAML([]byte(anchorYaml)))
	if err != nil {
		t.Fatalf("ParseRules failed: %v", err)
	}
}
