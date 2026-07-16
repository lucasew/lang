package dumpcheck

import "strings"

// Sentence ports org.languagetool.dev.dumpcheck.Sentence.
type Sentence struct {
	Text         string
	Source       string
	Title        string
	URL          string
	ArticleCount int
}

func NewSentence(text, source, title, url string, articleCount int) Sentence {
	return Sentence{
		Text:         strings.TrimSpace(text),
		Source:       source,
		Title:        title,
		URL:          url,
		ArticleCount: articleCount,
	}
}

func (s Sentence) GetText() string   { return s.Text }
func (s Sentence) GetSource() string { return s.Source }
func (s Sentence) GetTitle() string  { return s.Title }
func (s Sentence) String() string    { return s.Text }
