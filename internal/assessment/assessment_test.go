package assessment_test

import (
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
