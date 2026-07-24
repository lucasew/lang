package br

import (
	"bufio"
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	brtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/br"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

//go:embed data/topo.txt
var topoFS embed.FS

var (
	topoOnce sync.Once
	// wrongWords[n] maps (n+1)-word French place keys → Breton suggestions (Java loadWords).
	topoWrongWords []map[string]string
)

func loadTopoWords() []map[string]string {
	topoOnce.Do(func() {
		f, err := topoFS.Open("data/topo.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		var list []map[string]string
		// Java: Tokenizer wordTokenizer = new Breton().getWordTokenizer();
		wt := brtok.NewBretonWordTokenizer()
		sc := bufio.NewScanner(f)
		// Raise scanner buffer for long lines if needed
		for sc.Scan() {
			// Java: line = line.trim(); empty / # comments skipped
			line := tools.JavaStringTrim(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			// Java: line.split("=") requires exactly 2 parts (not SplitN limit-2)
			parts := strings.Split(line, "=")
			if len(parts) != 2 {
				// Java throws IOException; fail-closed skip (no invent partial key)
				continue
			}
			wrongForms := strings.Split(parts[0], "|")
			sug := parts[1] // Java puts parts[1] as-is (no extra trim)
			for _, wrongForm := range wrongForms {
				// Java does not trim wrongForm; count non-whitespace tokenizer tokens
				wordCount := 0
				for _, tok := range wt.Tokenize(wrongForm) {
					if !tools.IsWhitespace(tok) {
						wordCount++
					}
				}
				if wordCount == 0 {
					continue
				}
				for len(list) < wordCount {
					list = append(list, map[string]string{})
				}
				list[wordCount-1][wrongForm] = sug
			}
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
		topoWrongWords = list
	})
	return topoWrongWords
}

// TopoReplaceRule ports org.languagetool.rules.br.TopoReplaceRule.
// Case-sensitive multiword longest-first Match (Java isCaseSensitive() == true).
type TopoReplaceRule struct {
	Messages map[string]string
	Category *rules.Category
}

func NewTopoReplaceRule(messages map[string]string) *TopoReplaceRule {
	_ = loadTopoWords()
	return &TopoReplaceRule{
		Messages: messages,
		Category: rules.CatMisc.GetCategory(messages),
	}
}

func (r *TopoReplaceRule) GetID() string { return "BR_TOPO" }

func (r *TopoReplaceRule) GetDescription() string {
	return "anvioù-lec’h e brezhoneg"
}

func (r *TopoReplaceRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// Match ports TopoReplaceRule.match.
func (r *TopoReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	wrongWords := loadTopoWords()
	if len(wrongWords) == 0 {
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	capQ := len(wrongWords)
	prev := make([]*languagetool.AnalyzedTokenReadings, 0, capQ)
	var ruleMatches []*rules.RuleMatch

	for i := 1; i < len(tokens); i++ {
		if len(prev) == capQ {
			prev = prev[1:]
		}
		prev = append(prev, tokens[i])

		// Longest phrase first: windows starting at offset 0..len-1 of prev.
		for start := 0; start < len(prev); start++ {
			if prev[start] == nil || prev[start].IsImmunized() {
				continue
			}
			var b strings.Builder
			for k := start; k < len(prev); k++ {
				if k > start && prev[k] != nil && prev[k].IsWhitespaceBefore() {
					b.WriteByte(' ')
				}
				if prev[k] != nil {
					b.WriteString(prev[k].GetToken())
				}
			}
			crt := b.String()
			crtWordCount := len(prev) - start
			if crtWordCount < 1 || crtWordCount > len(wrongWords) {
				continue
			}
			// Java isCaseSensitive: exact key
			crtMatch, ok := wrongWords[crtWordCount-1][crt]
			if !ok {
				continue
			}
			replacements := strings.Split(crtMatch, "|")
			clean := make([]string, 0, len(replacements))
			for _, rep := range replacements {
				rep = tools.JavaStringTrim(rep)
				if rep != "" {
					clean = append(clean, rep)
				}
			}
			msg := crt + " zo un anv lec’h gallek. Ha fellout a rae deoc’h skrivañ "
			for k, rep := range clean {
				if k > 0 {
					if k == len(clean)-1 {
						msg += " pe "
					} else {
						msg += ", "
					}
				}
				msg += "<suggestion>" + rep + "</suggestion>"
			}
			msg += "?"
			startPos := prev[start].GetStartPos()
			endPos := prev[len(prev)-1].GetEndPos()
			rm := rules.NewRuleMatch(r, sentence, startPos, endPos, msg)
			rm.ShortMessage = "anvioù lec’h"
			rm.SetSuggestedReplacements(clean)
			ruleMatches = append(ruleMatches, rm)
			break // first (longest) hit only, like Java
		}
	}
	return ruleMatches
}
