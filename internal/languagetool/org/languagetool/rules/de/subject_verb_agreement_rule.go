package de

import (
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// Chunk tags used by German chunker (Java ChunkTag NPS/NPP/PP).
const (
	chunkNPS = "NPS" // noun phrase singular
	chunkNPP = "NPP" // noun phrase plural
	chunkPP  = "PP"  // prepositional phrase etc.
)

// SubjectVerbAgreementRule ports org.languagetool.rules.de.SubjectVerbAgreementRule.
// Checks ist/sind/war/waren against preceding NP number via chunk tags (NPS/NPP)
// + NOM morph + ANTI_PATTERNS. No POS invent when chunks are absent (Java).
// Optional LookupInfinitive for containsOnlyInfinitivesToTheLeft (Java GermanTagger.lookup).
type SubjectVerbAgreementRule struct {
	Messages map[string]string
	// Category ports setCategory(GRAMMAR).
	Category *rules.Category
	singular map[string]struct{}
	plural   map[string]struct{}
	// LookupInfinitive ports GermanTagger.lookup(token.toLowerCase()) used to test VER:INF.
	// When nil, containsOnlyInfinitivesToTheLeft is treated as false (never suppresses).
	LookupInfinitive func(lowerWord string) bool
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []rules.IncorrectExample
	correctExamples   []rules.CorrectExample
}

// SingularPluralPair ports Java SingularPluralPair.
type SingularPluralPair struct {
	Singular string
	Plural   string
}

// subjectVerbPairs ports PAIRS (ist/sind, war/waren only — Java comments others out).
var subjectVerbPairs = []SingularPluralPair{
	{"ist", "sind"},
	{"war", "waren"},
}

var subjectVerbCurrencies = map[string]struct{}{
	"Dollar": {}, "Euro": {}, "Yen": {},
}

var subjectVerbQuestionPronouns = map[string]struct{}{
	"wie": {},
}

var (
	subjectVerbAntiOnce  sync.Once
	subjectVerbAntiRules []*disambigrules.DisambiguationPatternRule
	// Java: "wer|(?i)alle[nr]?|(?i)jede[rs]?|(?i)manche[nrs]?"
	leftRegexAlle = regexp.MustCompile(`(?i)^(wer|alle[nr]?|jede[rs]?|manche[nrs]?)$`)
)

func NewSubjectVerbAgreementRule(messages map[string]string) *SubjectVerbAgreementRule {
	r := &SubjectVerbAgreementRule{
		Messages: messages,
		Category: rules.CatGrammar.GetCategory(messages),
		singular: map[string]struct{}{},
		plural:   map[string]struct{}{},
	}
	for _, p := range subjectVerbPairs {
		r.singular[p.Singular] = struct{}{}
		r.plural[p.Plural] = struct{}{}
	}
	// Java: Die Autos ist → sind
	r.AddExamplePair(
		rules.Wrong("Die Autos <marker>ist</marker> schnell."),
		rules.Fixed("Die Autos <marker>sind</marker> schnell."),
	)
	return r
}

// AddExamplePair ports Rule.addExamplePair.
func (r *SubjectVerbAgreementRule) AddExamplePair(incorrect rules.IncorrectExample, correct rules.CorrectExample) {
	if r == nil {
		return
	}
	var br rules.BaseRule
	br.AddExamplePair(incorrect, correct)
	r.incorrectExamples = append(r.incorrectExamples, br.GetIncorrectExamples()...)
	r.correctExamples = append(r.correctExamples, br.GetCorrectExamples()...)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *SubjectVerbAgreementRule) GetIncorrectExamples() []rules.IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]rules.IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *SubjectVerbAgreementRule) GetCorrectExamples() []rules.CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]rules.CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// WithLookupInfinitive sets the tagger hook for containsOnlyInfinitivesToTheLeft.
func (r *SubjectVerbAgreementRule) WithLookupInfinitive(fn func(lowerWord string) bool) *SubjectVerbAgreementRule {
	if r != nil {
		r.LookupInfinitive = fn
	}
	return r
}

func (r *SubjectVerbAgreementRule) GetID() string { return "DE_SUBJECT_VERB_AGREEMENT" }

// GetDescription ports SubjectVerbAgreementRule.getDescription (Java string).
func (r *SubjectVerbAgreementRule) GetDescription() string {
	return "Kongruenz von Subjekt und Prädikat (unvollständig)"
}

// EstimateContextForSureMatch ports estimateContextForSureMatch:
// max length of ANTI_PATTERNS lists.
func (r *SubjectVerbAgreementRule) EstimateContextForSureMatch() int {
	max := 0
	for _, ap := range SubjectVerbAntiPatterns {
		if n := len(ap); n > max {
			max = n
		}
	}
	return max
}

// GetURL ports SubjectVerbAgreementRule.getUrl.
func (r *SubjectVerbAgreementRule) GetURL() string {
	return "https://dict.leo.org/grammatik/deutsch/Wort/Verb/Kategorien/Numerus-Person/ProblemNum.html"
}

func (r *SubjectVerbAgreementRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func subjectVerbAntiPatternRules() []*disambigrules.DisambiguationPatternRule {
	subjectVerbAntiOnce.Do(func() {
		// Full table from SubjectVerbAntiPatterns (Java ANTI_PATTERNS).
		aps := SubjectVerbAntiPatterns
		subjectVerbAntiRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			rule := disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "de",
				toks, "", nil, disambigrules.ActionImmunize,
			)
			subjectVerbAntiRules = append(subjectVerbAntiRules, rule)
		}
	})
	return subjectVerbAntiRules
}

func (r *SubjectVerbAgreementRule) getSentenceWithImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := subjectVerbAntiPatternRules()
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

func (r *SubjectVerbAgreementRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if sentence == nil {
		return nil
	}
	imm := r.getSentenceWithImmunization(sentence)
	tokens := imm.GetTokensWithoutWhitespace()
	var matches []*rules.RuleMatch
	for i := 1; i < len(tokens); i++ {
		if tokens[i] == nil || tokens[i].IsImmunized() {
			continue
		}
		tokenStr := tokens[i].GetToken()
		if rm := r.getSingularMatchOrNull(tokens, i, tokens[i], tokenStr, sentence); rm != nil {
			matches = append(matches, rm)
		}
		if rm := r.getPluralMatchOrNull(tokens, i, tokens[i], tokenStr, sentence); rm != nil {
			matches = append(matches, rm)
		}
	}
	return matches
}

func hasChunk(t *languagetool.AnalyzedTokenReadings, tag string) bool {
	if t == nil {
		return false
	}
	for _, c := range t.GetChunkTags() {
		if c == tag {
			return true
		}
	}
	return false
}

// prevIsPluralNP ports prevChunkTags.contains(NPP) only (Java — no POS invent).
func prevIsPluralNP(prev *languagetool.AnalyzedTokenReadings) bool {
	return hasChunk(prev, chunkNPP)
}

// prevIsSingularNP ports prevChunkTags.contains(NPS) only (Java — no POS invent).
func prevIsSingularNP(prev *languagetool.AnalyzedTokenReadings) bool {
	return hasChunk(prev, chunkNPS)
}

func prevHasPP(prev *languagetool.AnalyzedTokenReadings) bool {
	return hasChunk(prev, chunkPP)
}

func (r *SubjectVerbAgreementRule) getSingularMatchOrNull(tokens []*languagetool.AnalyzedTokenReadings, i int,
	token *languagetool.AnalyzedTokenReadings, tokenStr string, sentence *languagetool.AnalyzedSentence) *rules.RuleMatch {
	if _, ok := r.singular[tokenStr]; !ok {
		return nil
	}
	if i < 1 {
		return nil
	}
	prev := tokens[i-1]
	var next *languagetool.AnalyzedTokenReadings
	if i+1 < len(tokens) {
		next = tokens[i+1]
	}
	// Java: NPP && !PP && …
	match := prevIsPluralNP(prev) &&
		!prevHasPP(prev) &&
		prev.GetToken() != "Uhr" &&
		!isSubjectVerbCurrency(prev) &&
		!(next != nil && next.GetToken() == "es") &&
		prevChunkIsNominative(tokens, i-1) &&
		!hasUnknownTokenToTheLeft(tokens, i) &&
		!hasQuestionPronounToTheLeft(tokens, i-1) &&
		!hasVerbToTheLeft(tokens, i-1) &&
		!containsRegexToTheLeft(leftRegexAlle, tokens, i-1) &&
		!r.containsOnlyInfinitivesToTheLeft(tokens, i-1)
	if !match {
		return nil
	}
	// Java embeds <suggestion>…</suggestion> in the message (RuleMatch extracts replacements).
	// No shortMessage in Java constructor. Keep structured suggestion for Go clients.
	sug := getPluralFor(tokenStr)
	msg := "Bitte prüfen, ob hier <suggestion>" + sug + "</suggestion> stehen sollte."
	rm := rules.NewRuleMatch(r, sentence, token.GetStartPos(), token.GetEndPos(), msg)
	rm.SetSuggestedReplacement(sug)
	return rm
}

func (r *SubjectVerbAgreementRule) getPluralMatchOrNull(tokens []*languagetool.AnalyzedTokenReadings, i int,
	token *languagetool.AnalyzedTokenReadings, tokenStr string, sentence *languagetool.AnalyzedSentence) *rules.RuleMatch {
	if _, ok := r.plural[tokenStr]; !ok {
		return nil
	}
	if i < 1 {
		return nil
	}
	prev := tokens[i-1]
	var next *languagetool.AnalyzedTokenReadings
	if i+1 < len(tokens) {
		next = tokens[i+1]
	}
	// tokens[1] may be first content word (after SENT_START at 0)
	firstContent := ""
	if len(tokens) > 1 && tokens[1] != nil {
		firstContent = tokens[1].GetToken()
	}
	match := prevIsSingularNP(prev) &&
		!(next != nil && next.GetToken() == "Sie") &&
		!prevIsPluralNP(prev) &&
		!prevHasPP(prev) &&
		!isSubjectVerbCurrency(prev) &&
		prevChunkIsNominative(tokens, i-1) &&
		!hasUnknownTokenToTheLeft(tokens, i) &&
		!hasUnknownTokenToTheRight(tokens, i+1) &&
		firstContent != "Alle" && firstContent != "Viele" &&
		!isFollowedByNominativePlural(tokens, i+1)
	if !match {
		return nil
	}
	sug := getSingularFor(tokenStr)
	msg := "Bitte prüfen, ob hier <suggestion>" + sug + "</suggestion> stehen sollte."
	rm := rules.NewRuleMatch(r, sentence, token.GetStartPos(), token.GetEndPos(), msg)
	rm.SetSuggestedReplacement(sug)
	return rm
}

func isSubjectVerbCurrency(token *languagetool.AnalyzedTokenReadings) bool {
	if token == nil {
		return false
	}
	_, ok := subjectVerbCurrencies[token.GetToken()]
	return ok
}

// prevChunkIsNominative ports SubjectVerbAgreementRule.prevChunkIsNominative exactly:
// walk left while tokens carry NPS/NPP; require NOM on some token in that span.
// No POS-only invent when chunk tags are absent (Java returns false).
func prevChunkIsNominative(tokens []*languagetool.AnalyzedTokenReadings, startPos int) bool {
	if startPos <= 0 || startPos >= len(tokens) {
		return false
	}
	for i := startPos; i > 0; i-- {
		if tokens[i] == nil {
			return false
		}
		if hasChunk(tokens[i], chunkNPS) || hasChunk(tokens[i], chunkNPP) {
			if tokens[i].HasPartialPosTag("NOM") {
				return true
			}
		} else {
			return false
		}
	}
	return false
}

func hasUnknownTokenToTheLeft(tokens []*languagetool.AnalyzedTokenReadings, startPos int) bool {
	return hasUnknownTokenAt(tokens, 0, startPos)
}

func hasUnknownTokenToTheRight(tokens []*languagetool.AnalyzedTokenReadings, startPos int) bool {
	if startPos < 0 {
		return false
	}
	return hasUnknownTokenAt(tokens, startPos, len(tokens)-1)
}

func hasUnknownTokenAt(tokens []*languagetool.AnalyzedTokenReadings, startPos, endPos int) bool {
	if endPos > len(tokens) {
		endPos = len(tokens)
	}
	for i := startPos; i < endPos; i++ {
		if tokens[i] == nil {
			continue
		}
		// Only treat as unknown when token claims to be open-class but has no tag
		// Java: any reading with hasNoTag. Soft: skip sentence markers.
		if tokens[i].IsSentenceStart() || tokens[i].IsSentenceEnd() {
			continue
		}
		for _, at := range tokens[i].GetReadings() {
			if at != nil && at.HasNoTag() {
				// Java: unknown token to the left/right suppresses the match.
				return true
			}
		}
	}
	return false
}

func hasQuestionPronounToTheLeft(tokens []*languagetool.AnalyzedTokenReadings, startPos int) bool {
	for i := startPos; i > 0; i-- {
		if tokens[i] == nil {
			continue
		}
		if _, ok := subjectVerbQuestionPronouns[strings.ToLower(tokens[i].GetToken())]; ok {
			return true
		}
	}
	return false
}

func hasVerbToTheLeft(tokens []*languagetool.AnalyzedTokenReadings, startPos int) bool {
	for i := startPos; i > 0; i-- {
		if tokens[i] != nil && tokens[i].MatchesPosTagRegex("VER:[1-3]:.+") {
			return true
		}
	}
	return false
}

func containsRegexToTheLeft(re *regexp.Regexp, tokens []*languagetool.AnalyzedTokenReadings, startPos int) bool {
	for i := startPos; i > 0; i-- {
		if tokens[i] != nil && re.MatchString(tokens[i].GetToken()) {
			return true
		}
	}
	return false
}

func isFollowedByNominativePlural(tokens []*languagetool.AnalyzedTokenReadings, startPos int) bool {
	for i := startPos; i < len(tokens); i++ {
		t := tokens[i]
		if t == nil {
			continue
		}
		if (t.HasPartialPosTag("SUB") || t.HasPartialPosTag("PRO")) &&
			(t.HasPartialPosTag("NOM:PLU") || hasChunk(t, chunkNPP)) {
			return true
		}
	}
	return false
}

// containsOnlyInfinitivesToTheLeft ports SubjectVerbAgreementRule.containsOnlyInfinitivesToTheLeft.
// "Das Kopieren und Einfügen ist sehr nützlich." — SUB tokens that are also VER:INF via tagger.
// Without LookupInfinitive, returns false (Java path not available → do not suppress errors).
func (r *SubjectVerbAgreementRule) containsOnlyInfinitivesToTheLeft(tokens []*languagetool.AnalyzedTokenReadings, startPos int) bool {
	if r == nil || r.LookupInfinitive == nil {
		return false
	}
	infinitives := 0
	for i := startPos; i > 0; i-- {
		if tokens[i] == nil {
			continue
		}
		// Java hasPartialPosTag("SUB:")
		if tokens[i].HasPartialPosTag("SUB:") {
			// Java: lookup(token.toLowerCase()) and hasPosTagStartingWith("VER:INF")
			if r.LookupInfinitive(strings.ToLower(tokens[i].GetToken())) {
				infinitives++
			} else {
				return false
			}
		}
	}
	return infinitives >= 2
}

func getSingularFor(token string) string {
	for _, p := range subjectVerbPairs {
		if p.Plural == token {
			return p.Singular
		}
	}
	return token
}

func getPluralFor(token string) string {
	for _, p := range subjectVerbPairs {
		if p.Singular == token {
			return p.Plural
		}
	}
	return token
}
