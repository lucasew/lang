package commandline

import (
	"encoding/json"
	"path/filepath"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// MatchesAsSARIF builds a minimal SARIF 2.1.0 document for rule matches (SPEC §2.2).
// lang is used for SoftRuleURL helpUri (defaults to en when empty).
func MatchesAsSARIF(matches []*rules.RuleMatch, text, filename, lang string) string {
	if filename == "" {
		filename = "stdin"
	}
	// normalize to URI-ish path
	uri := filename
	if filename != "stdin" && filename != "-" {
		if abs, err := filepath.Abs(filename); err == nil {
			uri = abs
		}
	}

	type msg struct {
		Text string `json:"text"`
	}
	type region struct {
		StartLine   int `json:"startLine"`
		StartColumn int `json:"startColumn"`
		EndLine     int `json:"endLine,omitempty"`
		EndColumn   int `json:"endColumn,omitempty"`
	}
	type artifactLoc struct {
		URI string `json:"uri"`
	}
	type physLoc struct {
		ArtifactLocation artifactLoc `json:"artifactLocation"`
		Region           region      `json:"region"`
	}
	type location struct {
		PhysicalLocation physLoc `json:"physicalLocation"`
	}
	type props struct {
		Type string `json:"type,omitempty"`
	}
	type result struct {
		RuleID    string     `json:"ruleId"`
		Level     string     `json:"level"`
		Message   msg        `json:"message"`
		Locations []location `json:"locations"`
		Properties props     `json:"properties,omitempty"`
	}
	type reportingDesc struct {
		ID               string `json:"id"`
		Name             string `json:"name,omitempty"`
		ShortDescription *msg   `json:"shortDescription,omitempty"`
		FullDescription  *msg   `json:"fullDescription,omitempty"`
		HelpURI          string `json:"helpUri,omitempty"`
	}
	type driver struct {
		Name  string          `json:"name"`
		Rules []reportingDesc `json:"rules,omitempty"`
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

	ruleSeen := map[string]struct{}{}
	var ruleIndex []reportingDesc
	var results []result

	for _, m := range matches {
		if m == nil {
			continue
		}
		id := ruleIDOfMatch(m)
		_, _, issue, short := languagetool.SoftRuleMeta(id)
		sev := languagetool.SeverityFromIssueType(issue)
		desc := languagetool.SoftRuleDescription(id)
		if desc == "" {
			desc = id
		}
		if _, ok := ruleSeen[id]; !ok && id != "" {
			ruleSeen[id] = struct{}{}
			rd := reportingDesc{ID: id, Name: desc, HelpURI: languagetool.SoftRuleURL(id, lang)}
			if short != "" {
				rd.ShortDescription = &msg{Text: short}
			}
			rd.FullDescription = &msg{Text: desc}
			ruleIndex = append(ruleIndex, rd)
		}
		line, col := LineColumnAt(text, m.FromPos)
		endLine, endCol := LineColumnAt(text, m.ToPos)
		results = append(results, result{
			RuleID:  id,
			Level:   sev,
			Message: msg{Text: m.GetMessage()},
			Locations: []location{{
				PhysicalLocation: physLoc{
					ArtifactLocation: artifactLoc{URI: uri},
					Region: region{
						StartLine:   line,
						StartColumn: col,
						EndLine:     endLine,
						EndColumn:   endCol,
					},
				},
			}},
			Properties: props{Type: issue},
		})
	}
	if results == nil {
		results = []result{}
	}

	out := doc{
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Version: "2.1.0",
		Runs: []run{{
			Tool: tool{Driver: driver{
				Name:  "lang",
				Rules: ruleIndex,
			}},
			Results: results,
		}},
	}
	b, err := json.Marshal(out)
	if err != nil {
		return `{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"lang"}},"results":[]}]}`
	}
	return string(b)
}
