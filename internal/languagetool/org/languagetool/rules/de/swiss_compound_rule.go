package de

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// SwissCompoundRule ports org.languagetool.rules.de.SwissCompoundRule.
//
// Java:
//
//	public class SwissCompoundRule extends GermanCompoundRule {
//	  public SwissCompoundRule(...) { super(messages, lang, userConfig); }
//	  @Override public String getId() { return "DE_CH_COMPOUNDS"; }
//	  @Override public CompoundRuleData getCompoundRuleData() {
//	    // lazy singleton CompoundRuleData(new SwissExpander(),
//	    //   "/de/compounds.txt", "/de/compound-cities.txt")
//	  }
//	  static class SwissExpander implements LineExpander { ... }
//	}
//
// Inherits GermanCompoundRule isMisspelled (GermanyGerman) and ANTI_PATTERNS.
type SwissCompoundRule struct {
	*GermanCompoundRule
}

var (
	chCompoundOnce sync.Once
	chCompoundData *rules.CompoundRuleData
)

// loadCHCompoundData ports SwissCompoundRule.getCompoundRuleData lazy singleton.
func loadCHCompoundData() *rules.CompoundRuleData {
	chCompoundOnce.Do(func() {
		// Java: new CompoundRuleData(new SwissExpander(), "/de/compounds.txt", "/de/compound-cities.txt")
		chCompoundData = mustLoadCompoundData(SwissExpandLine)
	})
	return chCompoundData
}

// NewSwissCompoundRule ports SwissCompoundRule(ResourceBundle, Language, UserConfig).
// Java only calls super(...); id + compound data come from overrides.
func NewSwissCompoundRule(messages map[string]string) *SwissCompoundRule {
	g := NewGermanCompoundRule(messages)
	// Override fields that Java overrides via getId / getCompoundRuleData.
	g.ID = "DE_CH_COMPOUNDS"
	g.Data = loadCHCompoundData()
	return &SwissCompoundRule{GermanCompoundRule: g}
}

// GetID ports SwissCompoundRule.getId.
func (r *SwissCompoundRule) GetID() string {
	return "DE_CH_COMPOUNDS"
}

// GetCompoundRuleData ports SwissCompoundRule.getCompoundRuleData.
func (r *SwissCompoundRule) GetCompoundRuleData() *rules.CompoundRuleData {
	return loadCHCompoundData()
}

// Match ports inherited GermanCompoundRule.match (ANTI_PATTERNS + AbstractCompoundRule).
func (r *SwissCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	return r.GermanCompoundRule.Match(sentence)
}

// GetAntiPatterns ports inherited GermanCompoundRule.getAntiPatterns.
func (r *SwissCompoundRule) GetAntiPatterns() []*disambigrules.DisambiguationPatternRule {
	return r.GermanCompoundRule.GetAntiPatterns()
}

// SwissExpandLine ports SwissCompoundRule.SwissExpander.expandLine.
//
// Java:
//
//	if (line.contains("ß")) {
//	  return Arrays.asList(line, line.replace("ß", "ss"));
//	} else {
//	  return Arrays.asList(line);
//	}
func SwissExpandLine(line string) []string {
	if strings.Contains(line, "ß") {
		return []string{line, strings.ReplaceAll(line, "ß", "ss")}
	}
	return []string{line}
}

// Ensure SwissExpandLine matches CompoundRuleData LineExpander signature.
var _ rules.LineExpander = SwissExpandLine
