package assessment_test

import (
	"strings"
	"testing"

	"github.com/h3ow3d/saw/internal/assessment"
)

func TestCriteriaCount(t *testing.T) {
	if got := len(assessment.Criteria); got != 8 {
		t.Errorf("expected 8 criteria, got %d", got)
	}
}

func TestCriteriaNames(t *testing.T) {
	want := []string{
		"Sensitivity",
		"Privilege",
		"Provenance",
		"Verifiability",
		"Traceability",
		"Operational Impact",
		"Recoverability",
		"Supply Chain Assurance",
	}
	for i, c := range assessment.Criteria {
		if c.Name != want[i] {
			t.Errorf("criteria[%d].Name = %q, want %q", i, c.Name, want[i])
		}
	}
}

func TestCriteriaScoreDescriptions(t *testing.T) {
	for _, c := range assessment.Criteria {
		if len(c.ScoreDescriptions) != 5 {
			t.Errorf("criterion %q: expected 5 score descriptions, got %d",
				c.Name, len(c.ScoreDescriptions))
		}
		for j, sd := range c.ScoreDescriptions {
			if sd.Score != j+1 {
				t.Errorf("criterion %q: description[%d].Score = %d, want %d",
					c.Name, j, sd.Score, j+1)
			}
			if sd.Description == "" {
				t.Errorf("criterion %q: description[%d] is empty", c.Name, j)
			}
		}
	}
}

func TestCriteriaQuestionsNonEmpty(t *testing.T) {
	for _, c := range assessment.Criteria {
		if c.Question == "" {
			t.Errorf("criterion %q has an empty question", c.Name)
		}
	}
}

func TestOutcomeOptionsCount(t *testing.T) {
	if got := len(assessment.OutcomeOptions); got != 4 {
		t.Errorf("expected 4 outcome options, got %d", got)
	}
}

func TestOutcomeOptionValues(t *testing.T) {
	want := []string{"A", "B", "C", "D"}
	for i, o := range assessment.OutcomeOptions {
		if o.Value != want[i] {
			t.Errorf("OutcomeOptions[%d].Value = %q, want %q", i, o.Value, want[i])
		}
		if o.Label == "" {
			t.Errorf("OutcomeOptions[%d].Label is empty", i)
		}
	}
}

func TestEvaluateSuitability(t *testing.T) {
	tests := []struct {
		name              string
		scores            []int
		selectedOutcome   string
		wantRisk          string
		wantAssurance     string
		wantOutcome       string
		wantReady         bool
		wantOverride      bool
		wantMissing       []string
		wantReasoningPart string
	}{
		{
			name:              "low risk high assurance recommends A",
			scores:            []int{2, 1, 5, 4, 4, 2, 1, 5},
			selectedOutcome:   "A",
			wantRisk:          "Low",
			wantAssurance:     "High",
			wantOutcome:       "A",
			wantReady:         true,
			wantReasoningPart: "recommends Outcome A",
		},
		{
			name:              "medium risk medium assurance recommends B",
			scores:            []int{3, 2, 5, 3, 4, 2, 1, 4},
			selectedOutcome:   "B",
			wantRisk:          "Medium",
			wantAssurance:     "Medium",
			wantOutcome:       "B",
			wantReady:         true,
			wantReasoningPart: "recommends Outcome B",
		},
		{
			name:              "high risk high assurance recommends C",
			scores:            []int{4, 3, 5, 4, 5, 2, 1, 4},
			selectedOutcome:   "C",
			wantRisk:          "High",
			wantAssurance:     "High",
			wantOutcome:       "C",
			wantReady:         true,
			wantReasoningPart: "recommends Outcome C",
		},
		{
			name:              "high risk medium assurance recommends D and override",
			scores:            []int{5, 4, 5, 3, 4, 2, 1, 4},
			selectedOutcome:   "B",
			wantRisk:          "High",
			wantAssurance:     "Medium",
			wantOutcome:       "D",
			wantReady:         true,
			wantOverride:      true,
			wantReasoningPart: "recommends Outcome D",
		},
		{
			name:              "missing scores keeps recommendation pending",
			scores:            []int{5, 0, 5, 3, 0, 2, 1, 4},
			wantMissing:       []string{"Privilege", "Traceability"},
			wantReasoningPart: "Complete scores for Privilege, Traceability",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &assessment.WorkbookAssessment{Outcome: tt.selectedOutcome}
			a.CriteriaAssessments = make([]assessment.CriterionAssessment, len(tt.scores))
			for i, score := range tt.scores {
				a.CriteriaAssessments[i].Score = score
			}

			got := assessment.EvaluateSuitability(a)
			if got.Ready != tt.wantReady {
				t.Fatalf("Ready = %t, want %t", got.Ready, tt.wantReady)
			}
			if got.RiskClassification != tt.wantRisk {
				t.Errorf("RiskClassification = %q, want %q", got.RiskClassification, tt.wantRisk)
			}
			if got.AssuranceClassification != tt.wantAssurance {
				t.Errorf("AssuranceClassification = %q, want %q", got.AssuranceClassification, tt.wantAssurance)
			}
			if got.RecommendedOutcome != tt.wantOutcome {
				t.Errorf("RecommendedOutcome = %q, want %q", got.RecommendedOutcome, tt.wantOutcome)
			}
			if got.OverrideRequired != tt.wantOverride {
				t.Errorf("OverrideRequired = %t, want %t", got.OverrideRequired, tt.wantOverride)
			}
			if len(got.MissingCriteria) != len(tt.wantMissing) {
				t.Fatalf("MissingCriteria length = %d, want %d", len(got.MissingCriteria), len(tt.wantMissing))
			}
			for i, want := range tt.wantMissing {
				if got.MissingCriteria[i] != want {
					t.Errorf("MissingCriteria[%d] = %q, want %q", i, got.MissingCriteria[i], want)
				}
			}
			if !strings.Contains(got.Reasoning, tt.wantReasoningPart) {
				t.Errorf("Reasoning = %q, want substring %q", got.Reasoning, tt.wantReasoningPart)
			}
		})
	}
}
