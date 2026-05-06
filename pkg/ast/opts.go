package ast

type ParseOpt func(*optT)

type optT struct {
	strict          bool
	luaValidator    ValidatorFunc
	promQLValidator ValidatorFunc
}

func WithStrict(strict bool) ParseOpt {
	return func(opts *optT) {
		opts.strict = strict
	}
}

type ValidatorFunc func(string) error

func WithLuaValidator(validator ValidatorFunc) ParseOpt {
	return func(opts *optT) {
		opts.luaValidator = validator
	}
}

func WithPromQLValidator(validator ValidatorFunc) ParseOpt {
	return func(opts *optT) {
		opts.promQLValidator = validator
	}
}

var stubValidator = func(string) error { return nil }

func parseOpts(opts ...ParseOpt) optT {
	opt := optT{
		strict:          false,
		luaValidator:    stubValidator,
		promQLValidator: stubValidator,
	}
	for _, f := range opts {
		f(&opt)
	}
	return opt
}
