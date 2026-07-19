package de

import (
	"regexp"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// GermanWordRepeatRule ports org.languagetool.rules.de.GermanWordRepeatRule.
// Anti-patterns immunize via DisambiguationPatternRule; ignore() is 1:1 with Java
// (POS-only gates — no surface invent for untagged AnalyzePlain).
type GermanWordRepeatRule struct {
	*rules.WordRepeatRule
}

// deSingleChar ports GermanWordRepeatRule.SINGLE_CHAR: (?i)^[a-z]$
var deSingleChar = regexp.MustCompile(`(?i)^[a-z]$`)

func NewGermanWordRepeatRule(messages map[string]string) *GermanWordRepeatRule {
	base := rules.NewWordRepeatRule(messages)
	base.IDOverride = "GERMAN_WORD_REPEAT_RULE"
	// Java GermanWordRepeatRule: super.setCategory(Categories.REDUNDANCY) (overrides base MISC).
	base.Category = rules.CatRedundancy.GetCategory(messages)
	r := &GermanWordRepeatRule{WordRepeatRule: base}
	base.ExtraIgnore = r.germanIgnore
	return r
}

func (r *GermanWordRepeatRule) GetID() string {
	return "GERMAN_WORD_REPEAT_RULE"
}

var (
	gwrAntiOnce  sync.Once
	gwrAntiRules []*disambigrules.DisambiguationPatternRule
)

func germanWordRepeatAntiPatternRules() []*disambigrules.DisambiguationPatternRule {
	gwrAntiOnce.Do(func() {
		aps := GermanWordRepeatAntiPatterns
		gwrAntiRules = make([]*disambigrules.DisambiguationPatternRule, 0, len(aps))
		for _, toks := range aps {
			if len(toks) == 0 {
				continue
			}
			rule := disambigrules.NewDisambiguationPatternRule(
				"INTERNAL_ANTIPATTERN", "(no description)", "de",
				toks, "", nil, disambigrules.ActionImmunize,
			)
			gwrAntiRules = append(gwrAntiRules, rule)
		}
	})
	return gwrAntiRules
}

func (r *GermanWordRepeatRule) getSentenceWithImmunization(sentence *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if sentence == nil {
		return nil
	}
	aps := germanWordRepeatAntiPatternRules()
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

// Match ports WordRepeatRule.match with German anti-pattern immunization.
func (r *GermanWordRepeatRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || r.WordRepeatRule == nil {
		return nil
	}
	return r.WordRepeatRule.Match(r.getSentenceWithImmunization(sentence))
}

// germanIgnore ports GermanWordRepeatRule.ignore (Java then super.ignore via WordRepeatRule.Ignore).
func (r *GermanWordRepeatRule) germanIgnore(tokens []*languagetool.AnalyzedTokenReadings, position int) bool {
	if position < 1 || position >= len(tokens) || tokens[position] == nil || tokens[position-1] == nil {
		return false
	}
	prev := tokens[position-1].GetToken()
	cur := tokens[position].GetToken()

	// Java: position != 2 && Sie sie || sie Sie  (|| lower precedence than &&)
	if (position != 2 && prev == "Sie" && cur == "sie") || (prev == "sie" && cur == "Sie") {
		return true
	}
	// Waren waren / waren Waren
	if (position != 2 && prev == "Waren" && cur == "waren") || (prev == "waren" && cur == "Waren") {
		return true
	}
	// sie sie after KON:UNT or VER:3+ZUS / VER:MOD:3+VER:INF (POS only — Java)
	if position > 2 && prev == "sie" && cur == "sie" && tokens[position-2] != nil {
		p2 := tokens[position-2]
		if p2.HasPosTag("KON:UNT") {
			return true
		}
		// Java: hasPosTag("ZUS") / hasPosTag("VER:INF:NON") — exact tags, not invent partial INF.
		if position+1 < len(tokens) && tokens[position+1] != nil {
			next := tokens[position+1]
			if p2.HasPosTagStartingWith("VER:3:") && next.HasPosTag("ZUS") {
				return true
			}
			if p2.HasPosTagStartingWith("VER:MOD:3") && next.HasPosTag("VER:INF:NON") {
				return true
			}
		}
	}
	// single-char spelling A B B A
	if deSingleChar.MatchString(cur) && position > 1 &&
		tokens[position-2] != nil && deSingleChar.MatchString(tokens[position-2].GetToken()) &&
		position+1 < len(tokens) && tokens[position+1] != nil &&
		deSingleChar.MatchString(tokens[position+1].GetToken()) {
		return true
	}
	return false
}
