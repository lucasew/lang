package de

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PotentialCompoundFilter ports org.languagetool.rules.de.PotentialCompoundFilter.
//
// Java acceptRuleMatch:
//  1. part1/part2 casing via StringTools.isMixedCase / isAllUppercase / uppercaseFirstChar
//  2. joinedWord / hyphenatedWord
//  3. tag(joinedWord); AnalyzedSentence(SENT_START + tagged); default spelling rule match
//  4. matches.length==0 → valid: length>20 adds hyphenated first, always add joined; else hyphenated only
//  5. new RuleMatch + setType(match.getType()) + setSuggestedReplacements
//
// Go wiring: GermanyGerman.getDefaultSpellingRule().match is GermanSpellerRule.Match on a
// mini sentence when FilterDict is process-wired (WireGermanFilterSpeller / WireGermanRuntimeResources).
// Without dict, fail-closed (hyphenated only) — do not invent that a joined form is correct.
// JoinedIsValid overrides the speller path when set (tests).
type PotentialCompoundFilter struct {
	// JoinedIsValid optional override; nil uses default-spelling-rule match (Java).
	JoinedIsValid func(joined string) bool
	// Speller optional GermanSpellerRule (Java getDefaultSpellingRule). Nil → process-wide bare rule.
	Speller *GermanSpellerRule
}

func NewPotentialCompoundFilter() *PotentialCompoundFilter {
	return &PotentialCompoundFilter{}
}

// process-wide bare GermanyGerman default spelling rule for filter match checks.
var (
	potentialCompoundSpellerOnce sync.Once
	potentialCompoundSpeller     *GermanSpellerRule
)

func defaultPotentialCompoundSpeller() *GermanSpellerRule {
	potentialCompoundSpellerOnce.Do(func() {
		potentialCompoundSpeller = NewGermanSpellerRule(nil)
	})
	return potentialCompoundSpeller
}

// joinedValid ports Java:
//
//	RuleMatch[] matches = getDefaultSpellingRule().match(analyzedSentence);
//	matches.length == 0  ⇒ joined form accepted
func (f *PotentialCompoundFilter) joinedValid(joined string) bool {
	if f != nil && f.JoinedIsValid != nil {
		return f.JoinedIsValid(joined)
	}
	// Java always has a spelling rule; without wired dict Match returns empty
	// (would look valid) — fail-closed instead of inventing correctness.
	if !FilterDictAvailable() {
		return false
	}
	matches := f.defaultSpellingMatch(joined)
	return len(matches) == 0
}

// defaultSpellingMatch ports the tag + AnalyzedSentence + getDefaultSpellingRule().match path.
func (f *PotentialCompoundFilter) defaultSpellingMatch(joinedWord string) []*rules.RuleMatch {
	// Java: tagger.tag(singletonList(joinedWord)); SENT_START + tagged token
	sentence := buildPotentialCompoundSentence(joinedWord)
	sp := defaultPotentialCompoundSpeller()
	if f != nil && f.Speller != nil {
		sp = f.Speller
	}
	if sp == nil {
		return []*rules.RuleMatch{{}} // non-empty ⇒ treat invalid (fail-closed)
	}
	return sp.Match(sentence)
}

// buildPotentialCompoundSentence ports Java mini-sentence construction:
// AnalyzedToken("", "SENT_START", null) + tagger.tag([joinedWord])[0].
func buildPotentialCompoundSentence(joinedWord string) *languagetool.AnalyzedSentence {
	ss := languagetool.SentenceStartTagName
	startTok := languagetool.NewAnalyzedToken("", &ss, nil)
	startATR := languagetool.NewAnalyzedTokenReadingsList([]*languagetool.AnalyzedToken{startTok}, 0)

	var contentATR *languagetool.AnalyzedTokenReadings
	if tagger := DiscoveredGermanTagger(); tagger != nil {
		tagged := tagger.Tag([]string{joinedWord})
		if len(tagged) > 0 && tagged[0] != nil {
			contentATR = tagged[0]
		}
	}
	if contentATR == nil {
		// No tagger: still surface-token the joined word (Match uses GetToken only).
		contentATR = languagetool.NewAnalyzedTokenReadingsAt(
			languagetool.NewAnalyzedToken(joinedWord, nil, nil), 0)
	}
	return languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{startATR, contentATR})
}

// Suggestions returns replacement strings for part1+part2 (Java acceptRuleMatch body).
func (f *PotentialCompoundFilter) Suggestions(part1, part2 string) []string {
	// Java: String part1capitalized = part1; part2capitalized = part2; part2lowercase = part2;
	part1capitalized := part1
	part2capitalized := part2
	part2lowercase := part2
	// Java: if (!isMixedCase(part2) && !isAllUppercase(part2))
	if !tools.IsMixedCase(part2) && !tools.IsAllUppercase(part2) {
		part2lowercase = strings.ToLower(part2)
		part2capitalized = tools.UppercaseFirstChar(strings.ToLower(part2))
	}
	if !tools.IsMixedCase(part1) && !tools.IsAllUppercase(part1) {
		part1capitalized = tools.UppercaseFirstChar(strings.ToLower(part1))
	}
	joinedWord := part1capitalized + part2lowercase
	hyphenatedWord := part1capitalized + "-" + part2capitalized
	var replacements []string
	if f.joinedValid(joinedWord) {
		// Java: if (joinedWord.length() > 20) replacements.add(hyphenatedWord);
		// length() = UTF-16 code units
		if utf16LenDE(joinedWord) > 20 {
			replacements = append(replacements, hyphenatedWord)
		}
		replacements = append(replacements, joinedWord)
	} else {
		replacements = append(replacements, hyphenatedWord)
	}
	return replacements
}

// AcceptRuleMatch ports PotentialCompoundFilter.acceptRuleMatch.
func (f *PotentialCompoundFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	replacements := f.Suggestions(arguments["part1"], arguments["part2"])
	// Java: if (!replacements.isEmpty()) { ... return ruleMatch; } return null;
	if len(replacements) == 0 {
		return nil
	}
	// Java: new RuleMatch(rule, sentence, from, to, message, shortMessage)
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), match.GetMessage())
	out.ShortMessage = match.ShortMessage
	// Java: ruleMatch.setType(match.getType())
	out.SetType(match.GetType())
	out.SetSuggestedReplacements(replacements)
	return out
}
