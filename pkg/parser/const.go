package parser

const (
	maxRank  = 2048
	maxDepth = 256
	maxGen   = 128

	// Root keys
	kwRules    = "rules"
	kwMetadata = "metadata"
	kwCre      = "cre"
	kwRule     = "rule"
	kwCompiler = "compiler"

	// Nodes
	kwSequence     = "sequence"
	kwSet          = "set"
	kwEvent        = "event"
	kwMatch        = "match"
	kwOrder        = "order"
	kwNegate       = "negate"
	kwCorrelations = "correlations"

	// Metadata
	kwName = "name"
	kwId   = "id"
	kwHash = "hash"
	kwGen  = "gen"
	kwKind = "kind"

	// CRE
	kwCreId           = "id"
	kwSeverity        = "severity"
	kwTitle           = "title"
	kwCategory        = "category"
	kwTags            = "tags"
	kwAuthor          = "author"
	kwDescription     = "description"
	kwImpact          = "impact"
	kwImpactScore     = "impactScore"
	kwCause           = "cause"
	kwMitigation      = "mitigation"
	kwMitigationScore = "mitigationScore"
	kwReferences      = "references"
	kwReports         = "reports"
	kwApplications    = "applications"

	// CRE Applications
	kwAppName        = "name"
	kwAppProcessName = "processName"
	kwAppProcessPath = "processPath"
	kwAppContainer   = "containerName"
	kwAppImage       = "imageUrl"
	kwAppRepo        = "repoUrl"
	kwAppVersion     = "version"

	// Event
	kwSource = "source"
	kwOrigin = "origin"

	// Extract
	kwExtractName  = "name"
	kwExtractJq    = "jq"
	kwExtractRegex = "regex"

	// Term
	kwField   = "field"
	kwValue   = "value"
	kwJq      = "jq"
	kwRegex   = "regex"
	kwCount   = "count"
	kwExtract = "extract"
	kwPromQL  = "promql"
	kwScript  = "script"

	// Negate opts
	kwWindow   = "window"
	kwSlide    = "slide"
	kwAnchor   = "anchor"
	kwAbsolute = "absolute"

	// PromQL
	kwPromExpr     = "expr"
	kwPromInterval = "interval"
	kwPromFor      = "for"
	kwPromEvent    = "event"

	// Script
	kwScriptCode    = "code"
	kwScriptLang    = "language"
	kwScriptTimeout = "timeout"
	kwScriptInput   = "input"
)
