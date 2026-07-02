// Package main is the entry point for the Suitability Assessment Workbook server.
package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/h3ow3d/saw/internal/assessment"
	"github.com/h3ow3d/saw/internal/attestation"
)

// workbookData is passed to the workbook HTML template.
type workbookData struct {
	Criteria       []assessment.Criterion
	OutcomeOptions []struct {
		Value string
		Label string
	}
}

// generateResult is passed to the inline generate response templates.
type generateResult struct {
	JSON     string
	Filename string
}

type errorResult struct {
	Errors []string
}

type matrixResult struct {
	Recommendation        assessment.SuitabilityRecommendation
	OverrideRationale     string
	AssuranceHeadings     []string
	Rows                  []matrixRow
	RecommendedOutcomeTag string
	SelectedOutcomeTag    string
}

type matrixRow struct {
	Key   string
	Label string
	Cells []matrixCell
}

type matrixCell struct {
	Key     string
	Outcome string
	Active  bool
}

var (
	workbookTmpl *template.Template
	downloadTmpl *template.Template
	errTmpl      *template.Template
	matrixTmpl   *template.Template
)

// downloadTemplate is returned on successful attestation generation. The
// JavaScript reads the base64-encoded JSON from a data attribute and triggers
// a browser file download — keeping the server stateless.
const downloadTemplate = `<div class="result result--success">
  <span class="result__icon">✓</span>
  <span class="result__message">Attestation generated and downloaded successfully.</span>
  <div id="attestation-payload" data-json="{{.JSON}}" data-filename="{{.Filename}}" hidden></div>
  <script>
    (function() {
      var el = document.getElementById('attestation-payload');
      var data = atob(el.getAttribute('data-json'));
      var filename = el.getAttribute('data-filename');
      var blob = new Blob([data], {type: 'application/json'});
      var url = URL.createObjectURL(blob);
      var a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      setTimeout(function() { URL.revokeObjectURL(url); }, 100);
    })();
  </script>
</div>`

// errorTemplate is returned when form validation fails.
const errorTemplate = `<div class="result result--error">
  <p class="result__message"><strong>Please correct the following before generating an attestation:</strong></p>
  <ul class="result__errors">
    {{range .Errors}}<li>{{.}}</li>{{end}}
  </ul>
</div>`

const matrixTemplate = `<div id="suitability-matrix"
  class="matrix-panel"
  hx-post="/suitability-matrix"
  hx-trigger="load, change from:input[name^='score_'], change from:#outcome"
  hx-include="#workbook-form"
  hx-swap="outerHTML">
  <div class="matrix-panel__header">
    <div>
      <h3 class="matrix-panel__title">CDSC Suitability Outcome Matrix</h3>
      <p class="matrix-panel__advisory">The recommendation is advisory only. The assessor must still select the final outcome.</p>
    </div>
  </div>

  <div class="matrix-summary">
    <div class="matrix-summary__item">
      <span class="matrix-summary__label">Risk Level</span>
      <strong class="matrix-summary__value">{{if .Recommendation.Ready}}{{.Recommendation.RiskClassification}} ({{.Recommendation.RiskLevel}}){{else}}Pending{{end}}</strong>
    </div>
    <div class="matrix-summary__item">
      <span class="matrix-summary__label">Assurance Level</span>
      <strong class="matrix-summary__value">{{if .Recommendation.Ready}}{{.Recommendation.AssuranceClassification}} ({{.Recommendation.AssuranceLevel}}){{else}}Pending{{end}}</strong>
    </div>
    <div class="matrix-summary__item">
      <span class="matrix-summary__label">Recommended Outcome</span>
      <strong class="matrix-summary__value">{{if .Recommendation.Ready}}{{.RecommendedOutcomeTag}}{{else}}Pending{{end}}</strong>
    </div>
  </div>

  <div class="matrix-table-wrap">
    <table class="matrix-table">
      <thead>
        <tr>
          <th scope="col">Risk \ Assurance</th>
          {{range .AssuranceHeadings}}<th scope="col">{{.}}</th>{{end}}
        </tr>
      </thead>
      <tbody>
        {{range .Rows}}
        <tr>
          <th scope="row">{{.Label}}</th>
          {{range .Cells}}
          <td class="{{if .Active}}matrix-table__cell--active{{end}}">
            <span class="matrix-table__outcome">Outcome {{.Outcome}}</span>
          </td>
          {{end}}
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>

  <div class="matrix-reasoning">
    <span class="matrix-reasoning__label">Reasoning</span>
    <p>{{.Recommendation.Reasoning}}</p>
  </div>

  <div class="field field--full">
    <label for="override_rationale">Override Rationale{{if .Recommendation.OverrideRequired}} <span class="field__required">required</span>{{end}}</label>
    <textarea id="override_rationale"
              name="override_rationale"
              rows="3"
              {{if .Recommendation.OverrideRequired}}required{{end}}
              placeholder="Explain why the final selected outcome differs from the advisory recommendation.">{{.OverrideRationale}}</textarea>
    <p class="field__hint">Only required when the final selected outcome differs from the recommendation.</p>
    {{if .Recommendation.OverrideRequired}}
    <p class="field__alert">{{.SelectedOutcomeTag}} differs from {{.RecommendedOutcomeTag}}. Provide an override rationale before generating an attestation.</p>
    {{end}}
  </div>
</div>`

func main() {
	var err error

	workbookTmpl, err = template.ParseFiles("templates/workbook.html")
	if err != nil {
		log.Fatalf("failed to parse workbook template: %v", err)
	}

	downloadTmpl, err = template.New("download").Parse(downloadTemplate)
	if err != nil {
		log.Fatalf("failed to parse download template: %v", err)
	}

	errTmpl, err = template.New("error").Parse(errorTemplate)
	if err != nil {
		log.Fatalf("failed to parse error template: %v", err)
	}

	matrixTmpl, err = template.New("matrix").Parse(matrixTemplate)
	if err != nil {
		log.Fatalf("failed to parse matrix template: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.HandleFunc("/generate", handleGenerate)
	mux.HandleFunc("/suitability-matrix", handleSuitabilityMatrix)
	mux.HandleFunc("/", handleWorkbook)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Suitability Assessment Workbook listening on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// handleWorkbook serves the assessment workbook page.
func handleWorkbook(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data := workbookData{
		Criteria:       assessment.Criteria,
		OutcomeOptions: assessment.OutcomeOptions,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := workbookTmpl.Execute(w, data); err != nil {
		log.Printf("template error: %v", err)
	}
}

func handleSuitabilityMatrix(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	a := buildAssessment(r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := matrixTmpl.Execute(w, newMatrixResult(a)); err != nil {
		log.Printf("matrix template render error: %v", err)
	}
}

// handleGenerate validates the submitted workbook form, builds the attestation,
// and returns either an error fragment or a download-trigger fragment for HTMX.
func handleGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form data", http.StatusBadRequest)
		return
	}

	a := buildAssessment(r)

	if errs := validate(a); len(errs) > 0 {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := errTmpl.Execute(w, errorResult{Errors: errs}); err != nil {
			log.Printf("error template render error: %v", err)
		}
		return
	}

	att := attestation.Generate(a)

	jsonBytes, err := json.MarshalIndent(att, "", "  ")
	if err != nil {
		http.Error(w, "failed to encode attestation", http.StatusInternalServerError)
		return
	}

	timestamp := time.Now().UTC().Format("20060102150405")
	filename := fmt.Sprintf("suitability-attestation-%s.json", timestamp)

	encoded := base64.StdEncoding.EncodeToString(jsonBytes)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := downloadTmpl.Execute(w, generateResult{JSON: encoded, Filename: filename}); err != nil {
		log.Printf("download template render error: %v", err)
	}
}

// buildAssessment constructs a WorkbookAssessment from the submitted HTTP form.
func buildAssessment(r *http.Request) *assessment.WorkbookAssessment {
	a := &assessment.WorkbookAssessment{
		Name:              strings.TrimSpace(r.FormValue("name")),
		Repository:        strings.TrimSpace(r.FormValue("repository")),
		Version:           strings.TrimSpace(r.FormValue("version")),
		Digest:            strings.TrimSpace(r.FormValue("digest")),
		Supplier:          strings.TrimSpace(r.FormValue("supplier")),
		Assessor:          strings.TrimSpace(r.FormValue("assessor")),
		Approver:          strings.TrimSpace(r.FormValue("approver")),
		AssessmentDate:    strings.TrimSpace(r.FormValue("assessment_date")),
		Outcome:           strings.TrimSpace(r.FormValue("outcome")),
		OverrideRationale: strings.TrimSpace(r.FormValue("override_rationale")),
		RequiredControls:  strings.TrimSpace(r.FormValue("required_controls")),
		DecisionRationale: strings.TrimSpace(r.FormValue("decision_rationale")),
	}

	a.CriteriaAssessments = make([]assessment.CriterionAssessment, len(assessment.Criteria))
	for i := range assessment.Criteria {
		scoreStr := r.FormValue(fmt.Sprintf("score_%d", i))
		score, _ := strconv.Atoi(scoreStr)
		a.CriteriaAssessments[i] = assessment.CriterionAssessment{
			Score:            score,
			Finding:          strings.TrimSpace(r.FormValue(fmt.Sprintf("finding_%d", i))),
			EvidenceReviewed: strings.TrimSpace(r.FormValue(fmt.Sprintf("evidence_%d", i))),
			Confidence:       strings.TrimSpace(r.FormValue(fmt.Sprintf("confidence_%d", i))),
			Rationale:        strings.TrimSpace(r.FormValue(fmt.Sprintf("rationale_%d", i))),
		}
	}

	return a
}

// validate checks required fields and returns a slice of human-readable error messages.
func validate(a *assessment.WorkbookAssessment) []string {
	var errs []string

	if a.Name == "" {
		errs = append(errs, "Artefact Name is required.")
	}
	if a.Assessor == "" {
		errs = append(errs, "Assessor is required.")
	}
	if a.AssessmentDate == "" {
		errs = append(errs, "Assessment Date is required.")
	}
	if a.Outcome == "" {
		errs = append(errs, "Suitability Outcome is required.")
	}
	rec := assessment.EvaluateSuitability(a)
	if rec.OverrideRequired && a.OverrideRationale == "" {
		errs = append(errs, "Override Rationale is required when the selected outcome differs from the advisory recommendation.")
	}

	for i, def := range assessment.Criteria {
		if i >= len(a.CriteriaAssessments) {
			errs = append(errs, fmt.Sprintf("%s: score is required.", def.Name))
			continue
		}
		ca := a.CriteriaAssessments[i]
		if ca.Score < 1 || ca.Score > 5 {
			errs = append(errs, fmt.Sprintf("%s: a score between 1 and 5 is required.", def.Name))
		}
		if ca.Confidence == "" {
			errs = append(errs, fmt.Sprintf("%s: confidence level is required.", def.Name))
		}
	}

	return errs
}

func newMatrixResult(a *assessment.WorkbookAssessment) matrixResult {
	rec := assessment.EvaluateSuitability(a)

	rows := []matrixRow{
		{
			Key:   "High",
			Label: "High Risk",
			Cells: []matrixCell{
				{Key: "Low", Outcome: "D"},
				{Key: "Medium", Outcome: "D"},
				{Key: "High", Outcome: "C"},
			},
		},
		{
			Key:   "Medium",
			Label: "Med Risk",
			Cells: []matrixCell{
				{Key: "Low", Outcome: "D"},
				{Key: "Medium", Outcome: "B"},
				{Key: "High", Outcome: "B"},
			},
		},
		{
			Key:   "Low",
			Label: "Low Risk",
			Cells: []matrixCell{
				{Key: "Low", Outcome: "D"},
				{Key: "Medium", Outcome: "B"},
				{Key: "High", Outcome: "A"},
			},
		},
	}

	for i := range rows {
		for j := range rows[i].Cells {
			rows[i].Cells[j].Active = rec.Ready &&
				rows[i].Key == rec.RiskClassification &&
				rows[i].Cells[j].Key == rec.AssuranceClassification
		}
	}

	return matrixResult{
		Recommendation:        rec,
		OverrideRationale:     a.OverrideRationale,
		AssuranceHeadings:     []string{"Low", "Med", "High"},
		Rows:                  rows,
		RecommendedOutcomeTag: outcomeTag(rec.RecommendedOutcome),
		SelectedOutcomeTag:    outcomeTag(rec.SelectedOutcome),
	}
}

func outcomeTag(outcome string) string {
	if outcome == "" {
		return "No outcome selected"
	}
	return "Outcome " + outcome
}
