package dumpcheck

import (
	"time"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

const (
	maxContextLength   = 500
	smallContextLength = 40
)

// CorpusMatch is one row that would be inserted into corpus_match.
type CorpusMatch struct {
	LanguageCode      string
	RuleID            string
	RuleCategory      string
	RuleSubID         string
	RuleDescription   string
	Message           string
	ErrorContext      string
	SmallErrorContext string
	SourceURI         string
	SourceType        string
	CheckDate         time.Time
}

// DatabaseHandler ports org.languagetool.dev.dumpcheck.DatabaseHandler without SQL:
// collects CorpusMatch rows in memory (real JDBC deferred).
type DatabaseHandler struct {
	*ResultHandler
	Matches []CorpusMatch
	// CategoryOf optional: map rule → category name
	CategoryOf func(rule any) string
	// DescriptionOf optional
	DescriptionOf func(rule any) string
	// SubIDOf optional
	SubIDOf func(rule any) string
}

func NewDatabaseHandler(maxSentences, maxErrors int) *DatabaseHandler {
	h := &DatabaseHandler{ResultHandler: NewResultHandler(maxSentences, maxErrors)}
	h.Handle = h.store
	return h
}

func (h *DatabaseHandler) store(sentence Sentence, matches []*rules.RuleMatch, langCode string) error {
	now := time.Now()
	for _, match := range matches {
		if match == nil {
			continue
		}
		ctx := plainTextContext(match.FromPos, match.ToPos, sentence.GetText(), maxContextLength)
		if utf8.RuneCountInString(ctx) > maxContextLength {
			// Java skips oversized contexts
			continue
		}
		small := plainTextContext(match.FromPos, match.ToPos, sentence.GetText(), smallContextLength)
		cat, desc, sub := "", "", ""
		if h.CategoryOf != nil {
			cat = h.CategoryOf(match.GetRule())
		}
		if h.DescriptionOf != nil {
			desc = h.DescriptionOf(match.GetRule())
		}
		if h.SubIDOf != nil {
			sub = h.SubIDOf(match.GetRule())
		}
		msg := match.GetMessage()
		if len([]rune(msg)) > 255 {
			msg = string([]rune(msg)[:255])
		}
		if utf8.RuneCountInString(small) > 255 {
			r := []rune(small)
			small = string(r[:255])
		}
		h.Matches = append(h.Matches, CorpusMatch{
			LanguageCode:      langCode,
			RuleID:            ruleIDOf(match),
			RuleCategory:      cat,
			RuleSubID:         sub,
			RuleDescription:   desc,
			Message:           msg,
			ErrorContext:      ctx,
			SmallErrorContext: small,
			SourceURI:         sentence.URL,
			SourceType:        sentence.GetSource(),
			CheckDate:         now,
		})
	}
	return nil
}

// Close is a no-op for the in-memory handler (SQL flush deferred).
func (h *DatabaseHandler) Close() error { return nil }
