package commandline

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
	cadis "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation/ca"
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

// hybridInstanceCache mirrors Java MultiWordChunker.getInstance / hybrid field
// singletons: rebuild once per (lang, paths), reuse across configureCoreLT calls
// (miss scans create many JLanguageTools).
var hybridInstanceCache sync.Map // cacheKey string -> languagetool.SentenceDisambiguator

func softHybridCacheKey(base string, paths SoftHybridPaths) string {
	gp := paths.SpellingGlobal
	if gp == "" {
		gp = discoverSoftSpellingGlobalPath()
	}
	// Remap mode depends on POS dict availability (no invent morph when untagged soft).
	pos := DiscoverLanguagePOSDict(nil, base)
	return strings.Join([]string{
		base,
		paths.Multiwords,
		paths.SoftDisambigXML,
		paths.DEMultitokenIgnore,
		paths.DEMultitokenSuggest,
		gp,
		pos,
	}, "\x00")
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
	key := softHybridCacheKey(base, paths)
	if v, ok := hybridInstanceCache.Load(key); ok {
		if d, ok := v.(languagetool.SentenceDisambiguator); ok && d != nil {
			lt.Disambiguator = d
			return true
		}
	}
	chain, ok := buildSoftHybridDisambiguator(base, paths)
	if !ok || chain == nil {
		return false
	}
	hybridInstanceCache.Store(key, chain)
	lt.Disambiguator = chain
	return true
}

// buildSoftHybridDisambiguator constructs the hybrid once (Java HybridDisambiguator fields).
func buildSoftHybridDisambiguator(base string, paths SoftHybridPaths) (languagetool.SentenceDisambiguator, bool) {
	prof := softHybridProfile(base)

	var global, multi, deIgnore, deSuggest *disambiguation.MultiWordChunker
	var xmlRules *disambigrules.XmlRuleDisambiguator

	if p := paths.SoftDisambigXML; p != "" {
		xmlRules = getCachedSoftXMLDisambiguator(base, p)
	}

	if p := paths.Multiwords; p != "" && prof.order != "de" {
		// Without a FreeLing/Morfologik POS dict, single-letter multiword tags
		// (e.g. FR "a priori\tA") replace soft-untagged surfaces and block
		// grammar postag soft-accept (CE_CET). Map those tags to TagForNotAddingTags
		// so multiword still ignore-spells without inventing morph POS. Multi-token
		// tags like "N f s" stay real for home-page style chunking.
		remapSingleLetter := DiscoverLanguagePOSDict(nil, base) == "" && prof.mwDefaultTag == ""
		multi = getCachedMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
			DefaultTag:            prof.mwDefaultTag,
			AllowFirstCapitalized: prof.mwAllowFirst,
			AllowAllUppercase:     prof.mwAllowAllUpper,
			AllowTitlecase:        prof.mwAllowTitle,
		}, prof.mwRemovePrev, prof.mwIgnoreSpell, false, remapSingleLetter)
	}

	if prof.useGlobal {
		gp := paths.SpellingGlobal
		if gp == "" {
			gp = discoverSoftSpellingGlobalPath()
		}
		if gp != "" {
			global = getCachedMultiWordChunker(gp, disambiguation.MultiWordChunkerSettings{
				DefaultTag:            prof.gDefaultTag,
				AllowFirstCapitalized: prof.gAllowFirst,
				AllowAllUppercase:     prof.gAllowAllUpper,
				AllowTitlecase:        prof.gAllowTitle,
			}, false, prof.gIgnoreSpell, true, false)
		}
	}

	if prof.order == "de" {
		if p := paths.DEMultitokenIgnore; p != "" {
			deIgnore = getCachedMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
				DefaultTag:            disambiguation.TagForNotAddingTags,
				AllowFirstCapitalized: true,
				AllowAllUppercase:     true,
				AllowTitlecase:        false,
			}, false, true, false, false)
		}
		if p := paths.DEMultitokenSuggest; p != "" {
			deSuggest = getCachedMultiWordChunker(p, disambiguation.MultiWordChunkerSettings{
				DefaultTag:            disambiguation.TagForNotAddingTags,
				AllowFirstCapitalized: true,
				AllowAllUppercase:     true,
				AllowTitlecase:        false,
			}, false, true, false, false)
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

	// Java CatalanHybridDisambiguator: multitokenDisambiguator after XML rules.
	// Soft: known multi-token phrases from multiwords.txt + spelling_global (not invent lists).
	if base == "ca" {
		if mt := softCatalanMultitokenDisambiguator(paths); mt != nil {
			steps = append(steps, mt)
		}
	}

	if len(steps) == 0 {
		return nil, false
	}
	if len(steps) == 1 {
		return steps[0], true
	}
	return softHybridChain(steps), true
}

// softCatalanMultitokenDisambiguator ports CatalanMultitokenDisambiguator with
// IsMisspelled backed by official multiword / spelling_global phrase sets
// (Java uses CatalanMorfologikMultitokenSpeller over multitoken data).
func softCatalanMultitokenDisambiguator(paths SoftHybridPaths) *cadis.CatalanMultitokenDisambiguator {
	known := softLoadKnownMultiTokenPhrases(paths.Multiwords)
	gp := paths.SpellingGlobal
	if gp == "" {
		gp = discoverSoftSpellingGlobalPath()
	}
	for p := range softLoadKnownMultiTokenPhrases(gp) {
		known[p] = struct{}{}
	}
	if len(known) == 0 {
		return nil
	}
	mt := cadis.NewCatalanMultitokenDisambiguator()
	mt.IsMisspelled = func(phrase string) bool {
		if phrase == "" {
			return true
		}
		if _, ok := known[phrase]; ok {
			return false
		}
		if _, ok := known[strings.ToLower(phrase)]; ok {
			return false
		}
		return true
	}
	return mt
}

// softLoadKnownMultiTokenPhrases loads phrase keys (before tab/semicolon) from
// multiwords or spelling lists; multi-token phrases only.
func softLoadKnownMultiTokenPhrases(path string) map[string]struct{} {
	out := map[string]struct{}{}
	if path == "" {
		return out
	}
	lines, err := loadCachedMultiWordFile(path)
	if err != nil {
		// spelling_global is phrase-only without separator markers
		pl, err2 := loadPhraseOnlyMultiTokenFile(path)
		if err2 != nil {
			return out
		}
		for _, line := range pl {
			if strings.Contains(line, " ") {
				out[line] = struct{}{}
				out[strings.ToLower(line)] = struct{}{}
			}
		}
		return out
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		phrase := line
		if i := strings.IndexByte(line, '\t'); i >= 0 {
			phrase = line[:i]
		} else if i := strings.IndexByte(line, ';'); i >= 0 {
			phrase = line[:i]
		}
		phrase = strings.TrimSpace(phrase)
		if phrase == "" || !strings.Contains(phrase, " ") {
			continue
		}
		out[phrase] = struct{}{}
		out[strings.ToLower(phrase)] = struct{}{}
	}
	return out
}

// multiWordChunkerInstanceCache: path+settings → *MultiWordChunker (Java getInstance).
var multiWordChunkerInstanceCache sync.Map

type multiWordChunkerCacheKey struct {
	path              string
	defaultTag        string
	allowFirst        bool
	allowAllUp        bool
	allowTitle        bool
	removePrev        bool
	ignoreSpell       bool
	phraseOnly        bool // spelling_global: multi-token phrases only
	remapSingleLetter bool // soft no-tagger: A → TagForNotAddingTags
}

func getCachedMultiWordChunker(
	path string,
	settings disambiguation.MultiWordChunkerSettings,
	removePrev, ignoreSpell, phraseOnly, remapSingleLetter bool,
) *disambiguation.MultiWordChunker {
	if path == "" {
		return nil
	}
	key := multiWordChunkerCacheKey{
		path: path, defaultTag: settings.DefaultTag,
		allowFirst: settings.AllowFirstCapitalized, allowAllUp: settings.AllowAllUppercase,
		allowTitle: settings.AllowTitlecase, removePrev: removePrev, ignoreSpell: ignoreSpell,
		phraseOnly: phraseOnly, remapSingleLetter: remapSingleLetter,
	}
	if v, ok := multiWordChunkerInstanceCache.Load(key); ok {
		if c, ok := v.(*disambiguation.MultiWordChunker); ok {
			return c
		}
	}
	var lines []string
	var err error
	if phraseOnly {
		lines, err = loadPhraseOnlyMultiTokenFile(path)
	} else {
		lines, err = loadCachedMultiWordFile(path)
	}
	if err != nil || len(lines) == 0 {
		return nil
	}
	if remapSingleLetter {
		lines = remapSingleLetterMultiwordTags(lines)
	}
	c := disambiguation.NewMultiWordChunker(lines, settings)
	if removePrev {
		c.SetRemovePreviousTags(true)
	}
	if ignoreSpell {
		c.SetIgnoreSpelling(true)
	}
	// Warm maps once (Java lazyInit on first use); share read-only after this.
	warmMultiWordChunker(c)
	multiWordChunkerInstanceCache.Store(key, c)
	return c
}

// remapSingleLetterMultiwordTags maps FreeLing single-letter POS tags (A, S, …)
// to TagForNotAddingTags for soft runs without a real POS dictionary.
// Keeps multi-token tags (N f s, Z m s, …) for real multiword POS chunking.
func remapSingleLetterMultiwordTags(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, "#separatorRegExp=") {
			out = append(out, line)
			continue
		}
		// Prefer tab, then semicolon (after soft separator handling in fillMaps).
		sep := "\t"
		i := strings.IndexByte(line, '\t')
		if i < 0 {
			i = strings.IndexByte(line, ';')
			sep = ";"
		}
		if i < 0 {
			out = append(out, line)
			continue
		}
		phrase := line[:i]
		tag := strings.TrimSpace(line[i+1:])
		// Single Latin letter FreeLing tags only (A, S, P, I, …).
		if len(tag) == 1 {
			r := tag[0]
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				out = append(out, phrase+sep+disambiguation.TagForNotAddingTags)
				continue
			}
		}
		out = append(out, line)
	}
	return out
}

func warmMultiWordChunker(c *disambiguation.MultiWordChunker) {
	if c == nil {
		return
	}
	// Disambiguate a trivial sentence so fillMaps runs under the chunker's mutex.
	tag := languagetool.SentenceStartTagName
	toks := []*languagetool.AnalyzedTokenReadings{
		languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("", &tag, nil)),
	}
	_ = c.Disambiguate(languagetool.NewAnalyzedSentence(toks))
}

// softXMLDisambigCache: lang\0path → *XmlRuleDisambiguator
var softXMLDisambigCache sync.Map

func getCachedSoftXMLDisambiguator(lang, path string) *disambigrules.XmlRuleDisambiguator {
	if path == "" {
		return nil
	}
	key := lang + "\x00" + path
	if v, ok := softXMLDisambigCache.Load(key); ok {
		if x, ok := v.(*disambigrules.XmlRuleDisambiguator); ok {
			return x
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	loader := disambigrules.NewDisambiguationRuleLoader()
	rules, err := loader.GetRulesFromString(string(data), lang, path)
	if err != nil || len(rules) == 0 {
		return nil
	}
	x := disambigrules.NewXmlRuleDisambiguator(rules)
	softXMLDisambigCache.Store(key, x)
	return x
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
