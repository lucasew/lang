package fr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

const (
	GrammalecteRuleID  = "GRAMMALECTE"
	grammalecteTimeout = 500 * time.Millisecond
	grammalecteDownMS  = 5000
)

// GrammalecteIgnoreRules ports the static ignore set (subset).
var GrammalecteIgnoreRules = map[string]struct{}{
	"tab_fin_ligne": {}, "apostrophe_typographique": {},
	"typo_guillemets_typographiques_doubles_ouvrants": {},
	"nbsp_avant_double_ponctuation":                   {},
	"typo_guillemets_typographiques_doubles_fermants": {},
	"nbsp_avant_deux_points":                          {}, "apostrophe_typographique_après_t": {},
	"typo_point_collé_à_mot_suivant": {}, "typo_tiret_début_ligne": {},
}

// GrammalecteRule ports rules.fr.GrammalecteRule — queries a local Grammalecte HTTP server.
type GrammalecteRule struct {
	ID        string
	ServerURL string // e.g. http://localhost:8080
	Client    *http.Client
	// Post optional override for tests.
	Post func(text string) ([]GrammalecteMatch, error)

	mu          sync.Mutex
	lastErrorAt time.Time
}

// GrammalecteMatch is one error from the Grammalecte JSON API (simplified).
type GrammalecteMatch struct {
	RuleID      string   `json:"rule"`
	Start       int      `json:"start"`
	End         int      `json:"end"`
	Message     string   `json:"message"`
	Suggestions []string `json:"suggestions"`
}

func NewGrammalecteRule(serverURL string) *GrammalecteRule {
	return &GrammalecteRule{
		ID:        GrammalecteRuleID,
		ServerURL: serverURL,
		Client:    &http.Client{Timeout: grammalecteTimeout},
	}
}

func (r *GrammalecteRule) GetID() string { return r.ID }

// Match runs Grammalecte on the sentence text.
func (r *GrammalecteRule) Match(sentence *languagetool.AnalyzedSentence) ([]*rules.RuleMatch, error) {
	if r == nil || sentence == nil {
		return nil, nil
	}
	r.mu.Lock()
	if !r.lastErrorAt.IsZero() && time.Since(r.lastErrorAt) < grammalecteDownMS*time.Millisecond {
		r.mu.Unlock()
		return nil, nil
	}
	r.mu.Unlock()

	text := sentence.GetText()
	var (
		gms []GrammalecteMatch
		err error
	)
	if r.Post != nil {
		gms, err = r.Post(text)
	} else {
		gms, err = r.query(text)
	}
	if err != nil {
		r.mu.Lock()
		r.lastErrorAt = time.Now()
		r.mu.Unlock()
		return nil, nil // soft-fail like Java logger
	}
	var out []*rules.RuleMatch
	for _, g := range gms {
		if _, skip := GrammalecteIgnoreRules[g.RuleID]; skip {
			continue
		}
		m := rules.NewRuleMatch(r, sentence, g.Start, g.End, g.Message)
		if len(g.Suggestions) > 0 {
			m.SetSuggestedReplacements(g.Suggestions)
		}
		out = append(out, m)
	}
	return out, nil
}

func (r *GrammalecteRule) query(text string) ([]GrammalecteMatch, error) {
	if r.ServerURL == "" {
		return nil, fmt.Errorf("grammalecte server URL not set")
	}
	// common local API shape: POST form text=
	form := url.Values{"text": {text}}.Encode()
	req, err := http.NewRequest(http.MethodPost, strings.TrimRight(r.ServerURL, "/")+"/gc_text", strings.NewReader(form))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := r.Client
	if client == nil {
		client = &http.Client{Timeout: grammalecteTimeout}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ParseGrammalecteJSON(body)
}

// ParseGrammalecteJSON extracts matches from a Grammalecte-like JSON payload.
func ParseGrammalecteJSON(data []byte) ([]GrammalecteMatch, error) {
	// accept either {data:[{...}]} or {errors:[...]} or flat array
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		var arr []GrammalecteMatch
		if err2 := json.Unmarshal(data, &arr); err2 != nil {
			return nil, err
		}
		return arr, nil
	}
	for _, key := range []string{"data", "errors", "matches", "lGrammarErrors"} {
		if v, ok := raw[key]; ok {
			var arr []GrammalecteMatch
			if err := json.Unmarshal(v, &arr); err == nil && len(arr) > 0 {
				return arr, nil
			}
			// nested: data[0].lGrammarErrors
			var nested []map[string]json.RawMessage
			if err := json.Unmarshal(v, &nested); err == nil {
				var all []GrammalecteMatch
				for _, n := range nested {
					for _, k2 := range []string{"lGrammarErrors", "errors", "matches"} {
						if v2, ok := n[k2]; ok {
							var a []GrammalecteMatch
							if json.Unmarshal(v2, &a) == nil {
								all = append(all, a...)
							}
						}
					}
				}
				if len(all) > 0 {
					return all, nil
				}
			}
		}
	}
	return nil, nil
}
