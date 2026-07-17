package en

import (
	"bufio"
	"os"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
	entag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/en"
)

// softEnglishMultiwords is a small tab-separated multiword list for the live check path.
// Full multiwords.txt uses glued tags for some lines; keep this soft and safe.
var softEnglishMultiwords = []string{
	"New York\tNNP",
	"Los Angeles\tNNP",
	"United States\tNNP",
	"United Kingdom\tNNP",
	"San Francisco\tNNP",
	"Hong Kong\tNNP",
	"New Zealand\tNNP",
	"South Africa\tNNP",
	"Costa Rica\tNNP",
	"Silicon Valley\tNNP",
	"Wall Street\tNNP",
	"Middle East\tNNP",
	"status quo\tNN",
	"Status Quo\tNN",
	"as well\tRB",
	"for example\tRB",
	"in fact\tRB",
	"of course\tRB",
	"at least\tRB",
	"by the way\tRB",
	"in general\tRB",
	"as soon as\tRB",
	"in addition\tRB",
	"Taj Mahal\tNNP",
	"Yom Kippur\tNNP",
}

// RegisterSoftEnglishDisambiguator installs multiword chunking, optional soft XML
// rules, and a data-driven ignore-spelling word list on lt.Disambiguator.
// ignoreSpellingPath is a plain-text list (one surface form per line); empty skips.
func RegisterSoftEnglishDisambiguator(lt *languagetool.JLanguageTool, multiwordsPath, softDisambigXMLPath, ignoreSpellingPath string) {
	if lt == nil {
		return
	}
	lines := append([]string(nil), softEnglishMultiwords...)
	if multiwordsPath != "" {
		if f, err := os.Open(multiwordsPath); err == nil {
			// Only append tab-separated lines to avoid panics on glued-tag format.
			if loaded, err := loadTabSeparatedMultiwords(f); err == nil && len(loaded) > 0 {
				lines = append(lines, loaded...)
			}
			_ = f.Close()
		}
	}
	chunker := disambiguation.NewMultiWordChunker(lines, disambiguation.MultiWordChunkerSettings{
		AllowFirstCapitalized: true,
		AllowAllUppercase:     true,
		AllowTitlecase:        true,
	})
	chunker.SetIgnoreSpelling(true)
	hyb := entag.NewEnglishHybridDisambiguator()
	hyb.Chunker = chunker
	if softDisambigXMLPath != "" {
		if data, err := os.ReadFile(softDisambigXMLPath); err == nil {
			if rules, err := disambigrules.NewDisambiguationRuleLoader().GetRulesFromString(string(data), "en", softDisambigXMLPath); err == nil && len(rules) > 0 {
				hyb.RulesDisambiguator = disambigrules.NewXmlRuleDisambiguator(rules)
			}
		}
	}
	var steps []languagetool.SentenceDisambiguator
	steps = append(steps, hyb)
	if ignoreSpellingPath != "" {
		if words, err := loadIgnoreSpellingWords(ignoreSpellingPath); err == nil && len(words) > 0 {
			steps = append(steps, &ignoreSpellingWordList{words: words})
		}
	}
	if len(steps) == 1 {
		lt.Disambiguator = steps[0]
		return
	}
	lt.Disambiguator = chainSentenceDisambiguator(steps)
}

// chainSentenceDisambiguator applies disambiguators in order.
type chainSentenceDisambiguator []languagetool.SentenceDisambiguator

func (c chainSentenceDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	s := input
	for _, step := range c {
		if step == nil || s == nil {
			continue
		}
		if out := step.Disambiguate(s); out != nil {
			s = out
		}
	}
	return s
}

// ignoreSpellingWordList marks listed surface forms with IgnoreSpelling (case-sensitive + lower).
type ignoreSpellingWordList struct {
	words map[string]struct{}
}

func (w *ignoreSpellingWordList) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil || w == nil || len(w.words) == 0 {
		return input
	}
	for _, tok := range input.GetTokensWithoutWhitespace() {
		if tok == nil {
			continue
		}
		surface := tok.GetToken()
		if _, ok := w.words[surface]; ok {
			tok.IgnoreSpelling()
			continue
		}
		if low := strings.ToLower(surface); low != surface {
			if _, ok := w.words[low]; ok {
				tok.IgnoreSpelling()
			}
		}
	}
	return input
}

func loadIgnoreSpellingWords(path string) (map[string]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]struct{}{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		if line == "" {
			continue
		}
		out[line] = struct{}{}
		out[strings.ToLower(line)] = struct{}{}
	}
	return out, sc.Err()
}

func loadTabSeparatedMultiwords(f *os.File) ([]string, error) {
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		// require a tab so glued "phraseTAG" lines from upstream multiwords.txt are skipped
		if !strings.Contains(line, "\t") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			continue
		}
		lines = append(lines, parts[0]+"\t"+parts[1])
	}
	return lines, sc.Err()
}
