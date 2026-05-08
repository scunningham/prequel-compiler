package ast

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

// A protoTerm represents either a field term or a child node term in the proto representation of the rule.
type protoTerm struct {
	field      *protoField
	child      AstNode
	negateOpts *AstNegateOptsT
}

func (t protoTerm) count() uint64 {
	if t.field != nil {
		return t.field.Count
	}
	return 1
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
		if term.field != nil {
			fields = append(fields, term.field.ToField(term.negateOpts))
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

	if t.Count == 0 {
		t.Count = 1
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
