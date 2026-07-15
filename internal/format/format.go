package format

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/lucasew/lang/internal/finding"
)

// Name is an output format.
type Name string

const (
	Text  Name = "text"
	JSON  Name = "json"
	SARIF Name = "sarif"
)

func Parse(s string) (Name, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "text", "":
		return Text, nil
	case "json":
		return JSON, nil
	case "sarif":
		return SARIF, nil
	default:
		return "", fmt.Errorf("unknown format %q (want text, json, sarif)", s)
	}
}

// Write emits findings in the chosen format.
func Write(w io.Writer, format Name, findings []finding.Finding) error {
	switch format {
	case Text:
		return writeText(w, findings)
	case JSON:
		return writeJSON(w, findings)
	case SARIF:
		return writeSARIF(w, findings)
	default:
		return fmt.Errorf("unsupported format %s", format)
	}
}

func writeText(w io.Writer, findings []finding.Finding) error {
	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	if _, err := fmt.Fprintln(tw, "location\tseverity\trule\tmessage\tsuggestion"); err != nil {
		return err
	}
	for _, f := range findings {
		loc := fmt.Sprintf("%s:%d:%d", f.File, f.Line, f.Column)
		sug := f.PrimarySuggestion()
		if _, err := fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", loc, f.Severity, f.Rule, f.Message, sug); err != nil {
			return err
		}
	}
	return tw.Flush()
}

func writeJSON(w io.Writer, findings []finding.Finding) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if findings == nil {
		findings = []finding.Finding{}
	}
	return enc.Encode(findings)
}

// Minimal SARIF 2.1.0 document.
func writeSARIF(w io.Writer, findings []finding.Finding) error {
	type region struct {
		StartLine   int `json:"startLine"`
		StartColumn int `json:"startColumn"`
		EndLine     int `json:"endLine,omitempty"`
		EndColumn   int `json:"endColumn,omitempty"`
	}
	type artifactLocation struct {
		URI string `json:"uri"`
	}
	type physicalLocation struct {
		ArtifactLocation artifactLocation `json:"artifactLocation"`
		Region           region           `json:"region"`
	}
	type location struct {
		PhysicalLocation physicalLocation `json:"physicalLocation"`
	}
	type fix struct {
		Description     map[string]string `json:"description"`
	}
	type result struct {
		RuleID    string            `json:"ruleId"`
		Level     string            `json:"level"`
		Message   map[string]string `json:"message"`
		Locations []location        `json:"locations"`
		Fixes     []fix            `json:"fixes,omitempty"`
	}
	type driver struct {
		Name           string `json:"name"`
		InformationURI string `json:"informationUri"`
		Version        string `json:"version"`
	}
	type tool struct {
		Driver driver `json:"driver"`
	}
	type run struct {
		Tool    tool     `json:"tool"`
		Results []result `json:"results"`
	}
	type doc struct {
		Schema  string `json:"$schema"`
		Version string `json:"version"`
		Runs    []run  `json:"runs"`
	}

	results := make([]result, 0, len(findings))
	for _, f := range findings {
		level := "warning"
		if f.Severity == "misspelling" {
			level = "error"
		}
		r := result{
			RuleID:  f.Rule,
			Level:   level,
			Message: map[string]string{"text": f.Message},
			Locations: []location{{
				PhysicalLocation: physicalLocation{
					ArtifactLocation: artifactLocation{URI: f.File},
					Region: region{
						StartLine:   f.Line,
						StartColumn: f.Column,
						EndLine:     f.EndLine,
						EndColumn:   f.EndColumn,
					},
				},
			}},
		}
		if s := f.PrimarySuggestion(); s != "" {
			r.Fixes = []fix{{Description: map[string]string{"text": s}}}
		}
		results = append(results, r)
	}

	out := doc{
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Version: "2.1.0",
		Runs: []run{{
			Tool: tool{Driver: driver{
				Name:           "lang",
				InformationURI: "https://github.com/lucasew/lang",
				Version:        "0.0.0-dev",
			}},
			Results: results,
		}},
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
