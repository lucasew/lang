package de

import (
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CompoundInfinitivRule ports org.languagetool.rules.de.CompoundInfinitivRule.
// Java: ZUS + "zu" + VER:INF and !isMisspelled(particle+infinitive) — no surface invent.
// Java: COMPOUNDING, Misspelling, setUrl.
type CompoundInfinitivRule struct {
	Messages     map[string]string
	Category     *rules.Category
	IssueType    rules.ITSIssueType
	IsMisspelled func(word string) bool // optional; nil → fail-closed misspelled (no false compounds)
}

func NewCompoundInfinitivRule(messages map[string]string) *CompoundInfinitivRule {
	return &CompoundInfinitivRule{
		Messages:  messages,
		Category:  rules.CatCompounding.GetCategory(messages),
		IssueType: rules.ITSMisspelling,
	}
}

func (r *CompoundInfinitivRule) GetID() string { return "COMPOUND_INFINITIV_RULE" }

func (r *CompoundInfinitivRule) GetDescription() string {
	return "Erweiterter Infinitiv mit zu (Zusammenschreibung)"
}

// GetURL ports CompoundInfinitivRule constructor setUrl.
func (r *CompoundInfinitivRule) GetURL() string {
	return "https://languagetool.org/insights/de/beitrag/zu-zusammen-oder-getrennt/"
}

func (r *CompoundInfinitivRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *CompoundInfinitivRule) GetLocQualityIssueType() rules.ITSIssueType {
	if r == nil || r.IssueType == "" {
		return rules.ITSMisspelling
	}
	return r.IssueType
}

var compoundInfAdjException = map[string]struct{}{
	"schwer": {}, "klar": {}, "verloren": {}, "bekannt": {},
	"rot": {}, "blau": {}, "gelb": {}, "grün": {}, "schwarz": {}, "weiß": {},
	"fertig": {}, "neu": {},
}

var (
	ciAntiOnce  sync.Once
	ciAntiRules []*disambigrules.DisambiguationPatternRule
)

func compoundInfAntiPatternRules() []*disambigrules.DisambiguationPatternRule {
	ciAntiOnce.Do(func() {
		aps := CompoundInfinitivAntiPatterns
		ciAntiRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			rule := disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "de",
				toks, "", nil, disambigrules.ActionImmunize,
			)
			ciAntiRules = append(ciAntiRules, rule)
		}
	})
	return ciAntiRules
}

func (r *CompoundInfinitivRule) getSentenceWithImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := compoundInfAntiPatternRules()
	if len(aps) == 0 {
		return sentence
	}
	src := sentence.GetTokens()
	cloned := make([]*languagetool.AnalyzedTokenReadings, len(src))
	for i, t := range src {
		if t == nil {
			continue
		}
		cloned[i] = languagetool.NewAnalyzedTokenReadingsFromOld(t, t.GetReadings(), "")
	}
	immunized := languagetool.NewAnalyzedSentence(cloned)
	for _, ap := range aps {
		if ap != nil {
			immunized = ap.Replace(immunized)
		}
	}
	return immunized
}

func (r *CompoundInfinitivRule) isMisspelled(word string) bool {
	// Java lowercaseFirstChar then speller; requires speller (IllegalStateException if null).
	w := tools.LowercaseFirstChar(word)
	if r != nil && r.IsMisspelled != nil {
		return r.IsMisspelled(w)
	}
	// Without dict: treat as misspelled so Match's !isMisspelled(join) fails closed
	// (do not invent that "particle+infinitive" is a known compound).
	if !FilterDictAvailable() {
		return true
	}
	return FilterDictIsMisspelled(w)
}

func isInfinitivCI(token *languagetool.AnalyzedTokenReadings) bool {
	return token != nil && token.HasPosTagStartingWith("VER:INF")
}

func isRelevantCI(token *languagetool.AnalyzedTokenReadings) bool {
	return token != nil && token.HasPosTag("ZUS") && !strings.EqualFold(token.GetToken(), "um")
}

func getLemmaCI(token *languagetool.AnalyzedTokenReadings) string {
	if token == nil {
		return ""
	}
	for _, reading := range token.GetReadings() {
		if reading == nil {
			continue
		}
		if l := reading.GetLemma(); l != nil && *l != "" {
			return *l
		}
	}
	return ""
}

func isPunctuationCI(word string) bool {
	if word == "" || utf8.RuneCountInString(word) != 1 {
		return false
	}
	switch word {
	case ".", "?", "!", "…", ":", ";", ",", "(", ")", "[", "]":
		return true
	}
	return false
}

func (r *CompoundInfinitivRule) isException(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
	if n < 2 || n+1 >= len(tokens) || tokens[n-1] == nil || tokens[n-2] == nil || tokens[n+1] == nil {
		return true
	}
	if tokens[n-2].HasPosTagStartingWith("VER") {
		return true
	}
	if _, ok := compoundInfAdjException[tokens[n-1].GetToken()]; ok {
		return true
	}
	if tokens[n+1].GetToken() == "sagen" &&
		(tokens[n-1].GetToken() == "weiter" || tokens[n-1].GetToken() == "dazu") {
		return true
	}
	if (tokens[n+1].GetToken() == "tragen" || tokens[n+1].GetToken() == "machen") &&
		tokens[n-1].GetToken() == "davon" {
		return true
	}
	if tokens[n+1].GetToken() == "geben" && tokens[n-1].GetToken() == "daran" {
		return true
	}
	if tokens[n+1].GetToken() == "gehen" && tokens[n-1].GetToken() == "ab" {
		return true
	}
	if tokens[n+1].GetToken() == "errichten" && tokens[n-1].GetToken() == "wieder" {
		return true
	}
	var verb string
	for i := n - 2; i > 0 && tokens[i] != nil && !isPunctuationCI(tokens[i].GetToken()); i-- {
		if tokens[i].HasPosTagStartingWith("VER:IMP") {
			verb = strings.ToLower(getLemmaCI(tokens[i]))
		} else if tokens[i].HasPosTagStartingWith("VER") {
			verb = strings.ToLower(tokens[i].GetToken())
		} else if tokens[i].GetToken() == "Fang" {
			verb = "fangen"
		}
		if verb != "" {
			if !r.isMisspelled(tokens[n-1].GetToken() + verb) {
				return true
			}
			break
		}
	}
	if tokens[n-1].GetToken() == "aus" || tokens[n-1].GetToken() == "an" {
		for i := n - 2; i > 0 && tokens[i] != nil && !isPunctuationCI(tokens[i].GetToken()); i-- {
			if tokens[i].GetToken() == "von" || tokens[i].GetToken() == "vom" {
				return true
			}
		}
	}
	if tokens[n-1].GetToken() == "her" {
		for i := n - 2; i > 0 && tokens[i] != nil && !isPunctuationCI(tokens[i].GetToken()); i-- {
			if tokens[i].GetToken() == "vor" {
				return true
			}
		}
	}
	return false
}

func (r *CompoundInfinitivRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	imm := r.getSentenceWithImmunization(sentence)
	tokens := imm.GetTokensWithoutWhitespace()
	var ruleMatches []*rules.RuleMatch
	for i := 2; i < len(tokens)-1; i++ {
		if tokens[i] == nil || tokens[i].GetToken() != "zu" {
			continue
		}
		if !isInfinitivCI(tokens[i+1]) || !isRelevantCI(tokens[i-1]) {
			continue
		}
		if tokens[i].IsImmunized() || r.isException(tokens, i) {
			continue
		}
		// Java: !isMisspelled(particle + infinitive) → compound form is a real word
		joined := tokens[i-1].GetToken() + tokens[i+1].GetToken()
		if r.isMisspelled(joined) {
			continue
		}
		// Java: RuleMatch without shortMessage.
		msg := "Wenn der erweiterte Infinitiv von dem Verb '" + joined +
			"' abgeleitet ist, sollte er zusammengeschrieben werden."
		rm := rules.NewRuleMatch(r, sentence, tokens[i-1].GetStartPos(), tokens[i+1].GetEndPos(), msg)
		rm.SetSuggestedReplacement(tokens[i-1].GetToken() + tokens[i].GetToken() + tokens[i+1].GetToken())
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}

// WireCompoundInfinitivRule attaches FilterDictIsMisspelled (Java Morfologik/LinguServices).
// Without a wired dict, joins are treated as misspelled (no false compound hits).
func WireCompoundInfinitivRule(messages map[string]string) *CompoundInfinitivRule {
	r := NewCompoundInfinitivRule(messages)
	r.IsMisspelled = func(w string) bool {
		if !FilterDictAvailable() {
			return true
		}
		return FilterDictIsMisspelled(w)
	}
	return r
}
