package de

import (
	"fmt"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// UppercaseNounReadingFilter ports org.languagetool.rules.de.UppercaseNounReadingFilter.
// TagPOS ports GermanTagger.INSTANCE.tag for the uppercased form (POS strings).
// Nil TagPOS (and no default) → fail-closed drop (do not invent noun readings).
type UppercaseNounReadingFilter struct {
	// TagPOS returns POS tags for a surface form; nil uses default or fail-closed.
	TagPOS func(uppercased string) []string
}

func NewUppercaseNounReadingFilter() *UppercaseNounReadingFilter {
	return &UppercaseNounReadingFilter{}
}

var (
	uppercaseNounTagMu      sync.RWMutex
	defaultUppercaseNounTag func(string) []string
)

// SetDefaultUppercaseNounTagger wires German tagger POS for this filter.
func SetDefaultUppercaseNounTagger(tag func(string) []string) {
	uppercaseNounTagMu.Lock()
	defer uppercaseNounTagMu.Unlock()
	defaultUppercaseNounTag = tag
}

func (f *UppercaseNounReadingFilter) resolveTag() func(string) []string {
	if f != nil && f.TagPOS != nil {
		return f.TagPOS
	}
	uppercaseNounTagMu.RLock()
	defer uppercaseNounTagMu.RUnlock()
	return defaultUppercaseNounTag
}

// hasNounReading ports the Java loop:
// tag.hasPartialPosTag("SUB:") && !tag.hasPartialPosTag("ADJ")
func hasNounReadingFromPOS(tags []string) bool {
	if len(tags) == 0 {
		return false
	}
	hasSUB, hasADJ := false, false
	for _, t := range tags {
		if strings.Contains(t, "SUB:") {
			hasSUB = true
		}
		if strings.Contains(t, "ADJ") {
			hasADJ = true
		}
	}
	return hasSUB && !hasADJ
}

// Accept reports whether the uppercased token has a pure noun reading (for tests).
// Without TagPOS → false (fail-closed; no soft invent).
func (f *UppercaseNounReadingFilter) Accept(token string) bool {
	if token == "" {
		panic("token required for UppercaseNounReadingFilter")
	}
	upper := tools.UppercaseFirstChar(token)
	tag := f.resolveTag()
	if tag == nil {
		return false
	}
	return hasNounReadingFromPOS(tag(upper))
}

// AcceptRuleMatch ports UppercaseNounReadingFilter.acceptRuleMatch.
// Args: token — surface form to uppercase and check for SUB: (non-ADJ) reading.
func (f *UppercaseNounReadingFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	token, ok := arguments["token"]
	if !ok || token == "" {
		// Java: throw new RuntimeException("Set 'token' for filter ... in rule " + id)
		ruleID := ""
		if match.GetRule() != nil {
			ruleID = fmt.Sprintf("%v", match.GetRule())
		}
		panic("Set 'token' for filter org.languagetool.rules.de.UppercaseNounReadingFilter in rule " + ruleID)
	}
	if !f.Accept(token) {
		return nil
	}
	return match
}
