package parser

import (
	"time"

	"github.com/prequel-dev/prequel-logmatch/pkg/match"
)

type protoNode struct {
	ty           AstNodeType
	window       time.Duration
	correlations []string
	event        *AstEventT
	terms        []*protoTerm
	negate       []*protoTerm
}

type protoTerm struct {
	leaf       *protoField
	child      AstNode
	promNode   *AstPromT
	negateOpts *AstNegateOptsT
}

type protoField struct {
	Field      string
	StrValue   string
	JqValue    string
	RegexValue string
	Count      uint64
	Extract    []AstExtractT
}

func protoTermsToAstFields(terms []*protoTerm) []AstFieldT {
	var fields []AstFieldT
	for _, term := range terms {
		if term.leaf != nil {
			fields = append(fields, term.leaf.ToField(term.negateOpts))
		}
	}
	return fields
}

func protoTermsToAstTerms(terms []*protoTerm) []AstTermT {
	var termsList []AstTermT
	for _, term := range terms {
		if term.child != nil {
			termsList = append(termsList, AstTermT{
				Term:       term.child,
				NegateOpts: term.negateOpts,
			})
		}
	}
	return termsList
}

// Validate the protoField and convert it to an AstField.
// This includes ensuring that exactly one of StrValue, JqValue, RegexValue, Count, or Extract is set, and that Field is set.
func (f *protoField) ToField(nOpts *AstNegateOptsT) AstFieldT {

	t := AstFieldT{
		Count:      f.Count,
		Field:      f.Field,
		Extracts:   f.Extract,
		NegateOpts: nOpts,
	}

	switch {
	case f.StrValue != "":
		t.TermValue = match.TermT{
			Type:  match.TermRaw,
			Value: f.StrValue,
		}

	case f.JqValue != "":
		t.TermValue = match.TermT{
			Type:  match.TermJqJson,
			Value: f.JqValue,
		}

	case f.RegexValue != "":
		t.TermValue = match.TermT{
			Type:  match.TermRegex,
			Value: f.RegexValue,
		}

	default:
		panic("invalid protoField: exactly one of StrValue, JqValue, or RegexValue must be set")
	}

	return t
}
