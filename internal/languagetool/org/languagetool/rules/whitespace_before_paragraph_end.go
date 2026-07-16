package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// WhiteSpaceBeforeParagraphEnd ports org.languagetool.rules.WhiteSpaceBeforeParagraphEnd.
type WhiteSpaceBeforeParagraphEnd struct {
	Messages                  map[string]string
	SingleLineBreaksMarksPara bool
}

func NewWhiteSpaceBeforeParagraphEnd(messages map[string]string) *WhiteSpaceBeforeParagraphEnd {
	return &WhiteSpaceBeforeParagraphEnd{Messages: messages}
}

func (r *WhiteSpaceBeforeParagraphEnd) GetID() string { return "WHITESPACE_PARAGRAPH" }

func (r *WhiteSpaceBeforeParagraphEnd) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	if nTest >= len(sentences)-1 {
		return true
	}
	text := sentences[nTest].GetText()
	if r.SingleLineBreaksMarksPara {
		if len(text) > 0 && text[len(text)-1] == '\n' {
			return true
		}
	} else if len(text) >= 2 && text[len(text)-2:] == "\n\n" {
		return true
	}
	next := sentences[nTest+1].GetText()
	if len(next) > 0 && next[0] == '\n' {
		return true
	}
	return false
}

// MatchList ports match(List<AnalyzedSentence>).
func (r *WhiteSpaceBeforeParagraphEnd) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	pos := 0
	msg := r.Messages["whitespace_before_parapgraph_end_msg"]
	if msg == "" {
		msg = "Don't end a paragraph with whitespace"
	}
	for n := 0; n < len(sentences); n++ {
		sentence := sentences[n]
		if r.isParagraphEnd(sentences, n) {
			tokens := sentence.GetTokens()
			lb := len(tokens) - 1
			for lb > 0 && tokens[lb].IsLinebreak() {
				lb--
			}
			lw := lb
			for lw > 0 && tokens[lw].IsWhitespace() && tokens[lw].GetToken() != "\u200B" {
				lw--
			}
			if lw < lb {
				fromPos := pos + tokens[lw].GetStartPos()
				if tokens[lw].IsWhitespace() && lw+1 < len(tokens) {
					fromPos = pos + tokens[lw+1].GetStartPos()
				}
				toPos := pos + tokens[lb].GetEndPos()
				rm := NewRuleMatch(r, sentence, fromPos, toPos, msg)
				if lw > 0 && !tokens[lw].IsWhitespace() {
					rm.SetSuggestedReplacement(tokens[lw].GetToken())
				} else {
					rm.SetSuggestedReplacement("")
				}
				ruleMatches = append(ruleMatches, rm)
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
