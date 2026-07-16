package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// NonSignificantVerbsRule is a surface stand-in for org.languagetool.rules.de.NonSignificantVerbsRule.
// Without lemmas/POS, common conjugations of haben/sein/machen/tun are matched by form.
// MinPercent 0 shows all matches (Java default is per-mill style filtering).
type NonSignificantVerbsRule struct {
	Messages   map[string]string
	MinPercent int
	// surface forms lowercased → lemma key
	forms map[string]string
}

func NewNonSignificantVerbsRule(messages map[string]string) *NonSignificantVerbsRule {
	// representative conjugations used in style checks / tests
	forms := map[string]string{
		// machen
		"mache": "machen", "machst": "machen", "macht": "machen", "machen": "machen",
		"machte": "machen", "machtest": "machen", "machten": "machen", "machtet": "machen",
		"gemacht": "machen",
		// tun
		"tue": "tun", "tu": "tun", "tust": "tun", "tut": "tun", "tun": "tun",
		"tat": "tun", "tatest": "tun", "taten": "tun", "tatet": "tun",
		"getan": "tun",
		// haben
		"habe": "haben", "hast": "haben", "hat": "haben", "haben": "haben",
		"hatte": "haben", "hattest": "haben", "hatten": "haben", "hattet": "haben",
		"gehabt": "haben",
		// sein (excluding "sein"/"Sein" via isException)
		"bin": "sein", "bist": "sein", "ist": "sein", "sind": "sein", "seid": "sein",
		"war": "sein", "warst": "sein", "waren": "sein", "wart": "sein",
		"gewesen": "sein",
	}
	return &NonSignificantVerbsRule{
		Messages:   messages,
		MinPercent: 0,
		forms:      forms,
	}
}

func (r *NonSignificantVerbsRule) GetID() string { return "NON_SIGNIFICANT_VERB_DE" }

func (r *NonSignificantVerbsRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r.MinPercent != 0 {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		lc := strings.ToLower(tok.GetToken())
		lemma, ok := r.forms[lc]
		if !ok {
			continue
		}
		if isNonSignificantException(tokens, i, lemma, tok.GetToken()) {
			continue
		}
		msg := "Dieses Verb hat wenig Aussagekraft. Verwenden Sie wenn möglich ein anderes oder formulieren Sie den Satz um."
		rm := rules.NewRuleMatch(r, sentence, tok.GetStartPos(), tok.GetEndPos(), msg)
		rm.ShortMessage = "wenig aussagekräftig"
		matches = append(matches, rm)
	}
	return matches
}

func isNonSignificantException(tokens []*languagetool.AnalyzedTokenReadings, num int, lemma, surface string) bool {
	// Java: tokens starting with sein/Sein are exceptions
	if strings.HasPrefix(surface, "sein") || strings.HasPrefix(surface, "Sein") {
		return true
	}
	if lemma == "machen" {
		for i := 1; i < len(tokens); i++ {
			s := tokens[i].GetToken()
			switch s {
			case "Angst", "Weg", "frisch", "bemerkbar", "aufmerksam":
				return true
			}
		}
	}
	if lemma == "haben" {
		for i := 1; i < len(tokens); i++ {
			s := tokens[i].GetToken()
			switch s {
			case "Glück", "Angst", "Mühe", "Recht", "recht":
				return true
			}
		}
	}
	// haben/sein with Flucht (PA2 / unknown-word exceptions need tagger)
	if lemma == "haben" || lemma == "sein" {
		for i := 1; i < len(tokens); i++ {
			if tokens[i].GetToken() == "Flucht" {
				return true
			}
		}
	}
	return false
}
