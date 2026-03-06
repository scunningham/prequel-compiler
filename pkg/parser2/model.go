package parser2

type AstRuleT struct {
	Metadata AstMetadataT
	Root     *AstNodeT
}

type AstNodeT struct {
}

type AstMetadataT struct {
	Metadata ParseMetadataT
	Cre      ParseCreT
}

type ParseMetadataT struct {
	Name    string `yaml:"name,omitempty"`
	Id      string `yaml:"id,omitempty"`
	Hash    string `yaml:"hash,omitempty"`
	Gen     uint   `yaml:"generation"`
	Kind    string `yaml:"kind,omitempty"`
	Version string `yaml:"version,omitempty"`
}

type ParseCreT struct {
	Id              string              `yaml:"id,omitempty"`
	Severity        uint                `yaml:"severity"`
	Title           string              `yaml:"title,omitempty"`
	Category        string              `yaml:"category,omitempty"`
	Tags            []string            `yaml:"tags,omitempty"`
	Author          string              `yaml:"author,omitempty"`
	Description     string              `yaml:"description,omitempty"`
	Impact          string              `yaml:"impact,omitempty"`
	ImpactScore     uint                `yaml:"impactScore,omitempty"`
	Cause           string              `yaml:"cause,omitempty"`
	Mitigation      string              `yaml:"mitigation,omitempty"`
	MitigationScore uint                `yaml:"mitigationScore,omitempty"`
	References      []string            `yaml:"references,omitempty"`
	Reports         uint                `yaml:"reports,omitempty"`
	Applications    []ParseApplicationT `yaml:"applications,omitempty"`
}

type ParseApplicationT struct {
	Name          string `yaml:"name,omitempty"`
	ProcessName   string `yaml:"processName,omitempty"`
	ProcessPath   string `yaml:"processPath,omitempty"`
	ContainerName string `yaml:"containerName,omitempty"`
	ImageUrl      string `yaml:"imageUrl,omitempty"`
	RepoUrl       string `yaml:"repoUrl,omitempty"`
	Version       string `yaml:"version,omitempty"`
}

const (
	SeverityCritical = 0
	SeverityHigh     = 1
	SeverityMedium   = 2
	SeverityLow      = 3
	SeverityInfo     = 4
)

const (
	kwRules    = "rules"
	kwMetadata = "metadata"
	kwCre      = "cre"
	kwRule     = "rule"
	kwSequence = "sequence"
	kwSet      = "set"
	kwEvent    = "event"
	kwMatch    = "match"
	kwNegate   = "negate"
)
