package parser2

import (
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/parser"
)

func ParseRules(yamlInput []byte, opts ...ParseOpt) ([]AstRuleT, error) {
	o := _parseOpts(opts)

	anchors, err := maybeAnchors(o)
	if err != nil {
		return nil, err
	}

	p := newParser(o.colorize, o.strict, anchors)
	return p.parse(yamlInput)
}

type parserT struct {
	strict   bool
	colorize bool
	anchors  map[string][]byte
}

func newParser(colorize, strict bool, anchors map[string][]byte) *parserT {
	return &parserT{
		strict:   strict,
		colorize: colorize,
		anchors:  anchors,
	}
}

func (p *parserT) parse(yamlInput []byte) ([]AstRuleT, error) {

	// First parse the YAML into a yaml AST ignoring comments.
	doc, err := parser.ParseBytes(yamlInput, 0)
	if err != nil {
		return nil, err
	}

	// A yaml file can contain multiple documents;
	// iterate across the documents and parse each one as a rule document.

	var rules []AstRuleT
	for _, d := range doc.Docs {
		nRules, err := p.parseDocument(d)
		if err != nil {
			return nil, err
		}
		rules = append(rules, nRules...)
	}

	return rules, nil
}

// A single YAML document should container a map with a single key "rules" that maps to a list of rules.
func (p *parserT) parseDocument(doc *ast.DocumentNode) ([]AstRuleT, error) {

	mapping, ok := doc.Body.(*ast.MappingNode)
	if !ok {
		err := fmt.Errorf("%w: expected document body to be a mapping, got %s", ErrUnexpectedType, doc.Body.Type())
		return nil, p.wrapError(doc.Body, err)
	}

	var rulesNode *ast.SequenceNode
	for _, v := range mapping.Values {

		key, ok := v.Key.(*ast.StringNode)
		switch {
		case !ok || key.Value != kwRules:
			if p.strict {
				err := fmt.Errorf("unexpected key '%s' in document body", key.Value)
				return nil, p.wrapError(v, err)
			}
		default:
			rulesNode, ok = v.Value.(*ast.SequenceNode)
			if !ok {
				err := fmt.Errorf("%w: expected '%s' key to map to a sequence, got %s", ErrUnexpectedType, kwRules, v.Value.Type())
				return nil, p.wrapError(v.Value, err)
			}
		}
	}

	if rulesNode == nil {
		err := fmt.Errorf("%w: %s", ErrMissingKey, kwRules)
		return nil, p.wrapError(mapping, err)
	}

	return p.parseRulesNode(rulesNode)
}

func (p *parserT) parseRulesNode(node *ast.SequenceNode) ([]AstRuleT, error) {

	var rules []AstRuleT

	for _, ruleNode := range node.Values {

		rule, err := p.parseRuleNode(ruleNode)
		if err != nil {
			return nil, err
		}

		rules = append(rules, *rule)
	}

	return rules, nil
}

// Expects layout of a single mapping node with keys 'metadata', 'cre', and 'rule'

func (p *parserT) parseRuleNode(node ast.Node) (*AstRuleT, error) {

	mapping, ok := node.(*ast.MappingNode)
	if !ok {
		err := fmt.Errorf("%w: expected rule body to be a mapping, got %s", ErrUnexpectedType, node.Type())
		return nil, p.wrapError(node, err)
	}

	var (
		err      error
		meta     *ParseMetadataT
		cre      *ParseCreT
		rootNode *AstNodeT
	)

	for _, v := range mapping.Values {

		key, ok := v.Key.(*ast.StringNode)
		switch {
		case !ok:
			if p.strict {
				err := fmt.Errorf("%w: %s", ErrUnexpectedType, v.Key.Type())
				return nil, p.wrapError(v.Key, err)
			}

		case key.Value == kwMetadata:
			if meta, err = p.parseMetadataNode(v.Value); err != nil {
				return nil, err
			}

		case key.Value == kwCre:
			if cre, err = p.parseCreNode(v.Value); err != nil {
				return nil, err
			}

		case key.Value == kwRule:
			if rootNode, err = p.parseRootNode(v.Value); err != nil {
				return nil, err
			}

		default:
			if p.strict {
				err := fmt.Errorf("%w: %s", ErrUnexpectedKey, key.Value)
				return nil, p.wrapError(v, err)
			}
		}
	}

	switch {
	case meta == nil:
		err := fmt.Errorf("%w: %s", ErrMissingKey, kwMetadata)
		return nil, p.wrapError(mapping, err)
	case rootNode == nil:
		err := fmt.Errorf("%w: %s", ErrMissingKey, kwRule)
		return nil, p.wrapError(mapping, err)
	}

	rule := &AstRuleT{
		Metadata: AstMetadataT{
			Metadata: *meta,
		},
		Root: rootNode,
	}

	if cre != nil {
		rule.Metadata.Cre = *cre
	}

	return rule, nil
}

func (p *parserT) parseMetadataNode(node ast.Node) (*ParseMetadataT, error) {
	var opts []yaml.DecodeOption
	if p.strict {
		opts = append(opts, yaml.Strict())
	}

	var meta ParseMetadataT
	if err := yaml.NodeToValue(node, &meta, opts...); err != nil {
		return nil, p.wrapError(nil, err)
	}

	// TODO: Add additional validation here

	return &meta, nil
}

func (p *parserT) parseCreNode(node ast.Node) (*ParseCreT, error) {

	var opts []yaml.DecodeOption
	if p.strict {
		opts = append(opts, yaml.Strict())
	}

	var cre ParseCreT
	if err := yaml.NodeToValue(node, &cre, opts...); err != nil {
		return nil, p.wrapError(nil, err)
	}

	// TODO: Add additional validation here

	return &cre, nil
}

// At the root, currently expecting:
// type ParseRuleDataT struct {
// 	Sequence *ParseSequenceT `yaml:"sequence,omitempty"`
// 	Set      *ParseSetT      `yaml:"set,omitempty"`
// }

func (p *parserT) parseRootNode(node ast.Node) (*AstNodeT, error) {

	mapping, ok := node.(*ast.MappingNode)
	if !ok {
		err := fmt.Errorf("%w: expected rule root to be a mapping, got %s", ErrUnexpectedType, node.Type())
		return nil, p.wrapError(node, err)
	}

	var (
		err     error
		setNode *AstNodeT
		seqNode *AstNodeT
	)

	for _, v := range mapping.Values {

		key, ok := v.Key.(*ast.StringNode)
		switch {
		case !ok:
			if p.strict {
				err := fmt.Errorf("%w: %s", ErrUnexpectedType, v.Key.Type())
				return nil, p.wrapError(v.Key, err)
			}

		case key.Value == kwSequence:

			if seqNode, err = p.parseSequenceNode(v.Value); err != nil {
				return nil, err
			}

		case key.Value == kwSet:
			if setNode, err = p.parseSetNode(v.Value); err != nil {
				return nil, err
			}

		default:
			if p.strict {
				err := fmt.Errorf("%w: %s", ErrUnexpectedKey, key.Value)
				return nil, p.wrapError(v, err)
			}
		}
	}

	var rootNode *AstNodeT

	// Check one or the other but not both are present
	switch {
	case setNode == nil && seqNode == nil:
		err = fmt.Errorf("%w: expected rule root to contain either '%s' or '%s' key", ErrMissingKey, kwSequence, kwSet)
		return nil, p.wrapError(mapping, err)
	case setNode != nil && seqNode != nil:
		err = fmt.Errorf("%w: rule root cannot contain both '%s' and '%s' keys", ErrUnexpectedKey, kwSequence, kwSet)
		return nil, p.wrapError(mapping, err)
	case seqNode != nil:
		rootNode = seqNode
	default:
		rootNode = setNode
	}

	return rootNode, nil
}

func (p *parserT) parseSequenceNode(node ast.Node) (sNode *AstNodeT, err error) {

	if node, err = p.maybeAlias(node); err != nil {
		return nil, err
	}

	return nil, nil
}

// type ParseSetT struct {
// 	Window       string       `yaml:"window,omitempty"`
// 	Correlations []string     `yaml:"correlations,omitempty"`
// 	Event        *ParseEventT `yaml:"event,omitempty"`
// 	Match        []ParseTermT `yaml:"match,omitempty"`
// 	Negate       []ParseTermT `yaml:"negate,omitempty"`
// }

func (p *parserT) parseSetNode(node ast.Node) (sNode *AstNodeT, err error) {

	if node, err = p.maybeAlias(node); err != nil {
		return nil, err
	}

	mapping, ok := node.(*ast.MappingNode)
	if !ok {
		err := fmt.Errorf("%w: expected set to be a mapping, got %s", ErrUnexpectedType, node.Type())
		return nil, p.wrapError(node, err)
	}

	var (
		matchNode *AstNodeT
		// negateNode *AstNodeT
		// eventNode  *AstNodeT
	)

	for _, v := range mapping.Values {

		key, ok := v.Key.(*ast.StringNode)
		switch {
		case !ok:
			if p.strict {
				err := fmt.Errorf("%w: %s", ErrUnexpectedType, v.Key.Type())
				return nil, p.wrapError(v.Key, err)
			}

		case key.Value == kwEvent:

			if _, err = p.parseEventNode(v.Value); err != nil {
				return nil, err
			}

		case key.Value == kwMatch:
			if matchNode, err = p.parseMatchNode(v.Value); err != nil {
				return nil, err
			}

		case key.Value == kwNegate:
			if _, err = p.parseNegateNode(v.Value); err != nil {
				return nil, err
			}

		default:
			if p.strict {
				err := fmt.Errorf("%w: %s", ErrUnexpectedKey, key.Value)
				return nil, p.wrapError(v, err)
			}
		}
	}

	if matchNode == nil {
		err := fmt.Errorf("%w: missing required key '%s' in set", ErrMissingKey, kwMatch)
		return nil, p.wrapError(mapping, err)
	}

	return nil, nil
}

type parseEventT struct {
	Source string `yaml:"source"`
	Origin bool   `yaml:"origin"`
}

func (p *parserT) parseEventNode(node ast.Node) (*parseEventT, error) {
	var err error
	if node, err = p.maybeAlias(node); err != nil {
		return nil, nil
	}

	var opts []yaml.DecodeOption
	if p.strict {
		opts = append(opts, yaml.Strict())
	}

	var event parseEventT
	if err := yaml.NodeToValue(node, &event, opts...); err != nil {
		return nil, err
	}

	return &event, nil
}

func (p *parserT) parseMatchNode(node ast.Node) (*AstNodeT, error) {
	return nil, nil
}

func (p *parserT) parseNegateNode(node ast.Node) (*AstNodeT, error) {
	return nil, nil
}

func (p *parserT) maybeAlias(node ast.Node) (ast.Node, error) {
	if node.Type() != ast.AliasType {
		return node, nil
	}

	var (
		aliasNode       = node.(*ast.AliasNode)
		anchorName      = aliasNode.Value.String()
		anchorBytes, ok = p.anchors[anchorName]
	)

	if !ok {
		err := fmt.Errorf("%w: %s", ErrUndefinedAnchor, anchorName)
		return nil, p.wrapError(node, err)
	}

	file, err := parser.ParseBytes(anchorBytes, 0)
	if err != nil {
		return nil, p.wrapError(nil, err)
	}

	if len(file.Docs) != 1 {
		return nil, p.wrapError(node, fmt.Errorf("unexpected number of documents in anchor YAML: expected 1, got %d", len(file.Docs)))
	}

	return file.Docs[0].Body, nil
}

type ParseError struct {
	Line   int
	Column int
	Offset int
	Msg    string
}

func (e ParseError) Error() string {
	return e.Msg
}

func (p *parserT) rewriteError(err error) error {

	if yErr, ok := err.(yaml.Error); ok {
		var (
			token = yErr.GetToken()
			pos   = token.Position
		)

		return ParseError{
			Line:   pos.Line,
			Column: pos.Column,
			Offset: pos.Offset,
			Msg:    yaml.FormatError(err, p.colorize, true),
		}
	}

	return err
}

func (p *parserT) wrapError(node ast.Node, err error) error {
	if node == nil {
		return p.rewriteError(err)
	}

	var (
		token = node.GetToken()
		pos   = token.Position
		msg   = yaml.FormatErrorWithToken("parse fail", token, p.colorize, true)
	)

	parseError := ParseError{
		Line:   pos.Line,
		Column: pos.Column,
		Offset: pos.Offset,
		Msg:    msg,
	}

	return fmt.Errorf("%w: %w", parseError, err)
}

func (p *parserT) wrapErrorPath(path string, node ast.Node, err error) error {

}

// func ParseRules(yamlInput []byte, flags ParseFlagsT) ([]AstNode, error) {

// 	var rules ParseRulesT

// 	anchorReader := strings.NewReader(anchorYaml)

// 	ropts := yaml.ReferenceReaders(anchorReader)

// 	decoder := yaml.NewDecoder(strings.NewReader(string(yamlInput)), ropts, yaml.Strict())

// 	//ctx := context.WithValue(context.Background(), yamlDecodeOptionsKey, ropts)

// 	if err := decoder.Decode(&rules); err != nil {
// 		return nil, err
// 	}

// 	// err := yaml.UnmarshalWithOptions(yamlInput, &rules, opts)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	return nil, nil
// }

// type ParseSequenceT struct {
// 	Window       string       `yaml:"window"`
// 	Correlations []string     `yaml:"correlations,omitempty"`
// 	Event        *ParseEventT `yaml:"event,omitempty"`
// 	Origin       bool         `yaml:"origin,omitempty"`
// 	Order        []ParseTermT `yaml:"order,omitempty"`
// 	Negate       []ParseTermT `yaml:"negate,omitempty"`
// }

// type ParseNegateOptsT struct {
// 	Window   string `yaml:"window,omitempty"`
// 	Slide    string `yaml:"slide,omitempty"`
// 	Anchor   uint32 `yaml:"anchor,omitempty"`
// 	Absolute bool   `yaml:"absolute,omitempty"`
// }

// type ParseSetT struct {
// 	Window       string       `yaml:"window,omitempty"`
// 	Correlations []string     `yaml:"correlations,omitempty"`
// 	Event        *ParseEventT `yaml:"event,omitempty"`
// 	Match        []ParseTermT `yaml:"match,omitempty"`
// 	Negate       []ParseTermT `yaml:"negate,omitempty"`
// }

// type ParseExtractT struct {
// 	Name       string `yaml:"name"`
// 	JqValue    string `yaml:"jq,omitempty"`
// 	RegexValue string `yaml:"regex,omitempty"`
// }

// type ParsePromQL struct {
// 	Expr     string       `yaml:"expr"`
// 	Interval string       `yaml:"interval,omitempty"`
// 	For      string       `yaml:"for,omitempty"`
// 	Event    *ParseEventT `yaml:"event,omitempty"`
// }

// type ParseScriptT struct {
// 	Code     string      `yaml:"code"`
// 	Language string      `yaml:"language,omitempty"` // Assumes 'lua' if empty
// 	Timeout  string      `yaml:"timeout,omitempty"`  // Uses default if empty; expects duration string
// 	Input    *ParseTermT `yaml:"input"`              // Required input
// }

// type ParseEventT struct {
// 	Source string `yaml:"source"`
// 	Origin bool   `yaml:"origin,omitempty" json:"origin,omitempty"`
// }

// type InnerParseTermT struct {
// 	Field      string            `yaml:"field,omitempty"`
// 	StrValue   string            `yaml:"value,omitempty"`
// 	JqValue    string            `yaml:"jq,omitempty"`
// 	RegexValue string            `yaml:"regex,omitempty"`
// 	Count      int               `yaml:"count,omitempty"`
// 	Set        *ParseSetT        `yaml:"set,omitempty"`
// 	Sequence   *ParseSequenceT   `yaml:"sequence,omitempty"`
// 	NegateOpts *ParseNegateOptsT `yaml:",inline,omitempty"`
// 	PromQL     *ParsePromQL      `yaml:"promql,omitempty"`
// 	Script     *ParseScriptT     `yaml:"script,omitempty"`
// 	Extract    []ParseExtractT   `yaml:"extract,omitempty"`
// }

// type ParseTermT struct {
// 	InnerParseTermT `yaml:",inline"`
// }

// func (t *ParseTermT) UnmarshalYAML(node ast.Node) error {
// 	return yaml.NodeToValue(node, &t.InnerParseTermT)
// }

// type ParseFlagsT int

// const (
// 	FlagNone   ParseFlagsT = 0
// 	FlagStrict ParseFlagsT = 1 << iota
// )

// func (f ParseFlagsT) Strict() bool {
// 	return f&FlagStrict != 0
// }

// type NodeFlags int

// type KeyMeta struct {
// 	Line int
// 	Col  int
// }

// type RulesT struct {
// 	Rules []Annotated[RuleT] `yaml:"rules"`
// }

// type RuleT struct {
// 	Cre      Annotated[CreT]      `yaml:"cre"`
// 	Metadata Annotated[MetadataT] `yaml:"metadata"`
// 	Rule     Annotated[RuleDefT]  `yaml:"rule"`
// }

// type CreT struct {
// 	ID Annotated[string] `yaml:"id"`
// }

// type MetadataT struct {
// 	ID   Annotated[string] `yaml:"id"`
// 	Hash Annotated[string] `yaml:"hash"`
// }

// type RuleDefT struct {
// 	Set Annotated[SetDef] `yaml:"set"`
// }

// type SetDef struct {
// 	Event  Annotated[Event]       `yaml:"event"`
// 	Match  []Annotated[MatchItem] `yaml:"match"`
// 	Negate []Annotated[string]    `yaml:"negate"`
// }

// type Event struct {
// 	Source Annotated[string] `yaml:"source"`
// }

// type MatchItem struct {
// 	Value Annotated[string] `yaml:"value"`
// }

// type Annotated[T any] struct {
// 	Value T
// 	Line  int
// 	Col   int
// }

// type RulesT struct {
// 	Rules []Annotated[RuleT] `yaml:"rules"`
// }

// type RuleT struct {
// 	Cre      Annotated[CreT]      `yaml:"cre"`
// 	Metadata Annotated[MetadataT] `yaml:"metadata"`
// 	Rule     Annotated[RuleDefT]  `yaml:"rule"`
// }

// type CreT struct {
// 	ID Annotated[string] `yaml:"id"`
// }

// type MetadataT struct {
// 	ID   Annotated[string] `yaml:"id"`
// 	Hash Annotated[string] `yaml:"hash"`
// }

// type RuleDefT struct {
// 	Set Annotated[SetDef] `yaml:"set"`
// }

// type SetDef struct {
// 	Event  Annotated[Event]       `yaml:"event"`
// 	Match  []Annotated[MatchItem] `yaml:"match"`
// 	Negate []Annotated[string]    `yaml:"negate"`
// }

// type Event struct {
// 	Source Annotated[string] `yaml:"source"`
// }

// type MatchItem struct {
// 	Value Annotated[string] `yaml:"value"`
// }

// type NodeFlags int

// const (
// 	FlagRequired NodeFlags = 1 << iota
// 	FlagDeprecated
// )

// type KeyMeta struct {
// 	Name      string
// 	Line      int
// 	Col       int
// 	Flags     NodeFlags
// 	Anchor    string
// 	IsAlias   bool
// 	AliasLine int
// 	AliasCol  int
// }

// type Annotated[T any] struct {
// 	Value T
// 	Meta  KeyMeta
// }

// Capture metadata, do not decode value here
// func (a *Annotated[T]) UnmarshalYAML(node ast.Node) error {
// 	pos := node.GetToken().Position
// 	a.Line = pos.Line
// 	a.Col = pos.Column

// 	return nil
// }

// 	var rules RulesT

// 	const anchorYaml string = `
// event: &nginx_event
//   source: nginx.access.log
// `

// 	anchorReader := strings.NewReader(anchorYaml)

// 	opts := yaml.ReferenceReaders(anchorReader)

// 	err := yaml.UnmarshalWithOptions(yamlInput, &rules, opts)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return nil, nil

// func ParseRules(yamlInput []byte, flags ParseFlagsT) ([]AstNode, error) {

// 	var rules RulesT

// 	const anchorYaml string = `
// event: &nginx_event
//   source: nginx.access.log
// `

// 	anchorReader := strings.NewReader(anchorYaml)

// 	opts := yaml.ReferenceReaders(anchorReader)

// 	// Parse with external anchors
// 	doc, err := parser.ParseBytes(yamlInput, 0)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Decode fully once at top-level

// 	if err := yaml.NodeToValue(doc.Docs[0].Body, &rules, opts); err != nil {
// 		return nil, err
// 	}

// 	return nil, nil
// }

// func (a *Annotated[T]) UnmarshalYAML(node ast.Node) error {
// 	pos := node.GetToken().Position
// 	a.Meta.Line = pos.Line
// 	a.Meta.Col = pos.Column

// 	const anchorYaml string = `
// event: &nginx_event
//   source: nginx.access.log
// `

// 	anchorReader := strings.NewReader(anchorYaml)

// 	opts := yaml.ReferenceReaders(anchorReader)

// 	var v T
// 	if err := yaml.NodeToValue(node, &v, opts); err != nil {
// 		return fmt.Errorf("failed to decode value at %d:%d: %w", pos.Line, pos.Column, err)
// 	}
// 	a.Value = v
// 	return nil
// }

// const yamlDecodeOptionsKey = "yamlDecodeOptions"

// func (a *Annotated[T]) UnmarshalYAML(ctx context.Context, node ast.Node) error {
// 	// Capture position metadata
// 	pos := node.GetToken().Position
// 	a.Line = pos.Line
// 	a.Col = pos.Column

// 	var opts []yaml.DecodeOption

// 	if aopts := ctx.Value(yamlDecodeOptionsKey); aopts != nil {
// 		opts = append(opts, aopts.(yaml.DecodeOption))
// 	}

// 	anchorReader := strings.NewReader(anchorYaml)
// 	ropts := yaml.ReferenceReaders(anchorReader)

// 	// Decode underlying value
// 	var v T
// 	if err := yaml.NodeToValue(node, &v, ropts); err != nil {
// 		return err
// 	}

// 	a.Value = v
// 	return nil
// }

// func bindUnmarshal[T any](opts ...yaml.DecodeOption) func(*Annotated[T], []byte) error {
// 	return func(a *Annotated[T], data []byte) error {

// 		var v T
// 		if err := yaml.UnmarshalWithOptions(data, &v, opts...); err != nil {
// 			return err
// 		}
// 		a.Value = v
// 		return nil
// 		// // Capture position metadata
// 		// pos := node.GetToken().Position
// 		// a.Line = pos.Line
// 		// a.Col = pos.Column

// 		// // Decode underlying value
// 		// var v T
// 		// if err := yaml.NodeToValue(node, &v, opts...); err != nil {
// 		// 	return err
// 		// }

// 		// a.Value = v
// 		// return nil
// 	}

// }

// func ParseRules(yamlInput []byte, flags ParseFlagsT) ([]AstNode, error) {

// 	var rules ParseRulesT

// 	anchorReader := strings.NewReader(anchorYaml)

// 	ropts := yaml.ReferenceReaders(anchorReader)

// 	decoder := yaml.NewDecoder(strings.NewReader(string(yamlInput)), ropts, yaml.Strict())

// 	//ctx := context.WithValue(context.Background(), yamlDecodeOptionsKey, ropts)

// 	if err := decoder.Decode(&rules); err != nil {
// 		return nil, err
// 	}

// 	// err := yaml.UnmarshalWithOptions(yamlInput, &rules, opts)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	return nil, nil
// }

// func ParseRules(yamlData []byte, flags ParseFlagsT) ([]AstNode, error) {

// 	// Parse YAML first.
// 	file, err := parser.ParseBytes(yamlData, 0)

// 	if err != nil {
// 		return nil, err
// 	}

// 	// Walk through all documents; there is typically only one document,
// 	// but YAML allows multiple documents in a single file.
// 	var nodes []AstNode
// 	for i, doc := range file.Docs {

// 		log.Trace().Int("docIndex", i).Msg("Parsing document")
// 		nodes, err := parseYamlDocument(doc, flags)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to parse document #%d: %w", i+1, err)
// 		}
// 		nodes = append(nodes, nodes...)
// 	}

// 	return nodes, nil
// }

// // A single YAML document should container a map with a single key "rules" that maps to a list of rules.
// func parseYamlDocument(doc *ast.DocumentNode, flags ParseFlagsT) ([]AstNode, error) {
// 	if doc.Body.Type() != ast.MappingType {
// 		return nil, fmt.Errorf("expected document body to be a mapping, got %s", doc.Body.Type())
// 	}

// 	mapping := doc.Body.(*ast.MappingNode)

// 	var rulesNode *ast.SequenceNode
// 	for _, v := range mapping.Values {

// 		key, ok := v.Key.(*ast.StringNode)
// 		switch {
// 		case !ok:
// 			if flags.Strict() {
// 				return nil, wrapError(v.Key, fmt.Errorf("expected document body keys to be strings, got %s", v.Key.Type()))
// 			}
// 		case key.Value != kwRules:
// 			if flags.Strict() {
// 				return nil, wrapError(v, fmt.Errorf("unexpected key '%s' in document body", key.Value))
// 			}
// 		default:
// 			rulesNode, ok = v.Value.(*ast.SequenceNode)
// 			if !ok {
// 				return nil, wrapError(v.Value, fmt.Errorf("expected 'rules' key to map to a sequence, got %T", v.Value))
// 			}
// 		}
// 	}

// 	if rulesNode == nil {
// 		return nil, wrapError(mapping, fmt.Errorf("document body does not contain a '%s' key", kwRules))
// 	}

// 	return parseRulesNode(rulesNode, flags)
// }

// func parseRulesNode(node *ast.SequenceNode, flags ParseFlagsT) ([]AstNode, error) {
// 	return nil, nil

// 	// var rules []AstNode

// 	// for i, ruleNode := range node.Values {
// 	// 	log.Trace().Int("ruleIndex", i).Msg("Parsing rule")

// 	// 	rp := NewRuleParser(flags)

// 	// 	rule, err := rp.Parse(ruleNode)

// 	// 	if err != nil {
// 	// 		return nil, fmt.Errorf("failed to parse rule #%d: %w", i+1, err)
// 	// 	}
// 	// 	rules = append(rules, rule)
// 	// }

// 	// return rules, nil
// }

// type RuleParser struct {
// 	depth int
// 	flags ParseFlagsT
// }

// func NewRuleParser(flags ParseFlagsT) *RuleParser {
// 	return &RuleParser{
// 		flags: flags,
// 	}
// }

// // At the rule level, we support 'metadata', 'cre', and 'rule' keys

// func (rp *RuleParser) Parse(root ast.Node) (AstNode, error) {
// 	return nil, nil
// }

// type ParseError struct {
// 	Line   int
// 	Column int
// 	Offset int
// 	Msg    string
// }

// func (e ParseError) Error() string {
// 	return e.Msg
// }

// func wrapError(node ast.Node, err error) error {
// 	var (
// 		token = node.GetToken()
// 		pos   = token.Position
// 		msg   = yaml.FormatErrorWithToken("parse fail", token, false, true)
// 	)

// 	parseError := ParseError{
// 		Line:   pos.Line,
// 		Column: pos.Column,
// 		Offset: pos.Offset,
// 		Msg:    msg,
// 	}

// 	return fmt.Errorf("%w: %w", parseError, err)
// }

// type ParseRulesT struct {
// 	Rules []ParseRuleT `yaml:"rules"`
// }

// type ParseRuleT struct {
// 	Metadata ParseRuleMetadataT `yaml:"metadata,omitempty" json:"metadata,omitempty"`
// 	Cre      ParseCreT          `yaml:"cre,omitempty" json:"cre,omitempty"`
// 	Rule     ParseRuleDataT     `yaml:"rule,omitempty" json:"rule,omitempty"`
// }

// type ParseRuleMetadataT struct {
// 	Name    string `yaml:"name,omitempty" json:"name,omitempty"`
// 	Id      string `yaml:"id,omitempty" json:"id,omitempty"`
// 	Hash    string `yaml:"hash,omitempty" json:"hash,omitempty"`
// 	Gen     uint   `yaml:"generation" json:"generation"`
// 	Kind    string `yaml:"kind,omitempty" json:"kind,omitempty"`
// 	Version string `yaml:"version,omitempty" json:"version,omitempty"`
// }

// type ParseRuleDataT struct {
// 	Sequence *ParseSequenceT `yaml:"sequence,omitempty"`
// 	Set      *ParseSetT      `yaml:"set,omitempty"`
// }

// type ParseApplicationT struct {
// 	Name          string `yaml:"name,omitempty" json:"name,omitempty"`
// 	ProcessName   string `yaml:"processName,omitempty" json:"process_name,omitempty"`
// 	ProcessPath   string `yaml:"processPath,omitempty" json:"process_path,omitempty"`
// 	ContainerName string `yaml:"containerName,omitempty" json:"container_name,omitempty"`
// 	ImageUrl      string `yaml:"imageUrl,omitempty" json:"image_url,omitempty"`
// 	RepoUrl       string `yaml:"repoUrl,omitempty" json:"repo_url,omitempty"`
// 	Version       string `yaml:"version,omitempty" json:"version,omitempty"`
// }

// const (
// 	SeverityCritical = 0
// 	SeverityHigh     = 1
// 	SeverityMedium   = 2
// 	SeverityLow      = 3
// 	SeverityInfo     = 4
// )

// type ParseCreT struct {
// 	Id              string              `yaml:"id,omitempty" json:"id,omitempty"`
// 	Severity        uint                `yaml:"severity" json:"severity"`
// 	Title           string              `yaml:"title,omitempty" json:"title,omitempty"`
// 	Category        string              `yaml:"category,omitempty" json:"category,omitempty"`
// 	Tags            []string            `yaml:"tags,omitempty" json:"tags,omitempty"`
// 	Author          string              `yaml:"author,omitempty" json:"author,omitempty"`
// 	Description     string              `yaml:"description,omitempty" json:"description,omitempty"`
// 	Impact          string              `yaml:"impact,omitempty" json:"impact,omitempty"`
// 	ImpactScore     uint                `yaml:"impactScore,omitempty" json:"impact_score,omitempty"`
// 	Cause           string              `yaml:"cause,omitempty" json:"cause,omitempty"`
// 	Mitigation      string              `yaml:"mitigation,omitempty" json:"mitigation,omitempty"`
// 	MitigationScore uint                `yaml:"mitigationScore,omitempty" json:"mitigation_score,omitempty"`
// 	References      []string            `yaml:"references,omitempty" json:"references,omitempty"`
// 	Reports         uint                `yaml:"reports,omitempty" json:"reports,omitempty"`
// 	Applications    []ParseApplicationT `yaml:"applications,omitempty" json:"applications,omitempty"`
// }

// type ParseSequenceT struct {
// 	Window       string       `yaml:"window"`
// 	Correlations []string     `yaml:"correlations,omitempty"`
// 	Event        *ParseEventT `yaml:"event,omitempty"`
// 	Origin       bool         `yaml:"origin,omitempty"`
// 	Order        []ParseTermT `yaml:"order,omitempty"`
// 	Negate       []ParseTermT `yaml:"negate,omitempty"`
// }

// type ParseNegateOptsT struct {
// 	Window   string `yaml:"window,omitempty"`
// 	Slide    string `yaml:"slide,omitempty"`
// 	Anchor   uint32 `yaml:"anchor,omitempty"`
// 	Absolute bool   `yaml:"absolute,omitempty"`
// }

// type ParseSetT struct {
// 	Window       string       `yaml:"window,omitempty"`
// 	Correlations []string     `yaml:"correlations,omitempty"`
// 	Event        *ParseEventT `yaml:"event,omitempty"`
// 	Match        []ParseTermT `yaml:"match,omitempty"`
// 	Negate       []ParseTermT `yaml:"negate,omitempty"`
// }

// type ParseExtractT struct {
// 	Name       string `yaml:"name"`
// 	JqValue    string `yaml:"jq,omitempty"`
// 	RegexValue string `yaml:"regex,omitempty"`
// }

// type ParsePromQL struct {
// 	Expr     string       `yaml:"expr"`
// 	Interval string       `yaml:"interval,omitempty"`
// 	For      string       `yaml:"for,omitempty"`
// 	Event    *ParseEventT `yaml:"event,omitempty"`
// }

// type ParseScriptT struct {
// 	Code     string      `yaml:"code"`
// 	Language string      `yaml:"language,omitempty"` // Assumes 'lua' if empty
// 	Timeout  string      `yaml:"timeout,omitempty"`  // Uses default if empty; expects duration string
// 	Input    *ParseTermT `yaml:"input"`              // Required input
// }

// type ParseEventT struct {
// 	Source string `yaml:"source"`
// 	Origin bool   `yaml:"origin,omitempty" json:"origin,omitempty"`
// }

// type InnerParseTermT struct {
// 	Field      string            `yaml:"field,omitempty"`
// 	StrValue   string            `yaml:"value,omitempty"`
// 	JqValue    string            `yaml:"jq,omitempty"`
// 	RegexValue string            `yaml:"regex,omitempty"`
// 	Count      int               `yaml:"count,omitempty"`
// 	Set        *ParseSetT        `yaml:"set,omitempty"`
// 	Sequence   *ParseSequenceT   `yaml:"sequence,omitempty"`
// 	NegateOpts *ParseNegateOptsT `yaml:",inline,omitempty"`
// 	PromQL     *ParsePromQL      `yaml:"promql,omitempty"`
// 	Script     *ParseScriptT     `yaml:"script,omitempty"`
// 	Extract    []ParseExtractT   `yaml:"extract,omitempty"`
// }

// type ParseTermT struct {
// 	InnerParseTermT `yaml:",inline"`
// }

// func (t *ParseTermT) UnmarshalYAML(node ast.Node) error {
// 	return yaml.NodeToValue(node, &t.InnerParseTermT)
// }
