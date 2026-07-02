package attestation_test

import (
	"testing"

	"github.com/h3ow3d/saw/internal/assessment"
	"github.com/h3ow3d/saw/internal/attestation"
)

func fullAssessment() *assessment.WorkbookAssessment {
	a := &assessment.WorkbookAssessment{
		Name:              "my-service",
		Repository:        "ghcr.io/org/my-service",
		Version:           "v1.0.0",
		Digest:            "sha256:abc123",
		Supplier:          "Acme Corp",
		Assessor:          "Jane Smith",
		Approver:          "Bob Jones",
		AssessmentDate:    "2026-07-02",
		Outcome:           "B",
		OverrideRationale: "Approved due to compensating control review",
		RequiredControls:  "Image must be signed",
		DecisionRationale: "Acceptable with controls",
	}
	a.CriteriaAssessments = make([]assessment.CriterionAssessment, len(assessment.Criteria))
	for i := range assessment.Criteria {
		a.CriteriaAssessments[i] = assessment.CriterionAssessment{
			Score:            i + 1,
			Finding:          "finding",
			EvidenceReviewed: "evidence",
			Confidence:       "High",
			Rationale:        "rationale",
		}
	}
	return a
}

func TestGenerateType(t *testing.T) {
	att := attestation.Generate(fullAssessment())
	if att.Type != "cdsc.suitability-attestation.v1" {
		t.Errorf("Type = %q, want %q", att.Type, "cdsc.suitability-attestation.v1")
	}
}

func TestGenerateSubject(t *testing.T) {
	a := fullAssessment()
	att := attestation.Generate(a)

	if att.Subject.Name != a.Name {
		t.Errorf("Subject.Name = %q, want %q", att.Subject.Name, a.Name)
	}
	if att.Subject.Repository != a.Repository {
		t.Errorf("Subject.Repository = %q, want %q", att.Subject.Repository, a.Repository)
	}
	if att.Subject.Version != a.Version {
		t.Errorf("Subject.Version = %q, want %q", att.Subject.Version, a.Version)
	}
	if att.Subject.Digest != a.Digest {
		t.Errorf("Subject.Digest = %q, want %q", att.Subject.Digest, a.Digest)
	}
	if att.Subject.Supplier != a.Supplier {
		t.Errorf("Subject.Supplier = %q, want %q", att.Subject.Supplier, a.Supplier)
	}
}

func TestGenerateAssessment(t *testing.T) {
	a := fullAssessment()
	att := attestation.Generate(a)

	if att.Assessment.Assessor != a.Assessor {
		t.Errorf("Assessment.Assessor = %q, want %q", att.Assessment.Assessor, a.Assessor)
	}
	if att.Assessment.Approver != a.Approver {
		t.Errorf("Assessment.Approver = %q, want %q", att.Assessment.Approver, a.Approver)
	}
	if att.Assessment.AssessmentDate != a.AssessmentDate {
		t.Errorf("Assessment.AssessmentDate = %q, want %q", att.Assessment.AssessmentDate, a.AssessmentDate)
	}
	if att.Assessment.Outcome != a.Outcome {
		t.Errorf("Assessment.Outcome = %q, want %q", att.Assessment.Outcome, a.Outcome)
	}
	if att.Assessment.OverrideRationale != a.OverrideRationale {
		t.Errorf("Assessment.OverrideRationale = %q, want %q", att.Assessment.OverrideRationale, a.OverrideRationale)
	}
	if att.Assessment.RequiredControls != a.RequiredControls {
		t.Errorf("Assessment.RequiredControls = %q, want %q", att.Assessment.RequiredControls, a.RequiredControls)
	}
	if att.Assessment.DecisionRationale != a.DecisionRationale {
		t.Errorf("Assessment.DecisionRationale = %q, want %q", att.Assessment.DecisionRationale, a.DecisionRationale)
	}
}

func TestGenerateCriteriaCount(t *testing.T) {
	att := attestation.Generate(fullAssessment())
	if got := len(att.Criteria); got != len(assessment.Criteria) {
		t.Errorf("len(Criteria) = %d, want %d", got, len(assessment.Criteria))
	}
}

func TestGenerateCriteriaNames(t *testing.T) {
	att := attestation.Generate(fullAssessment())
	for i, c := range att.Criteria {
		want := assessment.Criteria[i].Name
		if c.Name != want {
			t.Errorf("Criteria[%d].Name = %q, want %q", i, c.Name, want)
		}
	}
}

func TestGenerateCriteriaScores(t *testing.T) {
	a := fullAssessment()
	att := attestation.Generate(a)
	for i, c := range att.Criteria {
		want := a.CriteriaAssessments[i].Score
		if c.Score != want {
			t.Errorf("Criteria[%d].Score = %d, want %d", i, c.Score, want)
		}
	}
}

func TestGenerateCriteriaFields(t *testing.T) {
	a := fullAssessment()
	att := attestation.Generate(a)
	for i, c := range att.Criteria {
		ca := a.CriteriaAssessments[i]
		if c.Finding != ca.Finding {
			t.Errorf("Criteria[%d].Finding = %q, want %q", i, c.Finding, ca.Finding)
		}
		if c.EvidenceReviewed != ca.EvidenceReviewed {
			t.Errorf("Criteria[%d].EvidenceReviewed = %q, want %q", i, c.EvidenceReviewed, ca.EvidenceReviewed)
		}
		if c.Confidence != ca.Confidence {
			t.Errorf("Criteria[%d].Confidence = %q, want %q", i, c.Confidence, ca.Confidence)
		}
		if c.Rationale != ca.Rationale {
			t.Errorf("Criteria[%d].Rationale = %q, want %q", i, c.Rationale, ca.Rationale)
		}
	}
}

func TestGenerateEmptyAssessment(t *testing.T) {
	// Generate should not panic with an empty assessment.
	a := &assessment.WorkbookAssessment{}
	att := attestation.Generate(a)
	if att.Type != "cdsc.suitability-attestation.v1" {
		t.Errorf("unexpected type: %q", att.Type)
	}
	if len(att.Criteria) != len(assessment.Criteria) {
		t.Errorf("expected %d criteria, got %d", len(assessment.Criteria), len(att.Criteria))
	}
	for i, c := range att.Criteria {
		if c.Score != 0 {
			t.Errorf("Criteria[%d].Score = %d, want 0 for empty assessment", i, c.Score)
		}
	}
}
