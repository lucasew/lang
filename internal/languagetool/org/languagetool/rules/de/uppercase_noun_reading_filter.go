package de

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// UppercaseNounReadingFilter ports org.languagetool.rules.de.UppercaseNounReadingFilter.
// Accepts rule matches when the uppercased word of the "token" argument has a noun
// reading, e.g. German "stand" (past of "stehen") → "Stand" (noun) → accept.
//
// Java uses GermanTagger.INSTANCE.tag(singletonList(uppercase)).
// Go: DiscoveredGermanTagger / optional Tag override (same pattern as InsertCommaFilter).
// Without a tagger → fail-closed drop (do not invent noun readings).
type UppercaseNounReadingFilter struct {
	// Tag ports GermanTagger.INSTANCE for uppercased-token lookups (tests may inject).
	// Nil → resolveTagger → default / DiscoveredGermanTagger.
	Tag *detag.GermanTagger
}

func NewUppercaseNounReadingFilter() *UppercaseNounReadingFilter {
	return &UppercaseNounReadingFilter{}
}

var (
	uppercaseNounTaggerMu      sync.RWMutex
	defaultUppercaseNounTagger *detag.GermanTagger
)

// SetDefaultUppercaseNounGermanTagger installs the GermanTagger.INSTANCE twin.
func SetDefaultUppercaseNounGermanTagger(t *detag.GermanTagger) {
	uppercaseNounTaggerMu.Lock()
	defer uppercaseNounTaggerMu.Unlock()
	defaultUppercaseNounTagger = t
}

// SetDefaultUppercaseNounTagger is a legacy POS-func hook kept for WireGermanTaggerDefaults
// compatibility; production tagging uses GermanTagger.Tag (see AcceptRuleMatch), not the
// []string invent path.
func SetDefaultUppercaseNounTagger(_ func(string) []string) {
	uppercaseNounTaggerMu.Lock()
	defer uppercaseNounTaggerMu.Unlock()
	if defaultUppercaseNounTagger == nil {
		defaultUppercaseNounTagger = DiscoveredGermanTagger()
	}
}

func (f *UppercaseNounReadingFilter) resolveTagger() *detag.GermanTagger {
	if f != nil && f.Tag != nil {
		return f.Tag
	}
	uppercaseNounTaggerMu.RLock()
	t := defaultUppercaseNounTagger
	uppercaseNounTaggerMu.RUnlock()
	if t != nil {
		return t
	}
	return DiscoveredGermanTagger()
}

// Accept reports whether the uppercased token has a pure noun reading
// (hasPartialPosTag("SUB:") && !hasPartialPosTag("ADJ")). Used by tests.
// Without a tagger → false (fail-closed; no invent).
func (f *UppercaseNounReadingFilter) Accept(token string) bool {
	uppercase := tools.UppercaseFirstChar(token)
	tagger := f.resolveTagger()
	if tagger == nil {
		return false
	}
	// Java: GermanTagger.INSTANCE.tag(Collections.singletonList(uppercase))
	for _, tag := range tagger.Tag([]string{uppercase}) {
		if tag != nil && tag.HasPartialPosTag("SUB:") && !tag.HasPartialPosTag("ADJ") {
			return true
		}
	}
	return false
}

// AcceptRuleMatch ports UppercaseNounReadingFilter.acceptRuleMatch.
// Args: token — surface form to uppercase and check for SUB: (non-ADJ) reading.
func (f *UppercaseNounReadingFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	// Java: String token = arguments.get("token"); if (token == null) throw ...
	token, ok := arguments["token"]
	if !ok {
		ruleID := ""
		if r := match.GetRule(); r != nil {
			if idr, ok := r.(rules.RuleWithID); ok {
				ruleID = idr.GetID()
			}
		}
		// Java: "Set 'token' for filter " + UppercaseNounReadingFilter.class.getName() + " in rule " + match.getRule().getId()
		panic("Set 'token' for filter org.languagetool.rules.de.UppercaseNounReadingFilter in rule " + ruleID)
	}
	uppercase := tools.UppercaseFirstChar(token)
	tagger := f.resolveTagger()
	if tagger == nil {
		// Java INSTANCE is never null; without resources fail-closed (no invent).
		return nil
	}
	// Java: List<AnalyzedTokenReadings> tags = GermanTagger.INSTANCE.tag(Collections.singletonList(uppercase));
	// for (AnalyzedTokenReadings tag : tags) {
	//   if (tag.hasPartialPosTag("SUB:") && !tag.hasPartialPosTag("ADJ")) return match;
	// }
	// return null;
	for _, tag := range tagger.Tag([]string{uppercase}) {
		if tag != nil && tag.HasPartialPosTag("SUB:") && !tag.HasPartialPosTag("ADJ") {
			return match
		}
	}
	return nil
}
