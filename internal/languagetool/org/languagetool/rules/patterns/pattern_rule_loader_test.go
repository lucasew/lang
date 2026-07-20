package patterns

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
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

func TestPatternRuleLoader_DefaultTempOff(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category>
    <rule id="T1" name="temp" default="temp_off">
      <pattern><token>x</token></pattern>
      <message>m</message>
    </rule>
    <rule id="O1" name="off" default="off">
      <pattern><token>y</token></pattern>
      <message>m</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 2)
	byID := map[string]*AbstractPatternRule{}
	for _, ar := range ars {
		byID[ar.ID] = ar
	}
	require.True(t, byID["T1"].DefaultOff)
	require.True(t, byID["T1"].DefaultTempOff)
	require.True(t, byID["O1"].DefaultOff)
	require.False(t, byID["O1"].DefaultTempOff)
}

// Java min_prev_matches / distance_tokens on rulegroup (style REP_* rules).
func TestPatternRuleLoader_MinPrevMatchesDistanceTokens(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category id="STYLE" name="Style">
    <rulegroup id="REP_PASSIVE" name="rep" min_prev_matches="4" distance_tokens="80" tags="picky">
      <rule>
        <pattern><token>was</token><token>done</token></pattern>
        <message>passive</message>
      </rule>
    </rulegroup>
    <rule id="PLAIN" min_prev_matches="2">
      <pattern><token>x</token></pattern>
      <message>m</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "style.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 2)
	byID := map[string]*AbstractPatternRule{}
	for _, ar := range ars {
		byID[ar.ID] = ar
	}
	require.Equal(t, 4, byID["REP_PASSIVE"].MinPrevMatches)
	require.Equal(t, 80, byID["REP_PASSIVE"].DistanceTokens)
	require.Equal(t, []rules.Tag{rules.TagPicky}, byID["REP_PASSIVE"].Tags)
	require.Equal(t, 2, byID["PLAIN"].MinPrevMatches)
	require.Equal(t, 0, byID["PLAIN"].DistanceTokens)

	// Transformer marks min_prev_matches > 0 for text-level wrap
	tr := NewRepeatedPatternRuleTransformer("en")
	out := tr.Transform(ars)
	require.NotEmpty(t, out.GetTransformedRules())
	require.Len(t, out.GetRemainingRules(), 0) // both PLAIN and REP have min_prev>0
}

// Java RepeatedPatternRule: only emit after min_prev_matches prior hits in text.
func TestRepeatedPatternRule_MatchSentences(t *testing.T) {
	// min_prev_matches=2 → third occurrence in document is first reported
	pr := NewPatternRule("REP_X", "en", []*PatternToken{Token("foo")}, "rep", "again", "")
	pr.MinPrevMatches = 2
	pr.DistanceTokens = 100
	rep := &RepeatedPatternRule{
		LanguageCode:             "en",
		PatternRules:             []*PatternRule{pr},
		DefaultMaxDistanceTokens: 60,
	}
	// three sentences each with "foo"
	s1 := languagetool.AnalyzePlain("foo.")
	s2 := languagetool.AnalyzePlain("foo.")
	s3 := languagetool.AnalyzePlain("foo.")
	ms := rep.MatchSentences([]*languagetool.AnalyzedSentence{s1, s2, s3})
	require.NotEmpty(t, ms, "third hit should fire after min_prev=2")
	// only one match (the third occurrence)
	require.Len(t, ms, 1)
	// document offset: "foo." + "foo." + "foo." → third at char 8
	require.Equal(t, 8, ms[0].FromPos)
}

// Java ConsistencyPatternRule: report minority features when counts differ.
func TestConsistencyPatternRule_MatchSentences(t *testing.T) {
	// PREFIX_GROUP_us matches "color"; PREFIX_GROUP_uk matches "colour"
	// Document: color color colour → us=2, uk=1 → report minority "colour"
	us := NewPatternRule("PREFIXFORCONSISTENCYRULES_SPELL_us", "en",
		[]*PatternToken{Token("color")}, "spell", "US spelling", "")
	uk := NewPatternRule("PREFIXFORCONSISTENCYRULES_SPELL_uk", "en",
		[]*PatternToken{Token("colour")}, "spell", "UK spelling", "")
	consist := &ConsistencyPatternRule{
		MainID:       "PREFIXFORCONSISTENCYRULES_SPELL",
		LanguageCode: "en",
		PatternRules: []*PatternRule{us, uk},
	}
	s1 := languagetool.AnalyzePlain("color here.")
	s2 := languagetool.AnalyzePlain("color again.")
	s3 := languagetool.AnalyzePlain("colour odd.")
	ms := consist.MatchSentences([]*languagetool.AnalyzedSentence{s1, s2, s3})
	require.Len(t, ms, 1, "minority UK feature should be reported once")
	require.Equal(t, "PREFIXFORCONSISTENCYRULES_SPELL_uk", ms[0].RuleID)
	// "color here." (11) + "color again." (12) = 23 → colour at 23
	require.Equal(t, 23, ms[0].FromPos)

	// single feature only → no inconsistency
	ms2 := consist.MatchSentences([]*languagetool.AnalyzedSentence{s1, s2})
	require.Empty(t, ms2)

	// tie (1 us, 1 uk) → report all
	ms3 := consist.MatchSentences([]*languagetool.AnalyzedSentence{s1, s3})
	require.Len(t, ms3, 2)
}

// RegisterGrammarXML wires consistency-prefix rules as text-level checkers.
func TestRegisterGrammarXML_ConsistencyPatternRules(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules lang="en">
  <category id="STYLE" name="Style">
    <rule id="PREFIXFORCONSISTENCYRULES_SPELL_us">
      <pattern><token>color</token></pattern>
      <message>US spelling</message>
    </rule>
    <rule id="PREFIXFORCONSISTENCYRULES_SPELL_uk">
      <pattern><token>colour</token></pattern>
      <message>UK spelling</message>
    </rule>
    <rule id="PLAIN">
      <pattern><token>zzz</token></pattern>
      <message>plain</message>
    </rule>
  </category>
</rules>`
	lt := languagetool.NewJLanguageTool("en")
	n, err := RegisterGrammarXML(lt, xml, "t.xml", "en")
	require.NoError(t, err)
	require.Equal(t, 2, n) // one consistency group + PLAIN
	ids := lt.GetAllRegisteredRuleIDs()
	require.Contains(t, ids, "PREFIXFORCONSISTENCYRULES_SPELL")
	require.Contains(t, ids, "PLAIN")
	// disable PLAIN noise
	lt.DisableRule("PLAIN")
	ms := lt.Check("color color colour")
	require.NotEmpty(t, ms, "inconsistency should fire via text-level checker")
	// minority feature rule id preserved on match
	found := false
	for _, m := range ms {
		if m.RuleID == "PREFIXFORCONSISTENCYRULES_SPELL_uk" {
			found = true
		}
	}
	require.True(t, found, "expected UK minority match, got %#v", ms)
}

// Java premium= inheritance: rule > rulegroup > category > file (yes/no).
func TestPatternRuleLoader_PremiumInheritance(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules lang="en" premium="yes">
  <category id="C1" name="C1" premium="no">
    <rule id="CAT_NO">
      <pattern><token>a</token></pattern>
      <message>m</message>
    </rule>
    <rule id="RULE_YES" premium="yes">
      <pattern><token>b</token></pattern>
      <message>m</message>
    </rule>
  </category>
  <category id="C2" name="C2">
    <rulegroup id="G" premium="yes">
      <rule id="GROUP_YES">
        <pattern><token>c</token></pattern>
        <message>m</message>
      </rule>
      <rule id="GROUP_NO" premium="no">
        <pattern><token>d</token></pattern>
        <message>m</message>
      </rule>
    </rulegroup>
    <rule id="FILE_YES">
      <pattern><token>e</token></pattern>
      <message>m</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	byID := map[string]*AbstractPatternRule{}
	for _, ar := range ars {
		byID[ar.ID] = ar
	}
	require.False(t, byID["CAT_NO"].Premium, "category premium=no overrides file yes")
	require.True(t, byID["RULE_YES"].Premium, "rule premium=yes wins")
	require.True(t, byID["GROUP_YES"].Premium, "rulegroup premium=yes")
	require.False(t, byID["GROUP_NO"].Premium, "rule premium=no overrides group")
	require.True(t, byID["FILE_YES"].Premium, "file premium=yes when nothing else set")

	lt := languagetool.NewJLanguageTool("en")
	_, err = RegisterGrammarXML(lt, xml, "t.xml", "en")
	require.NoError(t, err)
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		if id != "RULE_YES" {
			lt.DisableRule(id)
		}
	}
	ms := lt.Check("b here")
	require.NotEmpty(t, ms)
	require.True(t, ms[0].IsPremium)
}

// Java prio= inheritance: category then group then rule (non-zero overwrites).
// Twin of grammar-withPrio.xml expectations.
func TestPatternRuleLoader_PrioInheritance(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category id="CAT-PRIO-5" name="CAT-PRIO-5" prio="5">
    <rulegroup id="CAT-PRIO-5-RG-PRIO-10" name="g" prio="10">
      <rule id="CAT-PRIO-5-RG-PRIO-10-R-PRIO-15" prio="15">
        <pattern><token>fake1</token></pattern>
        <message>msg1</message>
      </rule>
      <rule id="CAT-PRIO-5-RG-PRIO-10-R-PRIO-0">
        <pattern><token>fake1</token></pattern>
        <message>msg3</message>
      </rule>
    </rulegroup>
    <rulegroup id="CAT-PRIO-5-RG-PRIO-0" name="g0">
      <rule id="CAT-PRIO-5-RG-PRIO-0-R-PRIO-0">
        <pattern><token>fake1</token></pattern>
        <message>msg3</message>
      </rule>
    </rulegroup>
  </category>
  <category id="CAT-PRIO-0" name="CAT-PRIO-0">
    <rule id="CAT-PRIO-0-R-PRIO-0" name="n">
      <pattern><token>fake1</token></pattern>
      <message>msg1</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "grammar-withPrio.xml", "xx")
	require.NoError(t, err)
	byID := map[string]*AbstractPatternRule{}
	for _, ar := range ars {
		byID[ar.ID] = ar
	}
	require.Equal(t, 15, byID["CAT-PRIO-5-RG-PRIO-10-R-PRIO-15"].Priority)
	require.Equal(t, 10, byID["CAT-PRIO-5-RG-PRIO-10-R-PRIO-0"].Priority)
	require.Equal(t, 5, byID["CAT-PRIO-5-RG-PRIO-0-R-PRIO-0"].Priority)
	require.Equal(t, 0, byID["CAT-PRIO-0-R-PRIO-0"].Priority)

	// Register → LocalMatch.Priority for CleanOverlapping
	lt := languagetool.NewJLanguageTool("xx")
	_, err = RegisterGrammarXML(lt, xml, "grammar-withPrio.xml", "xx")
	require.NoError(t, err)
	// Enable only the prio-15 rule for a single match
	for _, id := range lt.GetAllRegisteredRuleIDs() {
		if id != "CAT-PRIO-5-RG-PRIO-10-R-PRIO-15" {
			lt.DisableRule(id)
		}
	}
	ms := lt.Check("fake1")
	require.NotEmpty(t, ms)
	require.Equal(t, 15, ms[0].Priority)
}

// Java type inheritance: rule > rulegroup > category; url: rule else rulegroup.
func TestPatternRuleLoader_TypeAndURLInheritance(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category id="C" name="Cat" type="misspelling">
    <rule id="R1" name="rule type" type="grammar">
      <pattern><token>a</token></pattern>
      <message>m</message>
      <url>https://example.com/rule</url>
    </rule>
    <rulegroup id="G" name="g" type="typographical">
      <url>https://example.com/group</url>
      <rule>
        <pattern><token>b</token></pattern>
        <message>m</message>
      </rule>
    </rulegroup>
    <rule id="R2" name="cat type only">
      <pattern><token>c</token></pattern>
      <message>m</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "/path/to/grammar.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 3)
	byID := map[string]*AbstractPatternRule{}
	for _, ar := range ars {
		byID[ar.ID] = ar
		require.Equal(t, "/path/to/grammar.xml", ar.SourceFile)
	}
	require.Equal(t, "grammar", byID["R1"].IssueType)
	require.Equal(t, "https://example.com/rule", byID["R1"].URL)
	require.Equal(t, "typographical", byID["G"].IssueType)
	require.Equal(t, "https://example.com/group", byID["G"].URL)
	require.Equal(t, "misspelling", byID["R2"].IssueType)
	require.Empty(t, byID["R2"].URL)
}

// Java category default="off" → Category.isDefaultOff; rules stay default-on until ignoreRule.
func TestPatternRuleLoader_CategoryDefaultOff(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category id="PLAIN_ENGLISH" name="Plain English" type="style" default="off">
    <rule id="PE1" name="plain">
      <pattern><token>utilize</token></pattern>
      <message>use</message>
    </rule>
  </category>
  <category id="GRAMMAR" name="Grammar" type="grammar" default="on">
    <rule id="G1" name="g">
      <pattern><token>foo</token></pattern>
      <message>m</message>
    </rule>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 2)
	byID := map[string]*AbstractPatternRule{}
	for _, ar := range ars {
		byID[ar.ID] = ar
	}
	require.True(t, byID["PE1"].CategoryDefaultOff)
	require.False(t, byID["PE1"].DefaultOff, "category default-off does not set rule.defaultOff")
	require.Equal(t, "style", byID["PE1"].CategoryType)
	require.False(t, byID["G1"].CategoryDefaultOff)
	require.Equal(t, "grammar", byID["G1"].CategoryType)

	lt := languagetool.NewJLanguageTool("en")
	n, err := RegisterGrammarXML(lt, xml, "t.xml", "en")
	require.NoError(t, err)
	require.Equal(t, 2, n)
	// PE1 inactive via category; G1 active
	require.Empty(t, lt.Check("I utilize tools"))
	require.NotEmpty(t, lt.Check("foo bar"))
	// Enable category → PE1 fires
	lt.EnableCategory("PLAIN_ENGLISH")
	ms := lt.Check("I utilize tools")
	require.NotEmpty(t, ms)
	require.Equal(t, "PE1", ms[0].RuleID)
	require.Equal(t, "style", ms[0].IssueType)
}

// Java rulegroup default="off" / "temp_off" is inherited by every child rule.
func TestPatternRuleLoader_RuleGroupDefaultInherited(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <category id="C" name="Cat">
    <rulegroup id="G_OFF" name="off group" default="off">
      <rule>
        <pattern><token>a</token></pattern>
        <message>m</message>
      </rule>
      <rule>
        <pattern><token>b</token></pattern>
        <message>m</message>
      </rule>
    </rulegroup>
    <rulegroup id="G_TEMP" name="temp group" default="temp_off">
      <rule>
        <pattern><token>c</token></pattern>
        <message>m</message>
      </rule>
    </rulegroup>
    <rulegroup id="G_ON" name="on group">
      <rule>
        <pattern><token>d</token></pattern>
        <message>m</message>
      </rule>
    </rulegroup>
  </category>
</rules>`
	ars, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, ars, 4)
	// group children share group id when rule id empty
	offCount, tempCount, onCount := 0, 0, 0
	for _, ar := range ars {
		if ar.DefaultTempOff {
			tempCount++
			require.True(t, ar.DefaultOff)
		} else if ar.DefaultOff {
			offCount++
		} else {
			onCount++
		}
	}
	require.Equal(t, 2, offCount, "G_OFF children")
	require.Equal(t, 1, tempCount, "G_TEMP child")
	require.Equal(t, 1, onCount, "G_ON child stays default-on")
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

// Java PatternRule.checkForAntiPatterns: IMMUNIZE via antipattern then rematch.
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
    <rule id="MULTI_PREV" name="multi">
      <pattern>
        <token>that
          <exception scope="previous" postag="SENT_START"/>
          <exception scope="previous">nor</exception>
          <exception scope="previous" regexp="yes">buzz|word</exception>
        </token>
      </pattern>
      <message>m</message>
      <short>s</short>
    </rule>
    <rule id="PREV_NEG_POS" name="negpos">
      <pattern>
        <token>one
          <exception scope="previous" postag="CD" negate="yes">one</exception>
        </token>
      </pattern>
      <message>m</message>
      <short>s</short>
    </rule>
  </category>
</rules>`
	rules, err := NewPatternRuleLoader().GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.Len(t, rules, 4)
	require.Equal(t, "not", rules[0].PatternTokens[0].PreviousException)
	require.Len(t, rules[0].PatternTokens[0].PreviousExceptions, 1)
	require.Equal(t, "be|do", rules[1].PatternTokens[0].NextException)
	require.True(t, rules[1].PatternTokens[0].NextExceptionRE)
	require.Len(t, rules[1].PatternTokens[0].NextExceptions, 1)

	// Multi previous exceptions (Java rareFields.previousExceptions)
	multi := rules[2].PatternTokens[0]
	require.Len(t, multi.PreviousExceptions, 3)
	require.Equal(t, "SENT_START", multi.PreviousExceptions[0].Pos.PosTag)
	require.Equal(t, "nor", multi.PreviousExceptions[1].Token)
	require.True(t, multi.PreviousExceptions[2].Regexp)

	// previous with surface + postag + negate (Java PatternToken exception isMatched)
	negPrev := rules[3].PatternTokens[0]
	require.Len(t, negPrev.PreviousExceptions, 1)
	ex := negPrev.PreviousExceptions[0]
	require.Equal(t, "one", ex.Token)
	require.True(t, ex.Negation)
	require.Equal(t, "CD", ex.Pos.PosTag)
	require.False(t, ex.Pos.Negate)
}

// Java isMatchedByPreviousException: any previous exception PatternToken.isMatched.
func TestPatternTokenMatcher_MultiPreviousException(t *testing.T) {
	pt := NewPatternToken("that", false, false, false)
	pt.AddPreviousException(NewPatternToken("nor", false, false, false))
	sentStart := NewPatternToken("", false, false, false)
	sentStart.SetPosToken(PosToken{PosTag: "SENT_START", Regexp: false, Negate: false})
	pt.AddPreviousException(sentStart)

	m := NewPatternTokenMatcher(pt)
	// previous surface "nor" → exception fires
	prevNor := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("nor", strPtr("CC"), strPtr("nor")),
	)
	require.True(t, m.IsMatchedByPreviousException(prevNor))
	// previous with SENT_START POS
	prevSS := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("", strPtr("SENT_START"), strPtr("")),
	)
	require.True(t, m.IsMatchedByPreviousException(prevSS))
	// unrelated previous
	prevOther := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("hello", strPtr("NN"), strPtr("hello")),
	)
	require.False(t, m.IsMatchedByPreviousException(prevOther))

	// negate=yes surface: exception matches when surface does NOT equal "one"
	// Java: exception.setNegation(true) → isMatched flips surface match
	pt2 := NewPatternToken("x", false, false, false)
	exNeg := NewPatternToken("one", false, false, false)
	exNeg.SetNegation(true)
	exNeg.SetPosToken(PosToken{PosTag: "CD", Regexp: false, Negate: false})
	pt2.AddPreviousException(exNeg)
	m2 := NewPatternTokenMatcher(pt2)
	// reading "two"/CD: surface != one (negation → true) AND pos CD → exception fires
	two := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("two", strPtr("CD"), strPtr("two")),
	)
	require.True(t, m2.IsMatchedByPreviousException(two))
	// reading "one"/CD: surface matches, negation flips → surface false → no match
	one := languagetool.NewAnalyzedTokenReadings(
		languagetool.NewAnalyzedToken("one", strPtr("CD"), strPtr("one")),
	)
	require.False(t, m2.IsMatchedByPreviousException(one))
}

func TestPatternRuleLoader_ToneTagsAndPicky(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules lang="en">
  <category id="C" name="Cat" tags="picky" tone_tags="clarity">
    <rule id="R1" name="n" tags="picky" tone_tags="formal" is_goal_specific="true">
      <pattern><token>foo</token></pattern>
      <message>m</message>
    </rule>
  </category>
</rules>`
	loader := NewPatternRuleLoader()
	ars, err := loader.GetRulesFromString(xml, "t.xml", "en")
	require.NoError(t, err)
	require.NotEmpty(t, ars)
	require.Contains(t, ars[0].Tags, rules.TagPicky)
	require.Contains(t, ars[0].ToneTags, languagetool.ToneFormal)
	require.Contains(t, ars[0].ToneTags, languagetool.ToneClarity)
	require.True(t, ars[0].GoalSpecific)

	lt := languagetool.NewJLanguageTool("en")
	n, err := RegisterGrammarXML(lt, xml, "t.xml", "en")
	require.NoError(t, err)
	require.Equal(t, 1, n)
	// picky filtered at DEFAULT
	require.Empty(t, lt.Check("foo bar"))
	lt.Level = languagetool.LevelPicky
	// still goal-specific under empty tone set
	require.Empty(t, lt.Check("foo bar"))
	lt.SetToneTags(languagetool.ToneFormal)
	require.NotEmpty(t, lt.Check("foo bar"))
}


