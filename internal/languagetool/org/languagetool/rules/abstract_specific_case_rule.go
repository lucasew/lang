package rules

import (
	"bufio"
	"io"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AbstractSpecificCaseRule ports org.languagetool.rules.AbstractSpecificCaseRule.
// Java ctor: Categories.CASING, ITSIssueType.Misspelling.
type AbstractSpecificCaseRule struct {
	Messages                   map[string]string
	LcToProper                 map[string]string
	MaxPhraseLen               int
	ID                         string
	Description                string
	InitialCapitalMessage      string
	OtherCapitalizationMessage string
	ShortMsg                   string
	// Category ports Rule.category (Java CASING).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Misspelling).
	IssueType ITSIssueType
	// URL ports Rule.url (Java setUrl).
	URL string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
}

// InitSpecificCaseMeta applies Java AbstractSpecificCaseRule constructor metadata.
func InitSpecificCaseMeta(r *AbstractSpecificCaseRule, messages map[string]string) {
	if r == nil {
		return
	}
	r.Messages = messages
	if r.Category == nil {
		r.Category = CatCasing.GetCategory(messages)
	}
	if r.IssueType == "" {
		r.IssueType = ITSMisspelling
	}
}

// LoadSpecificCasePhrases loads phrase list: each non-comment line is the proper spelling.
func LoadSpecificCasePhrases(r io.Reader) (map[string]string, int, error) {
	m := make(map[string]string)
	maxLen := 0
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.Split(line, " ")
		if len(parts) > maxLen {
			maxLen = len(parts)
		}
		m[strings.ToLower(line)] = line
	}
	return m, maxLen, sc.Err()
}

func (r *AbstractSpecificCaseRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "SPECIFIC_CASE"
}

// GetDescription ports Rule.getDescription.
func (r *AbstractSpecificCaseRule) GetDescription() string {
	if r == nil {
		return ""
	}
	return r.Description
}

// GetCategory ports Rule.getCategory.
func (r *AbstractSpecificCaseRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType.
func (r *AbstractSpecificCaseRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSMisspelling
	}
	return r.IssueType
}

// GetURL ports Rule.getUrl.
func (r *AbstractSpecificCaseRule) GetURL() string {
	if r == nil {
		return ""
	}
	return r.URL
}

// SetURL ports Rule.setUrl.
func (r *AbstractSpecificCaseRule) SetURL(u string) {
	if r != nil {
		r.URL = u
	}
}

// AddExamplePair ports Rule.addExamplePair.
func (r *AbstractSpecificCaseRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *AbstractSpecificCaseRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *AbstractSpecificCaseRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

func (r *AbstractSpecificCaseRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var matches []*RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	for i := 0; i < len(tokens); i++ {
		var l []string
		j := 0
		for len(l) < r.MaxPhraseLen && i+j < len(tokens) {
			l = append(l, tokens[i+j].GetToken())
			j++
			phrase := strings.Join(l, " ")
			lcPhrase := strings.ToLower(phrase)
			properSpelling, ok := r.LcToProper[lcPhrase]
			if !ok || tools.IsAllUppercase(phrase) || phrase == properSpelling {
				continue
			}
			// sentence start: avoid suggesting lowercase-start phrases
			if i > 0 && tokens[i-1].IsSentenceStart() && !tools.StartsWithUppercase(properSpelling) {
				continue
			}
			// Also: first real token after SENT_START is i==1 with tokens[0] SENT_START
			if i == 1 && tokens[0].IsSentenceStart() && !tools.StartsWithUppercase(properSpelling) {
				continue
			}
			msg := r.OtherCapitalizationMessage
			if msg == "" {
				msg = "The particular expression should follow the suggested capitalization."
			}
			if allWordsUppercase(properSpelling) {
				msg = r.InitialCapitalMessage
				if msg == "" {
					msg = "The initials of the particular phrase must be capitals."
				}
			}
			from := tokens[i].GetStartPos()
			to := tokens[i+j-1].GetEndPos()
			rm := NewRuleMatch(r, sentence, from, to, msg)
			short := r.ShortMsg
			if short == "" {
				short = "Special capitalization"
			}
			rm.ShortMessage = short
			rm.SetSuggestedReplacement(properSpelling)
			matches = append(matches, rm)
		}
	}
	return matches
}

func allWordsUppercase(s string) bool {
	for _, w := range strings.Split(s, " ") {
		if !tools.StartsWithUppercase(w) {
			return false
		}
	}
	return true
}
