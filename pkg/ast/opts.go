package ast

type ParseOpt func(*parserT)

func WithStrict(strict bool) ParseOpt {
	return func(opts *parserT) {
		opts.strict = strict
	}
}

func applyOpts(parser *parserT, opts ...ParseOpt) {
	for _, opt := range opts {
		opt(parser)
	}
}
