package ca

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CatalanRemoteRule ports org.languagetool.rules.ca.CatalanRemoteRule.
// Default-off unless CA_REMOTE_RULE_SERVER is set or PostFn is injected (tests).
type CatalanRemoteRule struct {
	ServerURLs              []string
	TimeoutMS               int
	MaxSentencesFirstServer int
	DefaultOff              bool
	// PostFn optional: maps sentences → corrected sentences (bypasses HTTP).
	PostFn func(sentences []string) ([]string, error)
	// Client optional HTTP client; Timeout applied when nil PostFn.
	Client *http.Client
}

func NewCatalanRemoteRule() *CatalanRemoteRule {
	r := &CatalanRemoteRule{
		TimeoutMS:               2000,
		MaxSentencesFirstServer: 4,
		DefaultOff:              true,
	}
	if v := tools.JavaStringTrim(os.Getenv("CA_REMOTE_RULE_SERVER")); v != "" {
		r.ServerURLs = strings.Split(v, ",")
		r.DefaultOff = false
	}
	if v := os.Getenv("CA_REMOTE_RULE_SERVER_TIMEOUT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			r.TimeoutMS = n
		}
	}
	if v := os.Getenv("MAX_SENTENCES_FIRST_SERVER"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			r.MaxSentencesFirstServer = n
		}
	}
	return r
}

func (r *CatalanRemoteRule) GetID() string { return "CA_REMOTE_RULE" }

func (r *CatalanRemoteRule) GetDescription() string {
	return "Recomanació del model d'aprenentatge automàtic."
}

func (r *CatalanRemoteRule) MinToCheckParagraph() int { return 0 }

var remoteTrimRE = regexp.MustCompile(`^[\s\x{00A0}\n]+|[\s\x{00A0}\n]+$`)

func trimAllSpaces(s string) string { return remoteTrimRE.ReplaceAllString(s, "") }

// MatchList runs the remote rewrite + DiffsAsMatches pipeline (text-level).
func (r *CatalanRemoteRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.DefaultOff {
		if r == nil || r.PostFn == nil {
			return nil
		}
	}
	if len(sentences) == 0 {
		return nil
	}
	plain := make([]string, len(sentences))
	for i, s := range sentences {
		if s != nil {
			plain[i] = s.GetText()
		}
	}
	corrected, err := r.fetchCorrections(plain)
	if err != nil || len(corrected) == 0 {
		return nil
	}
	var out []*rules.RuleMatch
	pos := 0
	diffs := tools.NewDiffsAsMatches()
	for i := 0; i < len(plain) && i < len(corrected); i++ {
		original := strings.ReplaceAll(plain[i], "\n", " ")
		corr := strings.ReplaceAll(corrected[i], "\n", " ")
		for _, pm := range diffs.GetPseudoMatches(original, corr) {
			if pm == nil || len(pm.GetReplacements()) == 0 {
				continue
			}
			suggestion := pm.GetReplacements()[0]
			if pm.GetToPos() > len(original) || pm.GetFromPos() < 0 || pm.GetToPos() < pm.GetFromPos() {
				continue
			}
			underlined := original[pm.GetFromPos():pm.GetToPos()]
			if (pm.GetToPos() == len(plain[i]) || pm.GetFromPos() == 0) && trimAllSpaces(underlined) == "" {
				continue
			}
			if trimAllSpaces(suggestion) == trimAllSpaces(underlined) {
				continue
			}
			// Java only keeps missing-comma suggestions for now.
			msg := "Canvi recomanat pel model d'aprenentatge automàtic."
			if suggestion == underlined+"," {
				msg = "Pot ser que hi falti una coma. Reviseu la puntuació."
			} else {
				continue
			}
			rm := rules.NewRuleMatch(r, sentences[i], pos+pm.GetFromPos(), pos+pm.GetToPos(), msg)
			rm.SetSuggestedReplacements(pm.GetReplacements())
			out = append(out, rm)
		}
		if sentences[i] != nil {
			pos += sentences[i].GetCorrectedTextLength()
		}
	}
	return out
}

func (r *CatalanRemoteRule) fetchCorrections(sentences []string) ([]string, error) {
	if r.PostFn != nil {
		return r.PostFn(sentences)
	}
	if len(r.ServerURLs) == 0 {
		return nil, nil
	}
	serverURL := r.ServerURLs[0]
	if len(r.ServerURLs) == 2 && len(sentences) > r.MaxSentencesFirstServer {
		serverURL = r.ServerURLs[1]
	}
	body, err := json.Marshal(map[string]any{"sentences": sentences})
	if err != nil {
		return nil, err
	}
	client := r.Client
	if client == nil {
		client = &http.Client{Timeout: time.Duration(r.TimeoutMS) * time.Millisecond}
	}
	req, err := http.NewRequest(http.MethodPost, serverURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, nil
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// Accept {"sentences":[...]} or a bare JSON array.
	var wrap struct {
		Sentences []string `json:"sentences"`
	}
	if err := json.Unmarshal(raw, &wrap); err == nil && wrap.Sentences != nil {
		return wrap.Sentences, nil
	}
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr, nil
	}
	return nil, nil
}
