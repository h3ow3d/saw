package main

import (
	"testing"

	"github.com/h3ow3d/saw/internal/assessment"
)

func TestValidateRequiresOverrideRationaleForRecommendationMismatch(t *testing.T) {
	a := fullWorkbookAssessment("B", "")

	errs := validate(a)
	if len(errs) == 0 {
		t.Fatal("expected validation errors")
	}

	found := false
	for _, err := range errs {
		if err == "Override Rationale is required when the selected outcome differs from the advisory recommendation." {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected override rationale error, got %v", errs)
	}
}

func TestValidateAllowsRecommendedOutcomeWithoutOverrideRationale(t *testing.T) {
	a := fullWorkbookAssessment("D", "")

	errs := validate(a)
	for _, err := range errs {
		if err == "Override Rationale is required when the selected outcome differs from the advisory recommendation." {
			t.Fatalf("unexpected override rationale error: %v", errs)
		}
	}
}

func fullWorkbookAssessment(outcome, overrideRationale string) *assessment.WorkbookAssessment {
	a := &assessment.WorkbookAssessment{
		Name:              "artefact",
		Assessor:          "assessor",
		AssessmentDate:    "2026-07-02",
		Outcome:           outcome,
		OverrideRationale: overrideRationale,
	}

	a.CriteriaAssessments = make([]assessment.CriterionAssessment, len(assessment.Criteria))
	for i := range a.CriteriaAssessments {
		a.CriteriaAssessments[i] = assessment.CriterionAssessment{
			Score:      3,
			Confidence: "High",
		}
	}

	a.CriteriaAssessments[0].Score = 4
	a.CriteriaAssessments[1].Score = 4
	a.CriteriaAssessments[2].Score = 3
	a.CriteriaAssessments[3].Score = 3
	a.CriteriaAssessments[4].Score = 3
	a.CriteriaAssessments[5].Score = 4
	a.CriteriaAssessments[6].Score = 4
	a.CriteriaAssessments[7].Score = 3

	return a
}
