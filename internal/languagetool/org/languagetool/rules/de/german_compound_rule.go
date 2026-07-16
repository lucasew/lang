package de

import (
	"embed"
	"io"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

//go:embed data/compounds.txt data/compound-cities.txt
var compoundsFS embed.FS

var (
	deCompoundOnce sync.Once
	deCompoundData *rules.CompoundRuleData
	chCompoundOnce sync.Once
	chCompoundData *rules.CompoundRuleData
)

func mustLoadCompoundData(expander rules.LineExpander) *rules.CompoundRuleData {
	f1, err := compoundsFS.Open("data/compounds.txt")
	if err != nil {
		panic(err)
	}
	defer f1.Close()
	f2, err := compoundsFS.Open("data/compound-cities.txt")
	if err != nil {
		panic(err)
	}
	defer f2.Close()
	d, err := rules.NewCompoundRuleDataMulti(expander, []io.Reader{f1, f2}, []string{"/de/compounds.txt", "/de/compound-cities.txt"})
	if err != nil {
		panic(err)
	}
	return d
}

func loadDECompoundData() *rules.CompoundRuleData {
	deCompoundOnce.Do(func() {
		deCompoundData = mustLoadCompoundData(nil)
	})
	return deCompoundData
}

func loadCHCompoundData() *rules.CompoundRuleData {
	chCompoundOnce.Do(func() {
		// SwissExpander: accept ß and ss variants
		expander := func(line string) []string {
			if strings.Contains(line, "ß") {
				return []string{line, strings.ReplaceAll(line, "ß", "ss")}
			}
			return []string{line}
		}
		chCompoundData = mustLoadCompoundData(expander)
	})
	return chCompoundData
}

// GermanCompoundRule ports org.languagetool.rules.de.GermanCompoundRule.
type GermanCompoundRule struct {
	*rules.AbstractCompoundRule
}

func NewGermanCompoundRule(messages map[string]string) *GermanCompoundRule {
	base := &rules.AbstractCompoundRule{
		Messages:                    messages,
		ID:                          "DE_COMPOUNDS",
		Description:                 "Zusammenschreibung von Wörtern, z. B. 'CD-ROM' statt 'CD ROM'",
		WithHyphenMessage:           "Dieses Wort wird mit Bindestrich geschrieben.",
		WithoutHyphenMessage:        "Dieses Wort wird zusammengeschrieben.",
		WithOrWithoutHyphenMessage:  "Diese Wörter werden zusammengeschrieben oder mit Bindestrich getrennt.",
		ShortDesc:                   "Zusammenschreibung von Wörtern",
		SentenceStartsWithUpperCase: true,
		Data:                        loadDECompoundData(),
	}
	return &GermanCompoundRule{AbstractCompoundRule: base}
}

func (r *GermanCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	immunizeDECompoundAntiPatterns(sentence)
	return r.AbstractCompoundRule.Match(sentence)
}

// SwissCompoundRule ports org.languagetool.rules.de.SwissCompoundRule.
type SwissCompoundRule struct {
	*rules.AbstractCompoundRule
}

func NewSwissCompoundRule(messages map[string]string) *SwissCompoundRule {
	base := &rules.AbstractCompoundRule{
		Messages:                    messages,
		ID:                          "DE_CH_COMPOUNDS",
		Description:                 "Zusammenschreibung von Wörtern, z. B. 'CD-ROM' statt 'CD ROM'",
		WithHyphenMessage:           "Dieses Wort wird mit Bindestrich geschrieben.",
		WithoutHyphenMessage:        "Dieses Wort wird zusammengeschrieben.",
		WithOrWithoutHyphenMessage:  "Diese Wörter werden zusammengeschrieben oder mit Bindestrich getrennt.",
		ShortDesc:                   "Zusammenschreibung von Wörtern",
		SentenceStartsWithUpperCase: true,
		Data:                        loadCHCompoundData(),
	}
	return &SwissCompoundRule{AbstractCompoundRule: base}
}

func (r *SwissCompoundRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	immunizeDECompoundAntiPatterns(sentence)
	return r.AbstractCompoundRule.Match(sentence)
}

// Light anti-patterns for GermanCompoundRuleTest good cases (subset of Java ANTI_PATTERNS).
func immunizeDECompoundAntiPatterns(sentence *languagetool.AnalyzedSentence) {
	tokens := sentence.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		t := tokens[i].GetToken()
		// an|um + die + digits
		if (t == "an" || t == "um" || t == "An" || t == "Um") && i+2 < len(tokens) {
			if equalFoldASCII(tokens[i+1].GetToken(), "die") && isDigits(tokens[i+2].GetToken()) {
				tokens[i].Immunize(0)
				tokens[i+1].Immunize(0)
				tokens[i+2].Immunize(0)
			}
		}
		// rund|etwa|... + digits
		switch strings.ToLower(t) {
		case "rund", "etwa", "zirka", "cirka", "ungefähr", "annähernd", "grob", "wohl", "gegen", "schätzungsweise":
			if i+1 < len(tokens) && isDigits(tokens[i+1].GetToken()) {
				tokens[i].Immunize(0)
				tokens[i+1].Immunize(0)
			}
		}
		// ca . digits
		if equalFoldASCII(t, "ca") && i+2 < len(tokens) && tokens[i+1].GetToken() == "." && isDigits(tokens[i+2].GetToken()) {
			tokens[i].Immunize(0)
			tokens[i+1].Immunize(0)
			tokens[i+2].Immunize(0)
		}
		// von|vom ... aus gedacht
		if equalFoldASCII(t, "aus") && i+1 < len(tokens) && equalFoldASCII(tokens[i+1].GetToken(), "gedacht") {
			tokens[i].Immunize(0)
			tokens[i+1].Immunize(0)
		}
	}
}

func equalFoldASCII(a, b string) bool {
	return strings.EqualFold(a, b)
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
