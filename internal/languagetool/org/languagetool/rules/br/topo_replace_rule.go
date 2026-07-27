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

// BretonTopo ports TopoReplaceRule.BRETON_TOPO.
const BretonTopo = "BR_TOPO"

// Java static finals (FILE_NAME / FILE_ENCODING / BR_LOCALE).
const (
	topoFileName = "/br/topo.txt"
	topoLocaleBR = "br"
)

var (
	topoOnce sync.Once
	// wrongWords[n] maps (n+1)-word French place keys → Breton suggestions (Java loadWords).
	topoWrongWords []map[string]string
)

// loadTopoWords ports TopoReplaceRule.loadWords(getFromRulesDirAsStream("/br/topo.txt")).
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
		for sc.Scan() {
			// Java: line = line.trim(); empty / # comments skipped
			line := tools.JavaStringTrim(sc.Text())
			if line == "" || line[0] == '#' {
				continue
			}
			// Java: line.split("=") requires exactly 2 parts
			parts := strings.Split(line, "=")
			if len(parts) != 2 {
				// Java throws IOException → RuntimeException
				panic("Format error in file " + topoFileName + ", line: " + line)
			}
			wrongForms := strings.Split(parts[0], "|") // multiple incorrect forms
			sug := parts[1]
			for _, wrongForm := range wrongForms {
				wordCount := 0
				for _, tok := range wt.Tokenize(wrongForm) {
					if !tools.IsWhitespace(tok) {
						wordCount++
					}
				}
				// grow if necessary
				for len(list) < wordCount {
					list = append(list, map[string]string{})
				}
				list[wordCount-1][wrongForm] = sug
			}
		}
		if err := sc.Err(); err != nil {
			panic(err)
		}
		// seal the result (prevent modification from outside this class)
		result := make([]map[string]string, len(list))
		for i, m := range list {
			cp := make(map[string]string, len(m))
			for k, v := range m {
				cp[k] = v
			}
			result[i] = cp
		}
		topoWrongWords = result
	})
	return topoWrongWords
}

// TopoReplaceRule ports org.languagetool.rules.br.TopoReplaceRule.
type TopoReplaceRule struct {
	Messages map[string]string
	Category *rules.Category
}

// NewTopoReplaceRule ports new TopoReplaceRule(messages).
func NewTopoReplaceRule(messages map[string]string) *TopoReplaceRule {
	_ = loadTopoWords()
	return &TopoReplaceRule{
		Messages: messages,
		// Java: super.setCategory(Categories.MISC.getCategory(messages));
		Category: rules.CatMisc.GetCategory(messages),
	}
}

// GetID ports getId() → BRETON_TOPO.
func (r *TopoReplaceRule) GetID() string { return BretonTopo }

// GetDescription ports getDescription().
func (r *TopoReplaceRule) GetDescription() string {
	return "anvioù-lec’h e brezhoneg"
}

// GetShort ports getShort().
func (r *TopoReplaceRule) GetShort() string {
	return "anvioù lec’h"
}

// GetSuggestion ports getSuggestion().
func (r *TopoReplaceRule) GetSuggestion() string {
	return " zo un anv lec’h gallek. Ha fellout a rae deoc’h skrivañ "
}

// GetSuggestionsSeparator ports getSuggestionsSeparator() → " pe ".
func (r *TopoReplaceRule) GetSuggestionsSeparator() string {
	return " pe "
}

// IsCaseSensitive ports isCaseSensitive() → true.
func (r *TopoReplaceRule) IsCaseSensitive() bool {
	return true
}

// GetLocale ports getLocale() → new Locale("br").
func (r *TopoReplaceRule) GetLocale() string {
	return topoLocaleBR
}

// GetCategory ports Rule.getCategory() after constructor MISC category.
func (r *TopoReplaceRule) GetCategory() *rules.Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// GetWrongWords ports getWrongWords().
func (r *TopoReplaceRule) GetWrongWords() []map[string]string {
	return loadTopoWords()
}

// addToTopoQueue ports addToQueue (ArrayBlockingQueue capacity = wrongWords.size()).
func addToTopoQueue(token *languagetool.AnalyzedTokenReadings, prev []*languagetool.AnalyzedTokenReadings, cap int) []*languagetool.AnalyzedTokenReadings {
	// Java: boolean inserted = prevTokens.offer(token);
	// if (!inserted) { prevTokens.poll(); prevTokens.offer(token); }
	if cap <= 0 {
		return prev
	}
	if len(prev) >= cap {
		prev = prev[1:]
	}
	return append(prev, token)
}

// Match ports TopoReplaceRule.match.
func (r *TopoReplaceRule) Match(sentence *languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	wrongWords := loadTopoWords()
	if len(wrongWords) == 0 {
		// Java new ArrayBlockingQueue<>(0) is illegal; official data is non-empty.
		return nil
	}
	tokens := sentence.GetTokensWithoutWhitespace()
	capQ := len(wrongWords)
	prev := make([]*languagetool.AnalyzedTokenReadings, 0, capQ)
	var ruleMatches []*rules.RuleMatch

	for i := 1; i < len(tokens); i++ {
		prev = addToTopoQueue(tokens[i], prev, capQ)

		// Java: build variants by walking queue reverse, add(0, sb) → longest first.
		prevList := prev
		variants := make([]string, 0, len(prevList))
		for j := len(prevList) - 1; j >= 0; j-- {
			var b strings.Builder
			for k := j; k < len(prevList); k++ {
				// space when token k has whitespace before (Java checks get(j+1).isWhitespaceBefore while prepending)
				if k > j && prevList[k] != nil && prevList[k].IsWhitespaceBefore() {
					b.WriteByte(' ')
				}
				if prevList[k] != nil {
					b.WriteString(prevList[k].GetToken())
				}
			}
			// variants.add(0, …)
			variants = append([]string{b.String()}, variants...)
		}

		lenV := len(variants) // prevTokensList and variants have now the same length
		for j := 0; j < lenV; j++ { // longest words first
			crtWordCount := lenV - j
			startIdx := lenV - crtWordCount // Java: prevTokensList.get(len - crtWordCount)
			if prevList[startIdx] != nil && prevList[startIdx].IsImmunized() {
				continue
			}
			crt := variants[j]
			var crtMatch string
			var ok bool
			if r.IsCaseSensitive() {
				crtMatch, ok = wrongWords[crtWordCount-1][crt]
			} else {
				// Java: crt.toLowerCase(getLocale()) — Locale("br") uses standard lowercasing
				crtMatch, ok = wrongWords[crtWordCount-1][strings.ToLower(crt)]
			}
			if !ok {
				continue
			}
			// Java: Arrays.asList(crtMatch.split("\\|")) — no trim, keep empties
			replacements := strings.Split(crtMatch, "|")
			msg := crt + r.GetSuggestion()
			for k := 0; k < len(replacements); k++ {
				if k > 0 {
					if k == len(replacements)-1 {
						msg += r.GetSuggestionsSeparator()
					} else {
						msg += ", "
					}
				}
				msg += "<suggestion>" + replacements[k] + "</suggestion>"
			}
			msg += "?"
			startPos := prevList[startIdx].GetStartPos()
			endPos := prevList[lenV-1].GetEndPos()
			// Java: new RuleMatch(this, sentence, startPos, endPos, msg, getShort())
			rm := rules.NewRuleMatch(r, sentence, startPos, endPos, msg)
			rm.SetShortMessage(r.GetShort())

			// Java: if (!isCaseSensitive() && StringTools.startsWithUppercase(crt)) …
			if !r.IsCaseSensitive() && tools.StartsWithUppercase(crt) {
				for k := range replacements {
					replacements[k] = tools.UppercaseFirstChar(replacements[k])
				}
			}
			rm.SetSuggestedReplacements(replacements)
			ruleMatches = append(ruleMatches, rm)
			break
		}
	}
	return ruleMatches
}
