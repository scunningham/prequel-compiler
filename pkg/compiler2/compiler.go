package compiler2

import (
	"errors"
	"fmt"
	"sort"

	"github.com/prequel-dev/prequel-compiler/pkg/parser2"
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
	Address       parser2.AstNodeAddressT
	ParentAddress *parser2.AstNodeAddressT
	Scope         parser2.AstScopeT
	AbstractType  parser2.AstNodeType
	ObjectType    ObjTypeT
	Event         parser2.AstEventT
	Object        any
	Cb            CallbackT
}

type compilerOptsT struct {
	runtime RuntimeI
	plugins map[parser2.AstScopeT]PluginI
}

type CompilerOptT func(*compilerOptsT)
type PluginI interface {
	Compile(runtime RuntimeI, node parser2.AstNode) (ObjsT, error)
}

func WithRuntime(cb RuntimeI) CompilerOptT {
	return func(o *compilerOptsT) {
		o.runtime = cb
	}
}

func WithPlugin(scope parser2.AstScopeT, plugin PluginI) CompilerOptT {
	return func(o *compilerOptsT) {
		o.plugins[scope] = plugin
	}
}

func parseOpts(opts []CompilerOptT) compilerOptsT {

	o := compilerOptsT{
		plugins: map[parser2.AstScopeT]PluginI{parser2.AstScopeNode: defaultPlugin},
		runtime: defaultRuntime,
	}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

func Compile(data []byte, scope parser2.AstScopeT, opts ...CompilerOptT) (ObjsT, error) {

	rules, err := parser2.ParseRules(data)
	if err != nil {
		return nil, err
	}

	return CompileRules(rules, scope, opts...)
}

func CompileRule(rule parser2.AstRuleT, scope parser2.AstScopeT, opts ...CompilerOptT) (ObjsT, error) {
	o := parseOpts(opts)
	return compileRule(o, rule, scope)
}

func CompileRules(rules []parser2.AstRuleT, scope parser2.AstScopeT, opts ...CompilerOptT) (ObjsT, error) {
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

func compileRule(o compilerOptsT, rule parser2.AstRuleT, scope parser2.AstScopeT) (ObjsT, error) {

	var (
		outObjs ObjsT
	)

	compile := func(node parser2.AstNode, _ *parser2.AstNegateOptsT) error {

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

	sortObjs(outObjs, parser2.AstNodeTypeSeq)
	sortObjs(outObjs, parser2.AstNodeTypeSet)

	for _, obj := range outObjs {
		log.Debug().
			Str("abstract_type", obj.AbstractType.String()).
			Str("abstract_address", obj.Address.String()).
			Str("object_type", obj.ObjectType.String()).
			Msg("Compiled object")
	}

	return outObjs, nil
}

func NewObj(node parser2.AstNode, objType ObjTypeT) *ObjT {

	return &ObjT{
		Address:       node.Address(),
		ParentAddress: node.Parent(),
		Scope:         node.Scope(),
		AbstractType:  node.Type(),
		ObjectType:    objType,
	}
}

// Should we sort by object type?
func sortObjs(items []*ObjT, t parser2.AstNodeType) {
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
