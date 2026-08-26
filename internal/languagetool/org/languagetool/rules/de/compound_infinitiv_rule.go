package de

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CompoundInfinitivRule ports org.languagetool.rules.de.CompoundInfinitivRule.
// Java: ZUS + "zu" + VER:INF and !isMisspelled(particle+infinitive) — no surface invent.
// Java: COMPOUNDING, Misspelling, setUrl.
//
// Java constructor (messages, Language lang, UserConfig userConfig):
//   - linguServices from userConfig when non-null; else null
//   - speller lazy-init in match() via MorfologikSpeller("/de/hunspell/de_DE.dict")
//
// Go maps linguServices + Speller to IsMisspelled / FilterDict (same de_DE.dict).
type CompoundInfinitivRule struct {
	Messages  map[string]string
	Category  *rules.Category
	IssueType rules.ITSIssueType
	// LangCode ports Java Language lang (used only with LinguServices in Java).
	LangCode string
	// IsMisspelled ports Java isMisspelled after lowercaseFirstChar.
	// nil → ensureSpeller then FilterDict (Java Speller path); no dict → fail-closed.
	IsMisspelled func(word string) bool
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// NewCompoundInfinitivRule ports CompoundInfinitivRule(messages, lang, null):
// linguServices null; speller is lazy-initialized on Match (ensureSpeller).
func NewCompoundInfinitivRule(messages map[string]string) *CompoundInfinitivRule {
	return NewCompoundInfinitivRuleFor(messages, "de")
}

// NewCompoundInfinitivRuleFor ports CompoundInfinitivRule(messages, lang, null)
// with an explicit language short code (Java Languages.getLanguageForShortCode).
func NewCompoundInfinitivRuleFor(messages map[string]string, langCode string) *CompoundInfinitivRule {
	if langCode == "" {
		langCode = "de"
	}
	r := &CompoundInfinitivRule{
		Messages:  messages,
		Category:  rules.CatCompounding.GetCategory(messages),
		IssueType: rules.ITSMisspelling,
		LangCode:  langCode,
	}
	// Java: sicher zu gehen → sicherzugehen
	r.AddExamplePair(
		rules.Wrong("Er überprüfte die Rechnungen noch einmal, um <marker>sicher zu gehen</marker>."),
		rules.Fixed("Er überprüfte die Rechnungen noch einmal, um <marker>sicherzugehen</marker>."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *CompoundInfinitivRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *CompoundInfinitivRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *CompoundInfinitivRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
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

// ensureSpeller ports CompoundInfinitivRule.match lazy Speller init:
//
//	if (linguServices == null && speller == null) {
//	  speller = new Speller(MorfologikSpeller.getDictionaryWithCaching("/de/hunspell/de_DE.dict"));
//	}
//
// Go: wire process-wide de_DE.dict into FilterDict when not yet available.
func (r *CompoundInfinitivRule) ensureSpeller() {
	if r != nil && r.IsMisspelled != nil {
		return
	}
	if FilterDictAvailable() {
		return
	}
	if p := DiscoverGermanHunspellDict("DE"); p != "" {
		_ = WireGermanFilterSpeller(p)
	}
}

func (r *CompoundInfinitivRule) isMisspelled(word string) bool {
	// Java: word = StringTools.lowercaseFirstChar(word); then Speller / LinguServices.
	// Requires speller after lazy init (IllegalStateException if still null).
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
	if word == "" || utf16LenDE(word) != 1 {
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

// GetAntiPatterns ports CompoundInfinitivRule.getAntiPatterns (cached ANTI_PATTERNS).
func (r *CompoundInfinitivRule) GetAntiPatterns() []*disambigrules.DisambiguationPatternRule {
	return compoundInfAntiPatternRules()
}

func (r *CompoundInfinitivRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	// Java match(): lazy-init Speller from /de/hunspell/de_DE.dict when linguServices null.
	r.ensureSpeller()
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
		// Java: suggestions.add(particle + "zu" + infinitive)
		rm.SetSuggestedReplacements([]string{
			tokens[i-1].GetToken() + tokens[i].GetToken() + tokens[i+1].GetToken(),
		})
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}

// WireCompoundInfinitivRule attaches FilterDictIsMisspelled (Java Morfologik Speller /
// LinguServices). Also ensures de_DE.dict is discovered like match() lazy init.
// Without a wired dict, joins are treated as misspelled (no false compound hits).
func WireCompoundInfinitivRule(messages map[string]string) *CompoundInfinitivRule {
	r := NewCompoundInfinitivRule(messages)
	r.ensureSpeller()
	r.IsMisspelled = func(w string) bool {
		if !FilterDictAvailable() {
			return true
		}
		return FilterDictIsMisspelled(w)
	}
	return r
}
