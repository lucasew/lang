package en

import (
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	enTok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// ConsistentApostrophesRule ports org.languagetool.rules.en.ConsistentApostrophesRule.
type ConsistentApostrophesRule struct {
	Messages map[string]string
}

func NewConsistentApostrophesRule(messages map[string]string) *ConsistentApostrophesRule {
	return &ConsistentApostrophesRule{Messages: messages}
}

func (r *ConsistentApostrophesRule) GetID() string { return "EN_CONSISTENT_APOS" }

// MatchList ports match(List<AnalyzedSentence>).
func (r *ConsistentApostrophesRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*rules.RuleMatch {
	if !hasTwoApostropheTypes(sentences) {
		return nil
	}
	var matches []*rules.RuleMatch
	pos := 0
	for _, sentence := range sentences {
		for _, token := range sentence.GetTokens() {
			if token == nil {
				continue
			}
			t := token.GetToken()
			var message, repl string
			if strings.Contains(t, "'") && !strings.Contains(t, "’") {
				message = "You used a typewriter-style apostrophe here, but a typographic apostrophe elsewhere in this text."
				repl = strings.ReplaceAll(t, "'", "’")
			} else if strings.Contains(t, "’") {
				message = "You used a typographic apostrophe here, but a typewriter-style apostrophe elsewhere in this text."
				repl = strings.ReplaceAll(t, "’", "'")
			}
			if message != "" {
				msg := message + " Both are correct, but consider using the same type everywhere in your text."
				rm := rules.NewRuleMatch(r, sentence, pos+token.GetStartPos(), pos+token.GetEndPos(), msg)
				rm.SetSuggestedReplacement(repl)
				matches = append(matches, rm)
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return matches
}

func hasTwoApostropheTypes(sentences []*languagetool.AnalyzedSentence) bool {
	hasTypewriter, hasTypographic := false, false
	for _, sentence := range sentences {
		for _, token := range sentence.GetTokens() {
			if token == nil {
				continue
			}
			t := token.GetToken()
			if strings.Contains(t, "'") && !strings.Contains(t, "’") {
				hasTypewriter = true
			}
			if strings.Contains(t, "’") {
				hasTypographic = true
			}
			if hasTypewriter && hasTypographic {
				return true
			}
		}
	}
	return false
}

// AnalyzeEnglishPlain analyzes text with EnglishWordTokenizer (contraction splits).
func AnalyzeEnglishPlain(text string) *languagetool.AnalyzedSentence {
	wt := enTok.NewEnglishWordTokenizer()
	raw := wt.Tokenize(text)
	positions := tokenizers.BuildPositions(raw)
	readings := make([]*languagetool.AnalyzedTokenReadings, 0, len(raw)+1)
	ss := languagetool.SentenceStartTagName
	startTok := languagetool.NewAnalyzedToken("", &ss, nil)
	startR := languagetool.NewAnalyzedTokenReadings(startTok)
	startR.SetStartPos(0)
	readings = append(readings, startR)
	prev := ""
	for i, tok := range raw {
		at := languagetool.NewAnalyzedToken(tok, nil, nil)
		ar := languagetool.NewAnalyzedTokenReadingsAt(at, positions[i])
		if prev != "" {
			ar.SetWhitespaceBeforeToken(prev)
		}
		readings = append(readings, ar)
		prev = tok
	}
	return languagetool.NewAnalyzedSentence(readings)
}

// AnalyzeEnglishText splits into sentences and analyzes each with EnglishWordTokenizer.
func AnalyzeEnglishText(text string) []*languagetool.AnalyzedSentence {
	// Split like SplitAndAnalyze but use English tokenizer per sentence.
	parts := splitSentencesEN(text)
	if len(parts) == 0 {
		return []*languagetool.AnalyzedSentence{AnalyzeEnglishPlain(text)}
	}
	out := make([]*languagetool.AnalyzedSentence, 0, len(parts))
	offset := 0
	for _, p := range parts {
		if p == "" {
			continue
		}
		s := AnalyzeEnglishPlain(p)
		if offset > 0 {
			for _, t := range s.GetTokens() {
				t.SetStartPos(t.GetStartPos() + offset)
			}
		}
		out = append(out, s)
		for _, r := range p {
			offset += len(utf16.Encode([]rune{r}))
		}
	}
	return out
}

func splitSentencesEN(text string) []string {
	// Reuse languagetool.SplitAndAnalyze structure via simple split
	sents := languagetool.SplitAndAnalyze(text)
	if len(sents) == 0 {
		return []string{text}
	}
	var parts []string
	for _, s := range sents {
		parts = append(parts, s.GetText())
	}
	return parts
}
