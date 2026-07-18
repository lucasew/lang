package commandline

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	disambigrules "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/rules"
)

// SoftHybridPaths supplies official multiword / soft disambig XML / optional DE multitoken paths.
// Empty strings mean "skip that resource" (auto-discovery is caller's job).
type SoftHybridPaths struct {
	Multiwords          string // /{lang}/multiwords.txt
	SoftDisambigXML     string // soft extract of disambiguation.xml
	DEMultitokenIgnore  string // German multitoken-ignore.txt
	DEMultitokenSuggest string // German multitoken-suggest.txt
	SpellingGlobal      string // Java /spelling_global.txt; empty → walk-up discovery
}

// hybridLangProfile captures Java HybridDisambiguator / GermanRuleDisambiguator flags.
type hybridLangProfile struct {
	mwAllowFirst, mwAllowAllUpper, mwAllowTitle bool
	mwDefaultTag                                string
	mwRemovePrev, mwIgnoreSpell                 bool
	useGlobal                                   bool
	gAllowFirst, gAllowAllUpper, gAllowTitle    bool
	gDefaultTag                                 string
	gIgnoreSpell                                bool
	// "global_mw_xml" | "mw_xml" | "xml_mw" | "de"
	order string
}

func softHybridProfile(lang string) hybridLangProfile {
	switch strings.ToLower(lang) {
	case "fr":
		// FrenchHybridDisambiguator: global → multiwords → XML
		return hybridLangProfile{
			mwAllowFirst: true, mwAllowAllUpper: true, mwAllowTitle: false,
			mwRemovePrev: true,
			useGlobal:    true,
			gAllowFirst:  false, gAllowAllUpper: true, gAllowTitle: false,
			gDefaultTag: disambiguation.TagForNotAddingTags, gIgnoreSpell: true,
			order: "global_mw_xml",
		}
	case "es":
		// SpanishHybridDisambiguator
		return hybridLangProfile{
			mwAllowFirst: true, mwAllowAllUpper: true, mwAllowTitle: false,
			mwRemovePrev: true,
			useGlobal:    true,
			gAllowFirst:  false, gAllowAllUpper: true, gAllowTitle: false,
			gDefaultTag: "NPCN000",
			order:       "global_mw_xml",
		}
	case "pt":
		// PortugueseHybridDisambiguator
		return hybridLangProfile{
			mwAllowFirst: true, mwAllowAllUpper: true, mwAllowTitle: true,
			mwRemovePrev: true, mwIgnoreSpell: true,
			useGlobal: true,
			gAllowFirst: false, gAllowAllUpper: true, gAllowTitle: true,
			gDefaultTag: "NPCN000", gIgnoreSpell: true,
			order: "global_mw_xml",
		}
	case "ca":
		// CatalanHybridDisambiguator (CatalanMultitokenDisambiguator deferred for soft)
		return hybridLangProfile{
			mwAllowFirst: true, mwAllowAllUpper: true, mwAllowTitle: false,
			mwRemovePrev: true,
			useGlobal:    true,
			gAllowFirst:  false, gAllowAllUpper: true, gAllowTitle: false,
			gDefaultTag: "NPCN000",
			order:       "global_mw_xml",
		}
	case "nl":
		// DutchHybridDisambiguator
		return hybridLangProfile{
			mwAllowFirst: true, mwAllowAllUpper: true, mwAllowTitle: false,
			mwDefaultTag: disambiguation.TagForNotAddingTags, mwIgnoreSpell: true,
			useGlobal: true,
			gAllowFirst: false, gAllowAllUpper: true, gAllowTitle: false,
			gDefaultTag: disambiguation.TagForNotAddingTags, gIgnoreSpell: true,
			order: "global_mw_xml",
		}
	case "en":
		// EnglishHybridDisambiguator (also wired via rules/en with ignore lists)
		return hybridLangProfile{
			mwAllowFirst: true, mwAllowAllUpper: true, mwAllowTitle: false,
			mwRemovePrev: true, mwIgnoreSpell: true,
			useGlobal: true,
			gAllowFirst: true, gAllowAllUpper: true, gAllowTitle: false,
			gDefaultTag: disambiguation.TagForNotAddingTags, gIgnoreSpell: true,
			order: "global_mw_xml",
		}
	case "de":
		// GermanRuleDisambiguator: ignore → global → suggest → XML
		return hybridLangProfile{
			useGlobal: true,
			gAllowFirst: false, gAllowAllUpper: true, gAllowTitle: false,
			gDefaultTag: disambiguation.TagForNotAddingTags, gIgnoreSpell: true,
			order: "de",
		}
	case "pl", "sv":
		// PolishHybridDisambiguator / SwedishHybridDisambiguator: XML then multiwords
		return hybridLangProfile{order: "xml_mw"}
	case "ru", "gl", "ga", "sr", "ar", "uk":
		return hybridLangProfile{order: "mw_xml"}
	default:
		return hybridLangProfile{order: "mw_xml"}
	}
}

// RegisterSoftHybridDisambiguator installs a Java-faithful hybrid disambiguator on lt
// for the given language base code (fr, pl, de, …). EN soft path prefers
// rules/en.RegisterSoftEnglishDisambiguator (adds spelling ignore lists).
// Returns true if any step was installed.
func RegisterSoftHybridDisambiguator(lt *languagetool.JLanguageTool, lang string, paths SoftHybridPaths) bool {
	if lt == nil {
		return false
	}
	base := languageBaseCode(lang)
	if base == "" {
		return false
	}
	prof := softHybridProfile(base)

	var global, multi, deIgnore, deSuggest *disambiguation.MultiWordChunker
	var xmlRules *disambigrules.XmlRuleDisambiguator

	if p := paths.SoftDisambigXML; p != "" {
		if data, err := os.ReadFile(p); err == nil {
			loader := disambigrules.NewDisambiguationRuleLoader()
			if rules, err := loader.GetRulesFromString(string(data), base, p); err == nil && len(rules) > 0 {
				xmlRules = disambigrules.NewXmlRuleDisambiguator(rules)
			}
		}
	}

	if p := paths.Multiwords; p != "" && prof.order != "de" {
		if lines, err := loadCachedMultiWordFile(p); err == nil && len(lines) > 0 {
			settings := disambiguation.MultiWordChunkerSettings{
				DefaultTag:            prof.mwDefaultTag,
				AllowFirstCapitalized: prof.mwAllowFirst,
				AllowAllUppercase:     prof.mwAllowAllUpper,
				AllowTitlecase:        prof.mwAllowTitle,
			}
			multi = disambiguation.NewMultiWordChunker(lines, settings)
			if prof.mwRemovePrev {
				multi.SetRemovePreviousTags(true)
			}
			if prof.mwIgnoreSpell {
				multi.SetIgnoreSpelling(true)
			}
		}
	}

	if prof.useGlobal {
		gp := paths.SpellingGlobal
		if gp == "" {
			gp = discoverSoftSpellingGlobalPath()
		}
		if gp != "" {
			if lines, err := loadPhraseOnlyMultiTokenFile(gp); err == nil && len(lines) > 0 {
				global = disambiguation.NewMultiWordChunker(lines, disambiguation.MultiWordChunkerSettings{
					DefaultTag:            prof.gDefaultTag,
					AllowFirstCapitalized: prof.gAllowFirst,
					AllowAllUppercase:     prof.gAllowAllUpper,
					AllowTitlecase:        prof.gAllowTitle,
				})
				if prof.gIgnoreSpell {
					global.SetIgnoreSpelling(true)
				}
			}
		}
	}

	if prof.order == "de" {
		if p := paths.DEMultitokenIgnore; p != "" {
			if lines, err := loadCachedMultiWordFile(p); err == nil && len(lines) > 0 {
				deIgnore = disambiguation.NewMultiWordChunker(lines, disambiguation.MultiWordChunkerSettings{
					DefaultTag:            disambiguation.TagForNotAddingTags,
					AllowFirstCapitalized: true,
					AllowAllUppercase:     true,
					AllowTitlecase:        false,
				})
				deIgnore.SetIgnoreSpelling(true)
			}
		}
		if p := paths.DEMultitokenSuggest; p != "" {
			if lines, err := loadCachedMultiWordFile(p); err == nil && len(lines) > 0 {
				deSuggest = disambiguation.NewMultiWordChunker(lines, disambiguation.MultiWordChunkerSettings{
					DefaultTag:            disambiguation.TagForNotAddingTags,
					AllowFirstCapitalized: true,
					AllowAllUppercase:     true,
					AllowTitlecase:        false,
				})
				deSuggest.SetIgnoreSpelling(true)
			}
		}
	}

	var steps []languagetool.SentenceDisambiguator
	// Append only non-nil concrete pointers. A nil *MultiWordChunker boxed into
	// SentenceDisambiguator is a non-nil interface and would panic on Disambiguate.
	appendChunker := func(c *disambiguation.MultiWordChunker) {
		if c != nil {
			steps = append(steps, c)
		}
	}
	appendXML := func(x *disambigrules.XmlRuleDisambiguator) {
		if x != nil {
			steps = append(steps, x)
		}
	}
	switch prof.order {
	case "global_mw_xml":
		// Java: disambiguator.disambiguate(chunker.disambiguate(chunkerGlobal.disambiguate(input)))
		appendChunker(global)
		appendChunker(multi)
		appendXML(xmlRules)
	case "mw_xml":
		appendChunker(multi)
		appendXML(xmlRules)
	case "xml_mw":
		// Polish/Swedish: chunker.disambiguate(disambiguator.disambiguate(input))
		appendXML(xmlRules)
		appendChunker(multi)
	case "de":
		// multitokenSpeller → global → multitokenSpeller2 → XML
		appendChunker(deIgnore)
		appendChunker(global)
		appendChunker(deSuggest)
		appendXML(xmlRules)
	default:
		appendChunker(multi)
		appendXML(xmlRules)
	}

	if len(steps) == 0 {
		return false
	}
	if len(steps) == 1 {
		lt.Disambiguator = steps[0]
		return true
	}
	lt.Disambiguator = softHybridChain(steps)
	return true
}

// softHybridChain applies steps in order (Java nesting: innermost first).
type softHybridChain []languagetool.SentenceDisambiguator

func (c softHybridChain) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
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

// multiWordFileCache mirrors Java MultiWordChunker.getInstance singleton cache.
var multiWordFileCache sync.Map // path -> []string

func loadCachedMultiWordFile(path string) ([]string, error) {
	if path == "" {
		return nil, os.ErrNotExist
	}
	if v, ok := multiWordFileCache.Load(path); ok {
		if lines, ok := v.([]string); ok {
			out := make([]string, len(lines))
			copy(out, lines)
			return out, nil
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	// Use exported reader constructor so German expander + separator marker are applied.
	chunker, err := disambiguation.NewMultiWordChunkerFromReader(f, disambiguation.MultiWordChunkerSettings{})
	if err != nil {
		return nil, err
	}
	lines := chunker.Lines
	stored := make([]string, len(lines))
	copy(stored, lines)
	multiWordFileCache.Store(path, stored)
	out := make([]string, len(lines))
	copy(out, lines)
	return out, nil
}

func loadPhraseOnlyMultiTokenFile(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
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
		lines = append(lines, line)
	}
	return lines, sc.Err()
}

var (
	softSpellingGlobalOnce sync.Once
	softSpellingGlobalPath string
)

func discoverSoftSpellingGlobalPath() string {
	softSpellingGlobalOnce.Do(func() {
		wd, err := os.Getwd()
		if err != nil {
			return
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
					softSpellingGlobalPath = cand
					return
				}
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				return
			}
			dir = parent
		}
	})
	return softSpellingGlobalPath
}
