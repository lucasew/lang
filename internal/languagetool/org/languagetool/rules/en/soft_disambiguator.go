package en

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
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
	"Bay Area\tNNP",
	"East Coast\tNNP",
	"West Coast\tNNP",
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
	// Java EnglishHybridDisambiguator: chunker.setRemovePreviousTags(true)
	// so <NNP></NNP> becomes NNP NNP and original POS is replaced.
	chunker.SetRemovePreviousTags(true)
	hyb := entag.NewEnglishHybridDisambiguator()
	// Java EnglishHybridDisambiguator: chunkerGlobal = MultiWordChunker(/spelling_global.txt,
	// tagForNotAddingTags) with setIgnoreSpelling(true), run before multiwords.
	if globalLines := loadSpellingGlobalMultiwordLines(); len(globalLines) > 0 {
		global := disambiguation.NewMultiWordChunker(globalLines, disambiguation.MultiWordChunkerSettings{
			DefaultTag:            disambiguation.TagForNotAddingTags,
			AllowFirstCapitalized: true,
			AllowAllUppercase:     true,
			AllowTitlecase:        true,
		})
		global.SetIgnoreSpelling(true)
		hyb.GlobalChunker = global
	}
	hyb.Chunker = chunker
	// Load upstream soft disambig first, then legacy en-soft.xml (hand soft immunize
	// abbreviations). Prefer both so ignore_spelling vs immunize stay distinct.
	var allDisambigRules []*disambigrules.DisambiguationPatternRule
	var sharedUnifier *patterns.UnifierConfiguration
	loader := disambigrules.NewDisambiguationRuleLoader()
	for _, p := range softDisambiguationXMLPaths(softDisambigXMLPath) {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		rules, cfg, err := loader.GetRulesAndUnifierFromString(string(data), "en", p)
		if err != nil || len(rules) == 0 {
			continue
		}
		if sharedUnifier == nil {
			sharedUnifier = cfg
		} else if cfg != nil {
			for loc, pt := range cfg.GetEquivalenceTypes() {
				sharedUnifier.SetEquivalence(loc.Feature, loc.Type, pt)
			}
		}
		// Hand en-soft.xml immunizes ordinary words (kind, …). Drop IMMUNIZE when
		// that pack is loaded explicitly; keep FILTER/REPLACE for modal soft tests.
		base := filepath.Base(p)
		if base == "en-soft.xml" || base == "en-soft-disambiguation.xml" {
			filtered := rules[:0]
			for _, r := range rules {
				if r == nil || r.Action == disambigrules.ActionImmunize {
					continue
				}
				filtered = append(filtered, r)
			}
			rules = filtered
		}
		for _, r := range rules {
			if r != nil {
				r.UnifierConfig = sharedUnifier
			}
		}
		allDisambigRules = append(allDisambigRules, rules...)
	}
	if len(allDisambigRules) > 0 {
		x := disambigrules.NewXmlRuleDisambiguator(allDisambigRules)
		x.UnifierConfig = sharedUnifier
		hyb.RulesDisambiguator = x
	}
	var steps []languagetool.SentenceDisambiguator
	steps = append(steps, hyb)
	words := map[string]struct{}{}
	// Prefer soft tech list when provided, then merge official spelling.txt extensions.
	for _, p := range ignoreSpellingPaths(ignoreSpellingPath) {
		if loaded, err := loadIgnoreSpellingWords(p); err == nil {
			for k := range loaded {
				words[k] = struct{}{}
			}
		}
	}
	if len(words) > 0 {
		steps = append(steps, &ignoreSpellingWordList{words: words})
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

// softDisambiguationXMLPaths returns soft/upstream disambiguation XML paths.
// Prefer official upstream soft extract. Do NOT auto-merge hand en-soft.xml:
// that pack immunizes thousands of ordinary words (e.g. "kind"), which blocks
// official pattern rules. Hand en-soft is only used when it is the primary path
// or when no upstream soft file is available.
func softDisambiguationXMLPaths(primary string) []string {
	var out []string
	seen := map[string]struct{}{}
	add := func(p string) {
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		if st, err := os.Stat(p); err != nil || !st.Mode().IsRegular() {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	add(primary)
	// walk-up fallbacks if primary empty — official upstream soft only.
	// Hand en-soft.xml is opt-in via primary path (tests); it must not merge
	// with upstream by default (immunize/filter interactions clear tags).
	if len(out) == 0 {
		wd, err := os.Getwd()
		if err == nil {
			dir := wd
			for {
				add(filepath.Join(dir, "testdata", "disambiguation", "en-disambiguation-upstream-soft.xml"))
				parent := filepath.Dir(dir)
				if parent == dir {
					break
				}
				dir = parent
			}
		}
	}
	return out
}

// ignoreSpellingPaths mirrors Java SpellingCheckRule.init(): soft list (if set), then
// hunspell/ignore.txt, spelling.txt, and spelling_global.txt (getAdditionalSpellingFileNames).
func ignoreSpellingPaths(primary string) []string {
	var out []string
	seen := map[string]struct{}{}
	add := func(p string) {
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	add(primary)
	// walk-up from cwd for vendored official EN hunspell + global spelling lists
	wd, err := os.Getwd()
	if err != nil {
		return out
	}
	dir := wd
	for {
		for _, rel := range []string{
			// SpellingCheckRule.getIgnoreFileName() → en/hunspell/ignore.txt
			filepath.Join("testdata", "upstream", "en", "resource", "hunspell", "ignore.txt"),
			filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
				"src", "main", "resources", "org", "languagetool", "resource", "en", "hunspell", "ignore.txt"),
			// SpellingCheckRule.getSpellingFileName() → en/hunspell/spelling.txt
			filepath.Join("testdata", "disambiguation", "en-spelling-upstream.txt"),
			filepath.Join("testdata", "upstream", "en", "resource", "hunspell", "spelling.txt"),
			// MorfologikAmericanSpellerRule.getLanguageVariantSpellingFileName()
			// → en/hunspell/spelling_en-US.txt (default soft EN is en-US)
			filepath.Join("testdata", "upstream", "en", "resource", "hunspell", "spelling_en-US.txt"),
			filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
				"src", "main", "resources", "org", "languagetool", "resource", "en", "hunspell", "spelling_en-US.txt"),
			// SpellingCheckRule.CUSTOM_SPELLING_FILE
			filepath.Join("testdata", "upstream", "en", "resource", "hunspell", "spelling_custom.txt"),
			// SpellingCheckRule.GLOBAL_SPELLING_FILE → spelling_global.txt
			filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
				"org", "languagetool", "resource", "spelling_global.txt"),
			filepath.Join("testdata", "upstream", "spelling_global.txt"),
		} {
			cand := filepath.Join(dir, rel)
			if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
				add(cand)
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return out
}

// discoverSpellingGlobalPath finds official spelling_global.txt (Java /spelling_global.txt).
func discoverSpellingGlobalPath() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		for _, rel := range []string{
			filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources",
				"org", "languagetool", "resource", "spelling_global.txt"),
			filepath.Join("testdata", "upstream", "spelling_global.txt"),
		} {
			cand := filepath.Join(dir, rel)
			if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
				return cand
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// Cached multi-token phrases from spelling_global.txt (Java chunkerGlobal is a singleton).
var (
	spellingGlobalMWOnce  sync.Once
	spellingGlobalMWLines []string
)

// loadSpellingGlobalMultiwordLines loads multi-token phrases from spelling_global.txt
// for MultiWordChunker (Java chunkerGlobal; tagForNotAddingTags + ignoreSpelling).
// Single-token lines are handled via ignoreSpellingWordList (wordsToBeIgnored).
func loadSpellingGlobalMultiwordLines() []string {
	spellingGlobalMWOnce.Do(func() {
		p := discoverSpellingGlobalPath()
		if p == "" {
			return
		}
		f, err := os.Open(p)
		if err != nil {
			return
		}
		defer f.Close()
		var lines []string
		sc := bufio.NewScanner(f)
		buf := make([]byte, 0, 64*1024)
		sc.Buffer(buf, 1024*1024)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if i := strings.IndexByte(line, '#'); i >= 0 {
				line = strings.TrimSpace(line[:i])
			}
			if line == "" || !strings.Contains(line, " ") {
				continue
			}
			// phrase-only; DefaultTag supplies TagForNotAddingTags
			lines = append(lines, line)
		}
		spellingGlobalMWLines = lines
	})
	return spellingGlobalMWLines
}

func loadIgnoreSpellingWords(path string) (map[string]struct{}, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	out := map[string]struct{}{}
	sc := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
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
		// Java addIgnoreWords: multi-token lines become IGNORE_SPELLING antipatterns
		// (handled by MultiWordChunker for spelling_global); only single tokens go
		// into the wordsToBeIgnored set used by ignoreWord().
		if strings.Contains(line, " ") {
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
	// Upstream multiwords files can include long proper-name lines.
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = strings.TrimSpace(line[:i])
		}
		phrase, tag, ok := splitMultiwordLine(line)
		if !ok {
			continue
		}
		lines = append(lines, phrase+"\t"+tag)
	}
	return lines, sc.Err()
}

// splitMultiwordLine accepts tab-separated "phrase\ttag" or LT glued "phraseTAG"
// (POS stuck to the last character of the phrase, e.g. "status quoNN:UN").
func splitMultiwordLine(line string) (phrase, tag string, ok bool) {
	if i := strings.IndexByte(line, '\t'); i >= 0 {
		phrase = strings.TrimSpace(line[:i])
		tag = strings.TrimSpace(line[i+1:])
		return phrase, tag, phrase != "" && tag != ""
	}
	// Glued form: trailing uppercase POS token (NN, NNP, NN:UN, NNS, RB, …).
	if len(line) < 3 {
		return "", "", false
	}
	end := len(line)
	j := end - 1
	for j >= 0 {
		c := line[j]
		if (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == ':' || c == '+' || c == '_' || c == '-' {
			j--
			continue
		}
		break
	}
	tagStart := j + 1
	if tagStart <= 0 || tagStart >= end {
		return "", "", false
	}
	tag = line[tagStart:]
	if len(tag) < 2 || tag[0] < 'A' || tag[0] > 'Z' {
		return "", "", false
	}
	phrase = strings.TrimSpace(line[:tagStart])
	// multiword requires a space in the phrase
	if phrase == "" || !strings.Contains(phrase, " ") {
		return "", "", false
	}
	return phrase, tag, true
}
