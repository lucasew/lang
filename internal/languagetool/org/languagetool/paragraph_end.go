package languagetool

import "strings"

// IsParagraphEnd ports org.languagetool.tools.Tools.isParagraphEnd.
// singleLineBreaksMarksPara is Language.getSentenceTokenizer().singleLineBreaksMarksPara().
// Lives in package languagetool (not tools) to avoid import cycles with tagging/tools.
func IsParagraphEnd(sentences []*AnalyzedSentence, nTest int, singleLineBreaksMarksPara bool) bool {
	if nTest >= len(sentences)-1 {
		return true
	}
	text := sentences[nTest].GetText()
	if singleLineBreaksMarksPara {
		if strings.HasSuffix(text, "\n") || strings.HasSuffix(text, "\n\r") {
			return true
		}
	} else {
		if strings.HasSuffix(text, "\n\n") || strings.HasSuffix(text, "\n\r\n\r") || strings.HasSuffix(text, "\r\n\r\n") {
			return true
		}
	}
	next := sentences[nTest+1].GetText()
	if strings.HasPrefix(next, "\n") || strings.HasPrefix(next, "\r\n") {
		return true
	}
	return false
}
