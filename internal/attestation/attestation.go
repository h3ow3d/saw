// Package attestation defines the CDSC Suitability Attestation model and
// the function that generates an attestation from a completed workbook assessment.
package attestation

import "github.com/h3ow3d/saw/internal/assessment"

const attestationType = "cdsc.suitability-attestation.v1"

// Subject identifies the artefact that was assessed.
type Subject struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Version    string `json:"version"`
	Digest     string `json:"digest"`
	Supplier   string `json:"supplier"`
}

// Assessment captures the outcome and meta-information about the assessment.
type Assessment struct {
	Assessor          string `json:"assessor"`
	Approver          string `json:"approver"`
	AssessmentDate    string `json:"assessmentDate"`
	Outcome           string `json:"outcome"`
	OverrideRationale string `json:"overrideRationale"`
	DecisionRationale string `json:"decisionRationale"`
	RequiredControls  string `json:"requiredControls"`
}

// Criterion records the assessor's evaluation of a single suitability criterion.
type Criterion struct {
	Name             string `json:"name"`
	Score            int    `json:"score"`
	Finding          string `json:"finding"`
	Confidence       string `json:"confidence"`
	EvidenceReviewed string `json:"evidenceReviewed"`
	Rationale        string `json:"rationale"`
}

// Attestation is the complete CDSC Suitability Attestation object.
type Attestation struct {
	Type       string     `json:"type"`
	Subject    Subject    `json:"subject"`
	Assessment Assessment `json:"assessment"`
	Criteria   []Criterion `json:"criteria"`
}

// Generate builds an Attestation from a completed WorkbookAssessment.
func Generate(a *assessment.WorkbookAssessment) Attestation {
	criteria := make([]Criterion, len(assessment.Criteria))
	for i, def := range assessment.Criteria {
		var ca assessment.CriterionAssessment
		if i < len(a.CriteriaAssessments) {
			ca = a.CriteriaAssessments[i]
		}
		criteria[i] = Criterion{
			Name:             def.Name,
			Score:            ca.Score,
			Finding:          ca.Finding,
			Confidence:       ca.Confidence,
			EvidenceReviewed: ca.EvidenceReviewed,
			Rationale:        ca.Rationale,
		}
	}

	return Attestation{
		Type: attestationType,
		Subject: Subject{
			Name:       a.Name,
			Repository: a.Repository,
			Version:    a.Version,
			Digest:     a.Digest,
			Supplier:   a.Supplier,
		},
		Assessment: Assessment{
			Assessor:          a.Assessor,
			Approver:          a.Approver,
			AssessmentDate:    a.AssessmentDate,
			Outcome:           a.Outcome,
			OverrideRationale: a.OverrideRationale,
			DecisionRationale: a.DecisionRationale,
			RequiredControls:  a.RequiredControls,
		},
		Criteria: criteria,
	}
}
