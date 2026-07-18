package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/stretchr/testify/require"
)

func TestFilter_ExclusiveMarkerTargetsFirst(t *testing.T) {
	// Java LET_GO style: <marker>let</marker> + go → FILTER fromPos = "let"
	// (matcher shrinks FromPos/ToPos to InsideMarker tokens).
	let := patterns.NewPatternToken("let", false, false, false)
	let.SetInsideMarker(true)
	goTok := patterns.NewPatternToken("go", false, false, false)
	goTok.SetInsideMarker(false)
	rule := NewDisambiguationPatternRule("LET_GO", "t", "en",
		[]*patterns.PatternToken{let, goTok}, "VB.*", nil, ActionFilter)

	vbTag, nnTag := "VB", "NN"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil)),
		func() *languagetool.AnalyzedTokenReadings {
			r := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("let", &vbTag, nil))
			r.AddReading(languagetool.NewAnalyzedToken("let", &nnTag, nil), "test")
			return r
		}(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("go", &vbTag, nil)),
	}
	pos := 0
	for _, t := range toks {
		t.SetStartPos(pos)
		pos += len(t.GetToken()) + 1
	}
	out := rule.Replace(languagetool.NewAnalyzedSentence(toks))
	var letPOS []string
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok.GetToken() != "let" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				letPOS = append(letPOS, *r.GetPOSTag())
			}
		}
	}
	require.Contains(t, letPOS, "VB")
	require.NotContains(t, letPOS, "NN", "FILTER fromPos (marker) on let should drop NN")
}

func TestFilterAll_UsesEachPatternTokenPOS(t *testing.T) {
	// Java TE_X style: te + <marker postag="WKW:.*|ENM:.*"/> → FILTERALL keeps
	// only readings matching that pattern token POS on the marked token.
	te := patterns.NewPatternToken("te", false, false, false)
	te.SetInsideMarker(false)
	verb := patterns.NewPatternToken("", false, false, false)
	verb.SetPosToken(patterns.PosToken{PosTag: "WKW:.*", Regexp: true})
	verb.SetInsideMarker(true)
	rule := NewDisambiguationPatternRule("TE_X", "t", "nl",
		[]*patterns.PatternToken{te, verb}, "", nil, ActionFilterAll)

	wkw, bnw := "WKW:TGW:INF", "BNW:STL:ONV"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("te", strp2("VZ"), nil)),
		func() *languagetool.AnalyzedTokenReadings {
			r := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("paard", &wkw, nil))
			r.AddReading(languagetool.NewAnalyzedToken("paard", &bnw, nil), "test")
			return r
		}(),
	}
	pos := 0
	for _, t := range toks {
		t.SetStartPos(pos)
		pos += len(t.GetToken()) + 1
	}
	out := rule.Replace(languagetool.NewAnalyzedSentence(toks))
	var tags []string
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok.GetToken() != "paard" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				tags = append(tags, *r.GetPOSTag())
			}
		}
	}
	require.Contains(t, tags, "WKW:TGW:INF")
	require.NotContains(t, tags, "BNW:STL:ONV", "FILTERALL should drop non-matching POS")
}

func TestFilter_MarkerOnSecondToken_JavaFromPos(t *testing.T) {
	// Soft modal style, Java-faithful: will + <marker>run</marker> → FILTER run.
	will := patterns.NewPatternToken("will", false, false, false)
	will.SetInsideMarker(false)
	run := patterns.NewPatternToken("run", false, false, false)
	run.SetInsideMarker(true)
	rule := NewDisambiguationPatternRule("SOFT_WILL_RUN_VB", "t", "en",
		[]*patterns.PatternToken{will, run}, "VB", nil, ActionFilter)

	vbTag, nnTag, mdTag := "VB", "NN", "MD"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("will", &mdTag, nil)),
		func() *languagetool.AnalyzedTokenReadings {
			r := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("run", &vbTag, nil))
			r.AddReading(languagetool.NewAnalyzedToken("run", &nnTag, nil), "test")
			return r
		}(),
	}
	pos := 0
	for _, t := range toks {
		t.SetStartPos(pos)
		pos += len(t.GetToken()) + 1
	}
	out := rule.Replace(languagetool.NewAnalyzedSentence(toks))
	var runPOS []string
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok.GetToken() != "run" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				runPOS = append(runPOS, *r.GetPOSTag())
			}
		}
	}
	require.Contains(t, runPOS, "VB")
	require.NotContains(t, runPOS, "NN", "FILTER fromPos on marked run should drop NN")
}

func strp2(s string) *string { return &s }

func TestRemove_WdPosPartialMatch(t *testing.T) {
	// Java REMOVE_JJ_FOR_OR: remove reading matching <wd pos="JJ"/>
	orTok := patterns.NewPatternToken("or", true, false, false)
	rule := NewDisambiguationPatternRule("REMOVE_JJ_FOR_OR", "t", "en",
		[]*patterns.PatternToken{orTok}, "", nil, ActionRemove)
	jj, cc := "JJ", "CC"
	rule.SetNewInterpretations([]*languagetool.AnalyzedToken{
		languagetool.NewAnalyzedToken("", &jj, nil),
	})
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil)),
		func() *languagetool.AnalyzedTokenReadings {
			r := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("or", &cc, nil))
			r.AddReading(languagetool.NewAnalyzedToken("or", &jj, nil), "dict")
			return r
		}(),
	}
	pos := 0
	for _, t := range toks {
		t.SetStartPos(pos)
		pos += len(t.GetToken()) + 1
	}
	out := rule.Replace(languagetool.NewAnalyzedSentence(toks))
	var tags []string
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok.GetToken() != "or" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				tags = append(tags, *r.GetPOSTag())
			}
		}
	}
	require.Contains(t, tags, "CC")
	require.NotContains(t, tags, "JJ", "REMOVE <wd pos=JJ> should drop JJ")
}

func TestRemove_PostagRegexFromPos(t *testing.T) {
	// Java REMOVE postag="VB.*" on fromPos only
	tok := patterns.NewPatternToken("index", false, false, false)
	rule := NewDisambiguationPatternRule("WHY_S", "t", "en",
		[]*patterns.PatternToken{tok}, "VB.*", nil, ActionRemove)
	vb, nn := "VBZ", "NN"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil)),
		func() *languagetool.AnalyzedTokenReadings {
			r := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("index", &nn, nil))
			r.AddReading(languagetool.NewAnalyzedToken("index", &vb, nil), "dict")
			return r
		}(),
	}
	pos := 0
	for _, t := range toks {
		t.SetStartPos(pos)
		pos += len(t.GetToken()) + 1
	}
	out := rule.Replace(languagetool.NewAnalyzedSentence(toks))
	var tags []string
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok.GetToken() != "index" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				tags = append(tags, *r.GetPOSTag())
			}
		}
	}
	require.Contains(t, tags, "NN")
	require.NotContains(t, tags, "VBZ")
}

func TestDisambigLoader_TokenExceptionBlocksMatch(t *testing.T) {
	// Java: token matches surface unless exception surface matches.
	xml := `<?xml version="1.0"?>
<rules>
  <rule id="EX_RUN" name="run except running">
    <pattern>
      <token>run<exception>running</exception></token>
    </pattern>
    <disambig action="filter" postag="VB"/>
  </rule>
</rules>`
	rules, err := NewDisambiguationRuleLoader().GetRulesFromString(xml, "en", "test")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.Equal(t, "running", rules[0].Tokens[0].TokenException)

	vb, nn := "VB", "NN"
	// "running" must NOT match (exception); leave NN|VB as-is
	running := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("running", &nn, nil))
	running.AddReading(languagetool.NewAnalyzedToken("running", &vb, nil), "dict")
	// "run" matches and FILTER keeps VB
	run := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("run", &nn, nil))
	run.AddReading(languagetool.NewAnalyzedToken("run", &vb, nil), "dict")
	for _, atr := range []*languagetool.AnalyzedTokenReadings{running, run} {
		_ = atr
	}
	pos := 0
	sentStart := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil))
	for _, atr := range []*languagetool.AnalyzedTokenReadings{sentStart, running} {
		atr.SetStartPos(pos)
		pos += len(atr.GetToken()) + 1
	}
	out1 := rules[0].Replace(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{sentStart, running}))
	var tagsRunning []string
	for _, tok := range out1.GetTokensWithoutWhitespace() {
		if tok.GetToken() != "running" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				tagsRunning = append(tagsRunning, *r.GetPOSTag())
			}
		}
	}
	require.Contains(t, tagsRunning, "NN", "exception should block FILTER on running")
	require.Contains(t, tagsRunning, "VB")

	pos = 0
	sentStart2 := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil))
	for _, atr := range []*languagetool.AnalyzedTokenReadings{sentStart2, run} {
		atr.SetStartPos(pos)
		pos += len(atr.GetToken()) + 1
	}
	out2 := rules[0].Replace(languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{sentStart2, run}))
	var tagsRun []string
	for _, tok := range out2.GetTokensWithoutWhitespace() {
		if tok.GetToken() != "run" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				tagsRun = append(tagsRun, *r.GetPOSTag())
			}
		}
	}
	require.Contains(t, tagsRun, "VB")
	require.NotContains(t, tagsRun, "NN", "FILTER should drop NN when pattern matches")
}

func TestDisambigLoader_SkipAndMin(t *testing.T) {
	xml := `<?xml version="1.0"?>
<rules>
  <rule id="SKIP1" name="be skip mine">
    <pattern>
      <token inflected="yes" skip="1">be</token>
      <token marker="yes">mine</token>
    </pattern>
    <disambig action="replace" postag="PRP$"/>
  </rule>
  <rule id="OPT" name="optional middle">
    <pattern>
      <token>very</token>
      <token min="0">well</token>
      <token marker="yes">done</token>
    </pattern>
    <disambig action="filter" postag="JJ"/>
  </rule>
</rules>`
	rules, err := NewDisambiguationRuleLoader().GetRulesFromString(xml, "en", "test")
	require.NoError(t, err)
	require.Len(t, rules, 2)
	require.Equal(t, 1, rules[0].Tokens[0].SkipNext)
	require.Equal(t, 0, rules[1].Tokens[1].MinOccurrence)

	// be + entirely + mine → skip=1 allows "entirely" between
	prp, nn, vb := "PRP$", "NN", "VB"
	md := "MD"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("is", &vb, strp2("be"))),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("entirely", strp2("RB"), nil)),
		func() *languagetool.AnalyzedTokenReadings {
			r := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("mine", &nn, nil))
			r.AddReading(languagetool.NewAnalyzedToken("mine", &prp, nil), "dict")
			r.AddReading(languagetool.NewAnalyzedToken("mine", &vb, nil), "dict")
			return r
		}(),
	}
	_ = md
	pos := 0
	for _, atr := range toks {
		atr.SetStartPos(pos)
		pos += len(atr.GetToken()) + 1
	}
	out := rules[0].Replace(languagetool.NewAnalyzedSentence(toks))
	var tags []string
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok.GetToken() != "mine" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				tags = append(tags, *r.GetPOSTag())
			}
		}
	}
	require.Equal(t, []string{"PRP$"}, tags, "REPLACE fromPos on mine after skip=1")
}


func TestAntiPattern_SuppressesOverlappingReplace(t *testing.T) {
	// Java keepByDisambig: antipattern "not mine" suppresses REPLACE on "mine".
	xml := `<?xml version="1.0"?>
<rules>
  <rule id="MINE_PRP" name="mine">
    <antipattern>
      <token>not</token>
      <token>mine</token>
    </antipattern>
    <pattern>
      <token marker="yes">mine</token>
    </pattern>
    <disambig action="replace" postag="PRP$"/>
  </rule>
</rules>`
	rules, err := NewDisambiguationRuleLoader().GetRulesFromString(xml, "en", "test")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.Len(t, rules[0].AntiPatterns, 1)

	prp, nn := "PRP$", "NN"
	mkMine := func() *languagetool.AnalyzedTokenReadings {
		r := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("mine", &nn, nil))
		r.AddReading(languagetool.NewAnalyzedToken("mine", &prp, nil), "dict")
		return r
	}
	// "is mine" → REPLACE applies
	toksOK := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("is", strp2("VBZ"), nil)),
		mkMine(),
	}
	pos := 0
	for _, atr := range toksOK {
		atr.SetStartPos(pos)
		pos += len(atr.GetToken()) + 1
	}
	outOK := rules[0].Replace(languagetool.NewAnalyzedSentence(toksOK))
	var tagsOK []string
	for _, tok := range outOK.GetTokensWithoutWhitespace() {
		if tok.GetToken() != "mine" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				tagsOK = append(tagsOK, *r.GetPOSTag())
			}
		}
	}
	require.Equal(t, []string{"PRP$"}, tagsOK)

	// "not mine" → antipattern overlaps, leave NN|PRP$
	toksBlock := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("not", strp2("RB"), nil)),
		mkMine(),
	}
	pos = 0
	for _, atr := range toksBlock {
		atr.SetStartPos(pos)
		pos += len(atr.GetToken()) + 1
	}
	outBlock := rules[0].Replace(languagetool.NewAnalyzedSentence(toksBlock))
	var tagsBlock []string
	for _, tok := range outBlock.GetTokensWithoutWhitespace() {
		if tok.GetToken() != "mine" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				tagsBlock = append(tagsBlock, *r.GetPOSTag())
			}
		}
	}
	require.Contains(t, tagsBlock, "NN")
	require.Contains(t, tagsBlock, "PRP$")
}

func TestReplace_WdLemmaPos(t *testing.T) {
	// Java ca/n't style: <marker>ca</marker> + n't → replace wd lemma=can pos=MD
	xml := `<?xml version="1.0"?>
<rules>
  <rule id="CANT_MD" name="ca n't">
    <pattern>
      <token marker="yes">ca</token>
      <token spacebefore="no">n't</token>
    </pattern>
    <disambig action="replace">
      <wd lemma="can" pos="MD"/>
    </disambig>
  </rule>
  <rule id="NY_NNP" name="New York">
    <pattern>
      <token>New</token>
      <token>York</token>
    </pattern>
    <disambig action="replace">
      <wd pos="NNP"/>
      <wd pos="NNP"/>
    </disambig>
  </rule>
</rules>`
	rules, err := NewDisambiguationRuleLoader().GetRulesFromString(xml, "en", "test")
	require.NoError(t, err)
	require.Len(t, rules, 2)
	require.Len(t, rules[0].NewTokenReadings, 1)
	require.Len(t, rules[1].NewTokenReadings, 2)

	md, nn := "MD", "NN"
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil)),
		func() *languagetool.AnalyzedTokenReadings {
			r := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("ca", &nn, nil))
			return r
		}(),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("n't", strp2("RB"), nil)),
	}
	// spacebefore=no on n't — set whitespace before false
	toks[2].SetWhitespaceBefore(false)
	pos := 0
	for _, atr := range toks {
		atr.SetStartPos(pos)
		pos += len(atr.GetToken()) + 1
	}
	out := rules[0].Replace(languagetool.NewAnalyzedSentence(toks))
	var tags []string
	var lemma string
	for _, tok := range out.GetTokensWithoutWhitespace() {
		if tok.GetToken() != "ca" {
			continue
		}
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil {
				tags = append(tags, *r.GetPOSTag())
			}
			if r != nil && r.GetLemma() != nil {
				lemma = *r.GetLemma()
			}
		}
	}
	require.Equal(t, []string{"MD"}, tags)
	require.Equal(t, "can", lemma)
	_ = md

	// multi-token REPLACE both NNP
	nnp := "NNP"
	toks2 := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", strp2(languagetool.SentenceStartTagName), nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("New", &nn, nil)),
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("York", &nn, nil)),
	}
	pos = 0
	for _, atr := range toks2 {
		atr.SetStartPos(pos)
		pos += len(atr.GetToken()) + 1
	}
	out2 := rules[1].Replace(languagetool.NewAnalyzedSentence(toks2))
	for _, want := range []string{"New", "York"} {
		var ts []string
		for _, tok := range out2.GetTokensWithoutWhitespace() {
			if tok.GetToken() != want {
				continue
			}
			for _, r := range tok.GetReadings() {
				if r != nil && r.GetPOSTag() != nil {
					ts = append(ts, *r.GetPOSTag())
				}
			}
		}
		require.Equal(t, []string{nnp}, ts, want)
	}
}
