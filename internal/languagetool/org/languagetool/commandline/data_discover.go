package commandline

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/en"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// WalkUpFind walks from start (or cwd) toward root looking for relPath.
// Soft data discovery (SPEC §10 nicer data discovery).
func WalkUpFind(start, relPath string) string {
	if relPath == "" {
		return ""
	}
	dir := start
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			return ""
		}
	}
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, relPath)
		if st, err := os.Stat(cand); err == nil && (st.IsDir() || st.Mode().IsRegular()) {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// DiscoverFalseFriendsFile finds false-friends XML via env/data-dir/walk-up.
// Official upstream only (no false-friends-soft invent path).
func DiscoverFalseFriendsFile(opts *CommandLineOptions) string {
	if p := resolveFalseFriendsFile(opts); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p
		}
		if opts != nil && (opts.GetDataDir() != "" || opts.FalseFriendsFile != "") {
			return p
		}
	}
	// Prefer Java classpath resource over testdata extracts (may drop DTD / lag upstream).
	for _, rel := range []string{
		filepath.Join("inspiration", "languagetool", "languagetool-core", "src", "main", "resources", "org", "languagetool", "rules", "false-friends.xml"),
		filepath.Join("testdata", "upstream", "false-friends.xml"),
		filepath.Join("testdata", "upstream", "false-friends-nodtd.xml"),
	} {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}

// DiscoverEnglishUSDict finds en_US.dict for the binary Morfologik speller.
// Order: LANG_EN_US_DICT, --data-dir/en/hunspell/en_US.dict, walk-up third_party and inspiration.
func DiscoverEnglishUSDict(opts *CommandLineOptions) string {
	return DiscoverEnglishVariantDict(opts, "en-US")
}

// DiscoverEnglishVariantDict finds the CFSA2 hunspell dict for an English locale
// (Java Morfologik*SpellerRule resource). Falls back to en_US when the variant
// file is missing. LANG_EN_US_DICT still forces a path for any locale.
func DiscoverEnglishVariantDict(opts *CommandLineOptions, lang string) string {
	if p := os.Getenv("LANG_EN_US_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	_, dictFile := en.EnglishVariantSpellerMeta(lang)
	if opts != nil && opts.GetDataDir() != "" {
		cand := filepath.Join(opts.GetDataDir(), "en", "hunspell", dictFile)
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		// also flat layout
		cand = filepath.Join(opts.GetDataDir(), dictFile)
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	if p := en.DiscoverEnglishVariantDictFile(dictFile); p != "" {
		return p
	}
	// No invent fallback to a different locale dict (wrong Java resource for rule ID).
	return ""
}

// DiscoverEnglishMultiwords finds official /en/multiwords.txt (Java MultiWordChunker path).
// Prefers LANG_EN_MULTIWORDS, data-dir, vendored upstream, inspiration submodule.
// Soft multiword lists are not used.
func DiscoverEnglishMultiwords(opts *CommandLineOptions) string {
	if p := os.Getenv("LANG_EN_MULTIWORDS"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), "en", "multiwords.txt"),
			filepath.Join(opts.GetDataDir(), "upstream", "en", "resource", "multiwords.txt"),
			filepath.Join(opts.GetDataDir(), "multiwords.txt"),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	// Prefer inspiration (Java module multiwords.txt) over testdata extracts.
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en", "src", "main", "resources", "org", "languagetool", "resource", "en", "multiwords.txt"),
		filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "multiwords.txt"),
		filepath.Join("testdata", "upstream", "en", "resource", "multiwords.txt"),
		filepath.Join("testdata", "disambiguation", "en-multiwords-upstream.txt"),
	}
	for _, rel := range relPaths {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}

// DiscoverEnglishPOSDict finds english.dict for the binary POS tagger.
// Order: LANG_ENGLISH_DICT, --data-dir/en/english.dict, walk-up third_party and inspiration.
func DiscoverEnglishPOSDict(opts *CommandLineOptions) string {
	if p := os.Getenv("LANG_ENGLISH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), "en", "english.dict"),
			filepath.Join(opts.GetDataDir(), "english.dict"),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	relPaths := []string{
		filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "english.dict"),
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en", "src", "main", "resources", "org", "languagetool", "resource", "en", "english.dict"),
	}
	for _, rel := range relPaths {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}

// DiscoverEnglishSynthDict finds english_synth.dict for EnglishSynthesizer
// (Java BaseSynthesizer resource /en/english_synth.dict).
// Order: LANG_ENGLISH_SYNTH_DICT, --data-dir, walk-up third_party / inspiration.
func DiscoverEnglishSynthDict(opts *CommandLineOptions) string {
	if p := os.Getenv("LANG_ENGLISH_SYNTH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), "en", "english_synth.dict"),
			filepath.Join(opts.GetDataDir(), "english_synth.dict"),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	// Prefer sibling of POS dict when discovered
	if pos := DiscoverEnglishPOSDict(opts); pos != "" {
		cand := filepath.Join(filepath.Dir(pos), "english_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
	}
	relPaths := []string{
		filepath.Join("third_party", "english-pos-dict", "org", "languagetool", "resource", "en", "english_synth.dict"),
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en", "src", "main", "resources", "org", "languagetool", "resource", "en", "english_synth.dict"),
	}
	for _, rel := range relPaths {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}

// languageSynthDictNames maps short code → official *_synth.dict basename
// (Java *Synthesizer.RESOURCE_FILENAME under org/languagetool/resource/{lang}/).
// Names match Java createDefaultSynthesizer resources exactly — not invent.
var languageSynthDictNames = map[string]string{
	"en":  "english_synth.dict",
	"pl":  "polish_synth.dict",
	"it":  "italian_synth.dict",
	"ru":  "russian_synth.dict",
	"ro":  "romanian_synth.dict",
	"sk":  "slovak_synth.dict",
	"sv":  "swedish_synth.dict",
	"el":  "greek_synth.dict",
	"gl":  "galician_synth.dict",
	"ar":  "arabic_synth.dict",
	"uk":  "ukrainian_synth.dict",
	"nl":  "dutch_synth.dict",         // DutchSynthesizer
	"fr":  "french_synth.dict",        // FrenchSynthesizer
	"de":  "german_synth.dict",        // GermanSynthesizer
	"pt":  "portuguese_synth.dict",    // PortugueseSynthesizer
	"ga":  "irish_synth.dict",         // IrishSynthesizer
	"crh": "crimean_tatar_synth.dict", // CrimeanTatarSynthesizer
	"ca":  "ca-ES_synth.dict",         // CatalanSynthesizer (/ca/ca-ES_synth.dict)
	"es":  "es-ES_synth.dict",         // SpanishSynthesizer (/es/es-ES_synth.dict)
	"sr":  "serbian_synth.dict",       // Ekavian/Jekavian under dictionary/
}

// languageSynthInspirationRels returns walk-up relative paths for Java resource layout.
// Serbian lives under resource/sr/dictionary/{ekavian,jekavian}/ (not resource/sr/).
func languageSynthInspirationRels(base, name string) []string {
	mod := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", base,
		"src", "main", "resources", "org", "languagetool", "resource")
	switch base {
	case "sr":
		// Java EkavianSynthesizer first (Serbian default), then Jekavian, then flat fallback.
		return []string{
			filepath.Join(mod, "sr", "dictionary", "ekavian", name),
			filepath.Join(mod, "sr", "dictionary", "jekavian", name),
			filepath.Join(mod, "sr", name),
		}
	default:
		return []string{filepath.Join(mod, base, name)}
	}
}

// DiscoverLanguageSynthDict finds *_synth.dict for lang (Java createDefaultSynthesizer resource).
// Order: LANG_{CODE}_SYNTH_DICT, sibling of POS dict, --data-dir, inspiration module resource.
func DiscoverLanguageSynthDict(opts *CommandLineOptions, lang string) string {
	base := lang
	if i := strings.IndexByte(lang, '-'); i > 0 {
		base = lang[:i]
	}
	base = strings.ToLower(base)
	if base == "" {
		return ""
	}
	if base == "en" {
		return DiscoverEnglishSynthDict(opts)
	}
	envKey := "LANG_" + strings.ToUpper(base) + "_SYNTH_DICT"
	if p := os.Getenv(envKey); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	name := languageSynthDictNames[base]
	if name == "" {
		return ""
	}
	// Sibling of POS dict directory (same resource/{lang}/ folder as Java).
	if pos := DiscoverLanguagePOSDict(opts, base); pos != "" {
		cand := filepath.Join(filepath.Dir(pos), name)
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			return cand
		}
		// Serbian POS may live under dictionary/ekavian; try sibling dirs.
		if base == "sr" {
			for _, sub := range []string{
				filepath.Join(filepath.Dir(pos), name),
				filepath.Join(filepath.Dir(pos), "ekavian", name),
				filepath.Join(filepath.Dir(filepath.Dir(pos)), "ekavian", name),
			} {
				if st, err := os.Stat(sub); err == nil && st.Mode().IsRegular() {
					return sub
				}
			}
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), base, name),
			filepath.Join(opts.GetDataDir(), name),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
		if base == "sr" {
			for _, rel := range []string{
				filepath.Join(opts.GetDataDir(), "sr", "dictionary", "ekavian", name),
				filepath.Join(opts.GetDataDir(), "sr", "dictionary", "jekavian", name),
			} {
				if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
					return rel
				}
			}
		}
	}
	for _, rel := range languageSynthInspirationRels(base, name) {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}

// DiscoverLanguagePOSDict finds a Morfologik POS dict for lang (e.g. "da" → danish.dict).
// Order: LANG_{CODE}_DICT env, --data-dir/{code}/*.dict, walk-up inspiration resource path.
func DiscoverLanguagePOSDict(opts *CommandLineOptions, lang string) string {
	base := lang
	if i := strings.IndexByte(lang, '-'); i > 0 {
		base = lang[:i]
	}
	base = strings.ToLower(base)
	if base == "" {
		return ""
	}
	envKey := "LANG_" + strings.ToUpper(base) + "_DICT"
	if p := os.Getenv(envKey); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	// Common dict basenames used by upstream modules
	names := languagePOSDictNames(base)
	if opts != nil && opts.GetDataDir() != "" {
		for _, name := range names {
			for _, rel := range []string{
				filepath.Join(opts.GetDataDir(), base, name),
				filepath.Join(opts.GetDataDir(), name),
			} {
				if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
					return rel
				}
			}
		}
	}
	for _, name := range names {
		rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", base,
			"src", "main", "resources", "org", "languagetool", "resource", base, name)
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}

// languagePOSDictNames maps ISO codes to upstream Morfologik POS basenames
// under resource/{code}/ (see languagetool-language-modules).
func languagePOSDictNames(base string) []string {
	switch base {
	case "ar":
		return []string{"arabic.dict"}
	case "br":
		return []string{"breton.dict"}
	case "ca":
		// Java CatalanTagger: /ca/ca-ES.dict (also historical catalan.dict name).
		return []string{"ca-ES.dict", "catalan.dict"}
	case "da":
		return []string{"danish.dict"}
	case "de":
		return []string{"german.dict"}
	case "el":
		return []string{"greek.dict"}
	case "en":
		return []string{"english.dict"}
	case "es":
		// Java SpanishTagger: /es/es-ES.dict (also historical spanish.dict name).
		return []string{"es-ES.dict", "spanish.dict"}
	case "fr":
		return []string{"french.dict"}
	case "gl":
		return []string{"galician.dict"}
	case "it":
		return []string{"italian.dict"}
	case "km":
		return []string{"khmer.dict"}
	case "ml":
		return []string{"malayalam.dict"}
	case "nl":
		return []string{"dutch.dict"}
	case "pl":
		return []string{"polish.dict"}
	case "pt":
		return []string{"portuguese.dict"}
	case "ro":
		return []string{"romanian.dict"}
	case "ru":
		return []string{"russian.dict"}
	case "sk":
		return []string{"slovak.dict"}
	// Java SerbianTagger: /sr/dictionary/ekavian/serbian.dict (default ekavian).
	case "sr":
		return []string{
			filepath.Join("dictionary", "ekavian", "serbian.dict"),
			"serbian.dict",
		}
	case "sv":
		return []string{"swedish.dict"}
	case "ta":
		return []string{"tamil.dict"}
	case "tl":
		return []string{"tagalog.dict"}
	default:
		return []string{base + ".dict"}
	}
}

// languageBaseCode returns the ISO base (en from en-US).
func languageBaseCode(lang string) string {
	base := lang
	if i := strings.IndexByte(lang, '-'); i > 0 {
		base = lang[:i]
	}
	return strings.ToLower(base)
}

// DiscoverLanguageMultiwords finds official multiwords.txt for lang
// (vendored {lang}-multiwords-upstream.txt or inspiration resource).
func DiscoverLanguageMultiwords(opts *CommandLineOptions, lang string) string {
	base := languageBaseCode(lang)
	if base == "" {
		return ""
	}
	envKey := "LANG_" + strings.ToUpper(base) + "_MULTIWORDS"
	if p := os.Getenv(envKey); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), "disambiguation", base+"-multiwords-upstream.txt"),
			filepath.Join(opts.GetDataDir(), base, "multiwords.txt"),
			filepath.Join(opts.GetDataDir(), "multiwords.txt"),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	// Prefer inspiration Java multiwords.txt over incomplete testdata copies.
	for _, rel := range []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", base,
			"src", "main", "resources", "org", "languagetool", "resource", base, "multiwords.txt"),
		filepath.Join("testdata", "upstream", base, "resource", "multiwords.txt"),
		filepath.Join("testdata", "disambiguation", base+"-multiwords-upstream.txt"),
	} {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}

// DiscoverGermanMultitokenIgnore finds de/multitoken-ignore.txt (Java GermanRuleDisambiguator).
func DiscoverGermanMultitokenIgnore(opts *CommandLineOptions) string {
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), "de", "multitoken-ignore.txt"),
			filepath.Join(opts.GetDataDir(), "resource", "de", "multitoken-ignore.txt"),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	return WalkUpFind("", filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "de",
		"src", "main", "resources", "org", "languagetool", "resource", "de", "multitoken-ignore.txt"))
}

// DiscoverGermanMultitokenSuggest finds de/multitoken-suggest.txt (Java GermanRuleDisambiguator).
func DiscoverGermanMultitokenSuggest(opts *CommandLineOptions) string {
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), "de", "multitoken-suggest.txt"),
			filepath.Join(opts.GetDataDir(), "resource", "de", "multitoken-suggest.txt"),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	return WalkUpFind("", filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "de",
		"src", "main", "resources", "org", "languagetool", "resource", "de", "multitoken-suggest.txt"))
}

// DiscoverEnglishL2GrammarXML finds EN L2 pattern rules for a mother tongue.
// Java English.getRelevantRules: grammar-l2-de.xml / grammar-l2-fr.xml when motherTongue is de/fr.
func DiscoverEnglishL2GrammarXML(opts *CommandLineOptions, motherTongue string) string {
	mt := languageBaseCode(motherTongue)
	if mt != "de" && mt != "fr" {
		return ""
	}
	fileName := "grammar-l2-" + mt + ".xml"
	if p := os.Getenv("LANG_EN_L2_GRAMMAR"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), "en", fileName),
			filepath.Join(opts.GetDataDir(), "rules", "en", fileName),
			filepath.Join(opts.GetDataDir(), "upstream", "en", "rules", fileName),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	// Prefer inspiration Java rules/{en}/ over testdata (SYSTEM .ent + full grammar).
	for _, rel := range []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
			"src", "main", "resources", "org", "languagetool", "rules", "en", fileName),
		filepath.Join("testdata", "upstream", "en", "rules", fileName),
	} {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}

// DiscoverLanguageGrammarXML finds official grammar.xml for lang.
// Java: /org/languagetool/rules/{lang}/grammar.xml
func DiscoverLanguageGrammarXML(opts *CommandLineOptions, lang string) string {
	return discoverLanguageRuleXML(opts, lang, "grammar.xml", "GRAMMAR")
}

// DiscoverLanguageStyleXML finds official style.xml for lang.
// Java: Language.getRuleFileNames() also loads /org/languagetool/rules/{lang}/style.xml
func DiscoverLanguageStyleXML(opts *CommandLineOptions, lang string) string {
	return discoverLanguageRuleXML(opts, lang, "style.xml", "STYLE")
}

// languageExtraRuleFiles ports Language subclasses that append RULE_FILES after
// super.getRuleFileNames() (grammar.xml / style.xml / …). Only existing files
// are discovered — no invent. Order matches Java Arrays.asList order.
var languageExtraRuleFiles = map[string][]string{
	// Ukrainian.RULE_FILES
	"uk": {
		"grammar-spelling.xml",
		"grammar-grammar.xml",
		"grammar-barbarism.xml",
		"grammar-style.xml",
		"grammar-punctuation.xml",
	},
	// Slovak.RULE_FILES
	"sk": {
		"grammar-typography.xml",
	},
	// Serbian.RULE_FILES
	"sr": {
		"grammar-barbarism.xml",
		"grammar-logical.xml",
		"grammar-punctuation.xml",
		"grammar-spelling.xml",
		"grammar-style.xml",
	},
}

// DiscoverLanguagePatternRuleFiles ports Language.getRuleFileNames() path order:
//
//	{lang}/grammar.xml, {lang}/style.xml (if any), {lang}/grammar_custom.xml (if any),
//	language-specific RULE_FILES extras (uk/sk/sr, if present),
//	and when lang has a country variant (e.g. en-US):
//	{lang}/{variant}/grammar.xml, style.xml, grammar-premium.xml (each if present).
//
// Only existing official files are returned — no soft invent paths.
func DiscoverLanguagePatternRuleFiles(opts *CommandLineOptions, lang string) []string {
	base := languageBaseCode(lang)
	if base == "" {
		return nil
	}
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
	// Java always adds grammar; style/custom only when they exist.
	add(discoverLanguageRuleXML(opts, base, "grammar.xml", "GRAMMAR"))
	if p := discoverLanguageRuleXML(opts, base, "style.xml", "STYLE"); p != "" {
		add(p)
	}
	if p := discoverLanguageRuleXML(opts, base, "grammar_custom.xml", ""); p != "" {
		add(p)
	}
	// Java Language overrides (Ukrainian/Slovak/Serbian) append RULE_FILES next.
	for _, name := range languageExtraRuleFiles[base] {
		// empty env suffix — no per-file LANG_*_GRAMMAR invent
		if p := discoverLanguageRuleXML(opts, base, name, ""); p != "" {
			add(p)
		}
	}
	// Variant files: shortCodeWithCountryAndVariant length > 2 in Java.
	variant := languageVariantCode(lang)
	if variant != "" && variant != base {
		add(discoverLanguageVariantRuleXML(opts, base, variant, "grammar.xml"))
		add(discoverLanguageVariantRuleXML(opts, base, variant, "style.xml"))
		add(discoverLanguageVariantRuleXML(opts, base, variant, "grammar-premium.xml"))
	}
	return out
}

// languageVariantCode returns the full short code with country when present (e.g. en-US).
func languageVariantCode(lang string) string {
	lang = tools.JavaStringTrim(lang)
	if lang == "" {
		return ""
	}
	// normalize underscore variants (en_US → en-US)
	lang = strings.ReplaceAll(lang, "_", "-")
	parts := strings.Split(lang, "-")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return ""
	}
	// en-US, de-DE, pt-BR — keep first two segments lower-Upper when 2-letter country
	base := strings.ToLower(parts[0])
	region := parts[1]
	if len(region) == 2 {
		region = strings.ToUpper(region)
	}
	return base + "-" + region
}

// discoverLanguageRuleXML finds rules/{lang}/{fileName} (grammar.xml / style.xml).
func discoverLanguageRuleXML(opts *CommandLineOptions, lang, fileName, envSuffix string) string {
	base := languageBaseCode(lang)
	if base == "" || fileName == "" {
		return ""
	}
	if envSuffix != "" {
		if p := os.Getenv("LANG_" + strings.ToUpper(base) + "_" + envSuffix); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), base, fileName),
			filepath.Join(opts.GetDataDir(), "rules", base, fileName),
			filepath.Join(opts.GetDataDir(), "upstream", base, "rules", fileName),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	// Prefer inspiration Java rules over testdata extracts (entities + currency).
	for _, rel := range []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", base,
			"src", "main", "resources", "org", "languagetool", "rules", base, fileName),
		filepath.Join("testdata", "upstream", base, "rules", fileName),
	} {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}

// discoverLanguageVariantRuleXML finds rules/{base}/{variant}/{fileName}
// e.g. en/en-US/grammar.xml (Java getShortCode()+"/"+getShortCodeWithCountryAndVariant()+"/"+file).
func discoverLanguageVariantRuleXML(opts *CommandLineOptions, base, variant, fileName string) string {
	if base == "" || variant == "" || fileName == "" {
		return ""
	}
	if opts != nil && opts.GetDataDir() != "" {
		for _, rel := range []string{
			filepath.Join(opts.GetDataDir(), base, variant, fileName),
			filepath.Join(opts.GetDataDir(), "rules", base, variant, fileName),
			filepath.Join(opts.GetDataDir(), "upstream", base, "rules", variant, fileName),
		} {
			if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
				return rel
			}
		}
	}
	for _, rel := range []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", base,
			"src", "main", "resources", "org", "languagetool", "rules", base, variant, fileName),
		filepath.Join("testdata", "upstream", base, "rules", variant, fileName),
	} {
		if p := WalkUpFind("", rel); p != "" {
			return p
		}
	}
	return ""
}
