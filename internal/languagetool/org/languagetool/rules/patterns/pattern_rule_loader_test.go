package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPatternRuleLoader(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="R1" name="Rule one">
      <pattern>
        <token>foo</token>
        <token regexp="yes">b.r</token>
      </pattern>
      <message>bad</message>
      <short>s</short>
    </rule>
  </category>
</rules>`
	rules, err := NewPatternRuleLoader().GetRulesFromString(xml, "test.xml", "en")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.Equal(t, "R1", rules[0].ID)
	require.Equal(t, "Rule one", rules[0].Description)
	require.Len(t, rules[0].PatternTokens, 2)
	require.Equal(t, "foo", rules[0].PatternTokens[0].Token)
	require.True(t, rules[0].PatternTokens[1].Regexp)
	require.Equal(t, "bad", rules[0].Message)

	// round-trip match
	pr := NewPatternRule(rules[0].ID, "en", rules[0].PatternTokens, rules[0].Description, rules[0].Message, rules[0].ShortMessage)
	require.NotNil(t, pr)
}

// Rules with <unify> load with UniFeatures and TestUnification (matcher ports testUnification).
func TestPatternRuleLoader_UnifyLoads(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <unification feature="number">
    <equivalence type="sg"><token postag="NN"/></equivalence>
    <equivalence type="pl"><token postag="NNS"/></equivalence>
  </unification>
  <category>
    <rule id="U1" name="needs unify">
      <pattern>
        <unify>
          <feature id="number"/>
          <token postag="NN"/>
          <token postag="VB"/>
        </unify>
      </pattern>
      <message>u</message>
    </rule>
    <rule id="U2" name="negate unify">
      <pattern>
        <unify negate="yes">
          <feature id="number"/>
          <token postag="NN"/>
          <token postag="NNS"/>
        </unify>
      </pattern>
      <message>neg</message>
    </rule>
    <rule id="OK" name="surface">
      <pattern>
        <token>hello</token>
      </pattern>
      <message>ok</message>
    </rule>
  </category>
</rules>`
	loader := NewPatternRuleLoader()
	ars, err := loader.GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 3)
	require.Equal(t, "U1", ars[0].ID)
	require.True(t, ars[0].TestUnification)
	require.NotNil(t, ars[0].UnifierConfig)
	require.Len(t, ars[0].PatternTokens, 2)
	require.True(t, ars[0].PatternTokens[0].IsUnified())
	require.True(t, ars[0].PatternTokens[1].IsLastInUnification())
	require.False(t, ars[0].PatternTokens[1].IsUniNegated())
	require.True(t, ars[1].PatternTokens[1].IsUniNegated())
	require.Equal(t, "OK", ars[2].ID)
	// Equivalence tables from <unification>
	require.NotNil(t, loader.LastUnifierConfig)
	types := loader.LastUnifierConfig.GetEquivalenceTypes()
	require.Contains(t, types, NewEquivalenceTypeLocator("number", "sg"))
}

func TestPatternRuleLoader_PhraserefExpands(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <phrases>
    <phrase id="A">
      <token>hello</token>
    </phrase>
    <phrase id="B">
      <token>hi</token>
    </phrase>
    <phrase id="START">
      <includephrases>
        <phraseref idref="A"/>
        <phraseref idref="B"/>
      </includephrases>
    </phrase>
  </phrases>
  <category>
    <rule id="R" name="phrase rule">
      <pattern>
        <phraseref idref="START"/>
        <token>world</token>
      </pattern>
      <message>m</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	// START expands to hello|hi, each + world → 2 rules
	require.Len(t, ars, 2)
	surfaces := map[string]bool{}
	for _, ar := range ars {
		require.Equal(t, "R", ar.ID)
		require.Len(t, ar.PatternTokens, 2)
		require.Equal(t, "world", ar.PatternTokens[1].Token)
		surfaces[ar.PatternTokens[0].Token] = true
	}
	require.True(t, surfaces["hello"])
	require.True(t, surfaces["hi"])

	// Match "hello world"
	for _, ar := range ars {
		if ar.PatternTokens[0].Token != "hello" {
			continue
		}
		pr := NewPatternRule(ar.ID, "en", ar.PatternTokens, ar.Description, ar.Message, "")
		toks := []*languagetool.AnalyzedTokenReadings{atr("hello", 0), atr("world", 6)}
		ms, err := pr.Match(languagetool.NewAnalyzedSentence(toks))
		require.NoError(t, err)
		require.Len(t, ms, 1)
	}
}

func TestPatternRuleLoader_OrExpands(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="OR1" name="or rule">
      <pattern>
        <or>
          <token>think</token>
          <token>hope</token>
        </or>
        <token>its</token>
      </pattern>
      <message>it's</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	// Java createRules: base + each OrGroup member → 2 rules
	require.Len(t, ars, 2)
	surfaces := map[string]bool{}
	for _, ar := range ars {
		require.Equal(t, "OR1", ar.ID)
		require.Len(t, ar.PatternTokens, 2)
		require.Equal(t, "its", ar.PatternTokens[1].Token)
		require.False(t, ar.PatternTokens[0].HasOrGroup(), "expanded tokens clear OrGroup")
		surfaces[ar.PatternTokens[0].Token] = true
	}
	require.True(t, surfaces["think"])
	require.True(t, surfaces["hope"])

	// Match either alternative
	for _, ar := range ars {
		pr := NewPatternRule(ar.ID, "en", ar.PatternTokens, ar.Description, ar.Message, "")
		toks := []*languagetool.AnalyzedTokenReadings{
			atr(ar.PatternTokens[0].Token, 0),
			atr("its", 10),
		}
		ms, err := pr.Match(languagetool.NewAnalyzedSentence(toks))
		require.NoError(t, err)
		require.Len(t, ms, 1, "surface %s its", ar.PatternTokens[0].Token)
	}
}

func TestPatternRuleLoader_IdPrefix(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules idprefix="L2_">
  <category>
    <rule id="THAN_AS" name="than as">
      <pattern>
        <token>as</token>
      </pattern>
      <message>use than</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 1)
	require.Equal(t, "L2_THAN_AS", ars[0].ID)
}

func TestPatternRuleLoaderRelaxed(t *testing.T) {
	xml := `<rules><category><rule><pattern><token>x</token></pattern></rule></category></rules>`
	l := NewPatternRuleLoader()
	l.SetRelaxedMode(true)
	rules, err := l.GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, rules, 1)
}

// Java: registered filter classes load; unknown filter classes skip the rule (fail-closed).
func TestPatternRuleLoader_KnownFilterLoaded(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="US" name="underline spaces">
      <pattern>
        <token>foo</token>
      </pattern>
      <filter class="org.languagetool.rules.UnderlineSpacesFilter" args="underlineSpaces:both"/>
      <message>m</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 1)
	require.NotNil(t, ars[0].Filter)
	require.Equal(t, "underlineSpaces:both", ars[0].FilterArgs)
}

func TestPatternRuleLoader_UnknownFilterSkipped(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="BAD" name="unknown filter">
      <pattern>
        <token>foo</token>
      </pattern>
      <filter class="org.languagetool.rules.en.NotPortedYetFilter" args="x:1"/>
      <message>m</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Empty(t, ars, "unsupported filter must not register (would false-fire)")
}

// MultitokenSpellerFilter without dict drops matches (Java empty suggestions → null).
func TestPatternRuleMatcher_MultitokenFilterFailClosed(t *testing.T) {
	// Ensure no default multitoken speller invents hits.
	SetDefaultMultitokenSpeller(nil, nil)
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="MT" name="multitoken">
      <pattern>
        <token>New</token>
        <token>Yrok</token>
      </pattern>
      <filter class="org.languagetool.rules.spelling.multitoken.MultitokenSpellerFilter" args="none:none"/>
      <message>spell</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 1)
	pr := NewPatternRule(ars[0].ID, "en", ars[0].PatternTokens, ars[0].Description, ars[0].Message, "")
	pr.Filter = ars[0].Filter
	pr.FilterArgs = ars[0].FilterArgs
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrTok("New", 0), atrTok("Yrok", 4),
	})
	ms, err := pr.Match(sent)
	require.NoError(t, err)
	require.Empty(t, ms, "no multitoken dict → filter drops (fail-closed)")
}

// Java: rules with <antipattern> load (not skipped); Match suppresses overlapping hits.
func TestPatternRuleLoader_AntiPatternsLoaded(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="REPEAT" name="repeated">
      <pattern>
        <token>go</token>
        <token>go</token>
      </pattern>
      <antipattern>
        <token>to</token>
        <token>go</token>
        <token>go</token>
      </antipattern>
      <message>repeated go</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 1)
	require.Len(t, ars[0].AntiPatterns, 1)
	require.Len(t, ars[0].AntiPatterns[0].Tokens, 3)

	pr := NewPatternRule(ars[0].ID, "en", ars[0].PatternTokens, ars[0].Description, ars[0].Message, "")
	pr.AntiPatterns = ars[0].AntiPatterns

	// "go go" alone → fire
	sentFire := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrTok("go", 0), atrTok("go", 3),
	})
	ms, err := pr.Match(sentFire)
	require.NoError(t, err)
	require.Len(t, ms, 1)

	// "to go go" → antipattern overlaps → suppress
	sentKeep := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{
		atrTok("to", 0), atrTok("go", 3), atrTok("go", 6),
	})
	ms, err = pr.Match(sentKeep)
	require.NoError(t, err)
	require.Empty(t, ms, "antipattern must suppress overlapping rule match")
}

// Java PatternRuleHandler: rulegroup <antipattern> attaches to every child rule.
func TestPatternRuleLoader_RuleGroupAntiPatterns(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rulegroup id="RG" name="group">
      <antipattern>
        <token>safe</token>
        <token>word</token>
      </antipattern>
      <rule>
        <pattern>
          <token>word</token>
          <token>word</token>
        </pattern>
        <message>dup</message>
      </rule>
      <rule id="RG_B">
        <pattern>
          <token>x</token>
        </pattern>
        <message>x</message>
      </rule>
    </rulegroup>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 2)
	require.Equal(t, "RG", ars[0].ID)
	require.Equal(t, "1", ars[0].SubID)
	require.Len(t, ars[0].AntiPatterns, 1)
	require.Len(t, ars[1].AntiPatterns, 1)
	require.Equal(t, "RG_B", ars[1].ID)
}

func TestPatternRuleLoader_ExceptionAndInflected(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="EX1" name="with exception">
      <pattern>
        <token inflected="yes">run<exception>running</exception></token>
        <token>fast</token>
      </pattern>
      <message>x</message>
    </rule>
  </category>
</rules>`
	rules, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	pt := rules[0].PatternTokens[0]
	require.True(t, pt.MatchInflected)
	require.Equal(t, "running", pt.TokenException)
	require.False(t, pt.TokenExceptionRE)

	m := NewPatternTokenMatcher(pt)
	runTok := languagetool.NewAnalyzedToken("run", nil, strPtr("run"))
	runningTok := languagetool.NewAnalyzedToken("running", nil, strPtr("run"))
	// Java isMatched is surface/POS only; exceptions apply via
	// isExceptionMatchedCompletely after any reading matches (IsMatchedReadings).
	require.True(t, m.IsMatched(runTok))
	require.True(t, m.IsMatched(runningTok), "lemma run still matches pattern before exception gate")
	require.True(t, m.IsMatchedReadings(languagetool.NewAnalyzedTokenReadings(runTok)))
	require.False(t, m.IsMatchedReadings(languagetool.NewAnalyzedTokenReadings(runningTok)),
		"surface exception running blocks via isExceptionMatchedCompletely")
}

func TestPatternRuleLoader_PreviousNextException(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category id="C" name="C">
    <rule id="PREV" name="prev">
      <pattern>
        <token>mine<exception scope="previous">not</exception></token>
      </pattern>
      <message>m</message>
      <short>s</short>
    </rule>
    <rule id="NEXT" name="next">
      <pattern>
        <token>can<exception scope="next" regexp="yes">be|do</exception></token>
      </pattern>
      <message>m</message>
      <short>s</short>
    </rule>
  </category>
</rules>`
	rules, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, rules, 2)
	require.Equal(t, "not", rules[0].PatternTokens[0].PreviousException)
	require.Equal(t, "be|do", rules[1].PatternTokens[0].NextException)
	require.True(t, rules[1].PatternTokens[0].NextExceptionRE)
}
