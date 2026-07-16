package dumpcheck

import (
	"fmt"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// CSVHandler ports org.languagetool.dev.dumpcheck.CSVHandler (tab-separated to W).
type CSVHandler struct {
	*ResultHandler
	W io.Writer
}

func NewCSVHandler(w io.Writer, maxSentences, maxErrors int) *CSVHandler {
	h := &CSVHandler{
		ResultHandler: NewResultHandler(maxSentences, maxErrors),
		W:             w,
	}
	h.Handle = h.printCSV
	return h
}

func (h *CSVHandler) printCSV(sentence Sentence, matches []*rules.RuleMatch, langCode string) error {
	sentenceStr := sentence.GetText()
	if len(matches) == 0 {
		fmt.Fprintf(h.W, "NOMATCH\t\t%s\n", noTabs(sentenceStr))
		return nil
	}
	for _, match := range matches {
		if match == nil {
			continue
		}
		from, to := match.FromPos, match.ToPos
		runes := []rune(sentenceStr)
		if from < 0 {
			from = 0
		}
		if to > len(runes) {
			to = len(runes)
		}
		if from > to {
			from = to
		}
		marked := noTabs(string(runes[:from])) + "__" +
			noTabs(string(runes[from:to])) + "__" +
			noTabs(string(runes[to:]))
		fmt.Fprintf(h.W, "MATCH\t%s\t%s\n", ruleIDOf(match), marked)
	}
	return nil
}

func noTabs(s string) string {
	return strings.ReplaceAll(s, "\t", `\t`)
}
