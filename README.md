# saw — Suitability Assessment Workbook

A digital assessment workbook that supports the **Cross-Domain Supply Chain
Assurance (CDSC)** suitability assessment process.

An assessor uses the workbook to:

1. Describe an artefact (name, repository, version, digest, supplier).
2. Assess the artefact against eight defined suitability criteria.
3. Record evidence, findings, confidence levels and rationale per criterion.
4. Review the live CDSC suitability matrix recommendation and determine a final suitability outcome (A–D).
5. Generate and download a **CDSC Suitability Attestation** as JSON.

## Technology

- Go 1.24 · `net/http` · `html/template`
- [HTMX](https://htmx.org) (bundled — no CDN dependency)
- Pure HTML / CSS
- Docker (single container, stateless, no database)

## Project structure

```
cmd/server/main.go          HTTP server and handlers
internal/assessment/        Domain model and criteria definitions
internal/attestation/       Attestation model and generator
templates/workbook.html     Single-page workbook template
static/styles.css           Workbook styles
static/htmx.min.js          Bundled HTMX
Dockerfile
```

## Build and run

### Docker

```bash
docker build -t suitability-workbook .
docker run -p 8080:8080 suitability-workbook
```

Open <http://localhost:8080> in a browser.

### Local development

```bash
go run ./cmd/server
```

The server reads `templates/` and `static/` relative to the working directory,
so run the command from the repository root.

## Attestation format

Clicking **Generate Attestation** validates all required fields and downloads a
JSON file named `suitability-attestation-{timestamp}.json`:

```json
{
  "type": "cdsc.suitability-attestation.v1",
  "subject": {
    "name": "",
    "repository": "",
    "version": "",
    "digest": "",
    "supplier": ""
  },
  "assessment": {
    "assessor": "",
    "approver": "",
    "assessmentDate": "",
    "outcome": "",
    "overrideRationale": "",
    "decisionRationale": "",
    "requiredControls": ""
  },
  "criteria": [
    {
      "name": "",
      "score": 0,
      "finding": "",
      "confidence": "",
      "evidenceReviewed": "",
      "rationale": ""
    }
  ]
}
```

The workbook computes a live advisory recommendation from the Section 2 scores:

- **Risk Level** uses the highest of Sensitivity, Privilege, Operational Impact
  and Recoverability.
- **Assurance Level** uses the lowest of Provenance, Verifiability,
  Traceability and Supply Chain Assurance.
- If the final selected outcome differs from the advisory recommendation, an
  **Override Rationale** is required before attestation generation.

The attestation is intended for external signing and publication and is not
stored by this application.
