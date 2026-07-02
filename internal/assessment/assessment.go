// Package assessment defines the suitability assessment domain model and the
// eight CDSC suitability criteria.
package assessment

// ScoreDescription maps a numeric score to its qualitative description.
type ScoreDescription struct {
	Score       int
	Description string
}

// Criterion defines a single suitability assessment criterion, including its
// question text and the five score descriptors used by an assessor.
type Criterion struct {
	Name              string
	Question          string
	ScoreDescriptions []ScoreDescription
}

// CriterionAssessment captures the assessor's inputs for a single criterion.
type CriterionAssessment struct {
	Score            int
	Finding          string
	EvidenceReviewed string
	Confidence       string
	Rationale        string
}

// WorkbookAssessment represents the complete artefact assessment entered by an
// assessor in the digital workbook.
type WorkbookAssessment struct {
	// Artefact information
	Name           string
	Repository     string
	Version        string
	Digest         string
	Supplier       string
	Assessor       string
	Approver       string
	AssessmentDate string

	// Per-criterion assessments, ordered to match Criteria.
	CriteriaAssessments []CriterionAssessment

	// Suitability outcome
	Outcome           string
	RequiredControls  string
	DecisionRationale string
}

// Criteria is the ordered list of eight CDSC suitability assessment criteria.
var Criteria = []Criterion{
	{
		Name:     "Sensitivity",
		Question: "To what extent does the artefact embody sensitive functionality, operational knowledge or mission-specific behaviour?",
		ScoreDescriptions: []ScoreDescription{
			{1, "Public functionality only"},
			{2, "Low sensitivity"},
			{3, "Operational procedures or infrastructure knowledge"},
			{4, "Sensitive operational capability"},
			{5, "Mission-sensitive capability"},
		},
	},
	{
		Name:     "Privilege",
		Question: "What authority, access or control will the artefact possess?",
		ScoreDescriptions: []ScoreDescription{
			{1, "User-level"},
			{2, "Service account"},
			{3, "Constrained administration"},
			{4, "System administration"},
			{5, "Cluster-admin or domain-admin"},
		},
	},
	{
		Name:     "Provenance",
		Question: "Can the origin, ownership and lifecycle of the artefact be established and verified?",
		ScoreDescriptions: []ScoreDescription{
			{1, "Unknown origin"},
			{2, "Supplier known"},
			{3, "Source repository known"},
			{4, "Build process documented"},
			{5, "Cryptographically verifiable provenance"},
		},
	},
	{
		Name:     "Verifiability",
		Question: "Can integrity and composition be independently validated?",
		ScoreDescriptions: []ScoreDescription{
			{1, "No verification"},
			{2, "Hashes only"},
			{3, "Signed artefact"},
			{4, "Signed artefact with SBOM"},
			{5, "Signed artefact with SBOM and provenance"},
		},
	},
	{
		Name:     "Traceability",
		Question: "Can assurance evidence, approvals and promotion activities be reconstructed?",
		ScoreDescriptions: []ScoreDescription{
			{1, "No traceability"},
			{2, "Release history"},
			{3, "Source and release traceability"},
			{4, "Source, build and approval traceability"},
			{5, "Complete assurance chain of custody"},
		},
	},
	{
		Name:     "Operational Impact",
		Question: "What could occur if the artefact were compromised, defective, unavailable or malicious?",
		ScoreDescriptions: []ScoreDescription{
			{1, "Negligible impact"},
			{2, "Limited disruption"},
			{3, "Significant degradation"},
			{4, "Major operational impact"},
			{5, "Mission failure or severe consequence"},
		},
	},
	{
		Name:     "Recoverability",
		Question: "Can the artefact be isolated, removed, replaced or rolled back?",
		ScoreDescriptions: []ScoreDescription{
			{1, "Immediate rollback available"},
			{2, "Recovery within hours"},
			{3, "Recovery within days"},
			{4, "Difficult recovery"},
			{5, "No practical recovery path"},
		},
	},
	{
		Name:     "Supply Chain Assurance",
		Question: "What confidence exists in the processes, dependencies, tooling and organisations involved in producing the artefact?",
		ScoreDescriptions: []ScoreDescription{
			{1, "No meaningful evidence"},
			{2, "Supplier assertion"},
			{3, "SBOM available"},
			{4, "SBOM and provenance available"},
			{5, "SBOM, provenance, signatures and assurance evidence available"},
		},
	},
}

// OutcomeOptions lists the valid suitability outcome values.
var OutcomeOptions = []struct {
	Value string
	Label string
}{
	{"A", "Outcome A — Suitable for Promotion"},
	{"B", "Outcome B — Suitable with Additional Controls"},
	{"C", "Outcome C — Hybrid Treatment Required"},
	{"D", "Outcome D — Higher-Assurance Treatment Required"},
}
