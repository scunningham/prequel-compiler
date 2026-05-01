package parser2

type AstNode interface {
	Addr() AstNodeAddressT
	Parent() *AstNodeAddressT
}

type AstNodeAddressT struct {
	Version  string  `json:"version"`   // Version of the address format
	Name     string  `json:"name"`      // Name of the node. Currently using type
	RuleHash string  `json:"rule_hash"` // unique semantic identifier for the rule
	Depth    uint32  `json:"depth"`     // Depth of the node in the rule tree
	NodeId   uint32  `json:"node_id"`   // globally unique identifier for the match in the rule tree
	TermIdx  *uint32 `json:"term_idx"`  // Index of term/condition into parent's conditions. Used for assertion to assign term idx into parent machines
}
