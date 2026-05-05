package compiler

import (
	"errors"
	"fmt"
	"sort"

	"github.com/prequel-dev/prequel-compiler/pkg/parser"
	"github.com/rs/zerolog/log"
)

type ObjsT []*ObjT

type ObjTypeT int

const (
	ObjTypeMatcher ObjTypeT = iota
	ObjTypeAssert
)

func (o ObjTypeT) String() string {
	switch o {
	case ObjTypeMatcher:
		return "matcher"
	case ObjTypeAssert:
		return "assert"
	default:
		return "unknown"
	}
}

type ObjT struct {
	Address       parser.AstNodeAddressT
	ParentAddress *parser.AstNodeAddressT
	Scope         parser.AstScopeT
	AbstractType  parser.AstNodeType
	ObjectType    ObjTypeT
	Event         parser.AstEventT
	Object        any
	Cb            CallbackT
}

type compilerOptsT struct {
	runtime RuntimeI
	plugins map[parser.AstScopeT]PluginI
}

type CompilerOptT func(*compilerOptsT)
type PluginI interface {
	Compile(runtime RuntimeI, node parser.AstNode) (ObjsT, error)
}

func WithRuntime(cb RuntimeI) CompilerOptT {
	return func(o *compilerOptsT) {
		o.runtime = cb
	}
}

func WithPlugin(scope parser.AstScopeT, plugin PluginI) CompilerOptT {
	return func(o *compilerOptsT) {
		o.plugins[scope] = plugin
	}
}

func parseOpts(opts []CompilerOptT) compilerOptsT {

	o := compilerOptsT{
		plugins: map[parser.AstScopeT]PluginI{parser.AstScopeNode: defaultPlugin},
		runtime: defaultRuntime,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

func Compile(data []byte, scope parser.AstScopeT, opts ...CompilerOptT) (ObjsT, error) {

	rules, err := parser.ParseRules(data)
	if err != nil {
		return nil, err
	}

	return CompileRules(rules, scope, opts...)
}

func CompileRule(rule parser.AstRuleT, scope parser.AstScopeT, opts ...CompilerOptT) (ObjsT, error) {
	o := parseOpts(opts)
	return compileRule(o, rule, scope)
}

func CompileRules(rules []parser.AstRuleT, scope parser.AstScopeT, opts ...CompilerOptT) (ObjsT, error) {
	o := parseOpts(opts)

	var (
		outObjs ObjsT
		errList []error
	)

	for _, rule := range rules {
		objs, err := compileRule(o, rule, scope)
		if err != nil {
			errList = append(errList, err)
		} else {
			outObjs = append(outObjs, objs...)
		}
	}

	return outObjs, errors.Join(errList...)
}

func compileRule(o compilerOptsT, rule parser.AstRuleT, scope parser.AstScopeT) (ObjsT, error) {

	var (
		outObjs ObjsT
	)

	compile := func(node parser.AstNode, _ *parser.AstNegateOptsT) error {

		if node.Scope() != scope {
			return nil
		}

		plugin, ok := o.plugins[scope]
		if !ok {
			log.Error().Str("scope", scope.String()).Msg("No plugin found")
			return fmt.Errorf("%w: %s", ErrUnsupportedScope, scope.String())
		}

		objs, err := plugin.Compile(o.runtime, node)
		if err != nil {
			return err
		}

		outObjs = append(outObjs, objs...)

		return nil
	}

	if err := rule.Walk(compile); err != nil {
		return nil, err
	}

	sortObjs(outObjs, parser.AstNodeTypeSeq)
	sortObjs(outObjs, parser.AstNodeTypeSet)

	for _, obj := range outObjs {
		log.Debug().
			Str("abstract_type", obj.AbstractType.String()).
			Str("abstract_address", obj.Address.String()).
			Str("object_type", obj.ObjectType.String()).
			Msg("Compiled object")
	}

	return outObjs, nil
}

func NewObj(node parser.AstNode, objType ObjTypeT) *ObjT {

	return &ObjT{
		Address:       node.Address(),
		ParentAddress: node.Parent(),
		Scope:         node.Scope(),
		AbstractType:  node.Type(),
		ObjectType:    objType,
	}
}

// Should we sort by object type?
func sortObjs(items []*ObjT, t parser.AstNodeType) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].AbstractType == t && items[j].AbstractType != t {
			return true
		}
		if items[j].AbstractType == t && items[i].AbstractType != t {
			return false
		}
		return false
	})
}
