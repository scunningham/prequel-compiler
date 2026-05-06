package ast

type ValidatorFunc func(string) error

var stubFunc = func(string) error { return nil }

// Hooks exposed to avoid importing dependencies in compiler.
var (
	PromQLValidator ValidatorFunc = stubFunc // PromQLValidator validates a PromQL expression.
	LuaValidator    ValidatorFunc = stubFunc // LuaValidator validates Lua script syntax.
)
