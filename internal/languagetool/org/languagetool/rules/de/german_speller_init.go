package de

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
	compoundtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/de"
)

// DiscoverGermanResourceDir finds inspiration LT resource/de directory by walking
// upward from the working directory. Empty if not found (fail-closed, no invent).
func DiscoverGermanResourceDir() string {
	const rel = "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de"
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, rel)
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
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

// DiscoverGermanPOSDict finds de/german.dict (POS tagger binary) by walking upward.
// Empty if not present (not shipped in all checkouts).
func DiscoverGermanPOSDict() string {
	const rel = "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/resource/de/german.dict"
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, rel)
		if fileExists(cand) {
			return cand
		}
		// also accept testdata/upstream layout
		cand2 := filepath.Join(dir, "testdata/upstream/de/resource/german.dict")
		if fileExists(cand2) {
			return cand2
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// beside DiscoverGermanResourceDir
	if root := DiscoverGermanResourceDir(); root != "" {
		p := filepath.Join(root, "german.dict")
		if fileExists(p) {
			return p
		}
	}
	return ""
}

// DiscoverGermanHunspellDict finds de/hunspell/de_DE.dict (or AT/CH) for filter spelling.
// Empty if missing (fail-closed).
func DiscoverGermanHunspellDict(variant string) string {
	root := DiscoverGermanResourceDir()
	if root == "" {
		return ""
	}
	name := "de_DE.dict"
	switch strings.ToUpper(variant) {
	case "AT":
		name = "de_AT.dict"
	case "CH":
		name = "de_CH.dict"
	}
	p := filepath.Join(root, "hunspell", name)
	if fileExists(p) {
		return p
	}
	// fall back to de_DE when variant dict missing
	if name != "de_DE.dict" {
		p = filepath.Join(root, "hunspell", "de_DE.dict")
		if fileExists(p) {
			return p
		}
	}
	return ""
}

// DiscoverGermanRemoteRuleFiltersXML finds official rules/de/remote-rule-filters.xml
// (Java RemoteRuleFilters.getFilename → de/remote-rule-filters.xml). Empty if not present.
func DiscoverGermanRemoteRuleFiltersXML() string {
	const rel = "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/rules/de/remote-rule-filters.xml"
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, rel)
		if fileExists(cand) {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if root := DiscoverGermanResourceDir(); root != "" {
		p := filepath.Clean(filepath.Join(root, "..", "rules", "de", "remote-rule-filters.xml"))
		if fileExists(p) {
			return p
		}
	}
	return ""
}

// DiscoverGermanGrammarXML finds official rules/de/grammar.xml (Java getRuleFileNames).
// Empty if not present.
func DiscoverGermanGrammarXML() string {
	const rel = "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/rules/de/grammar.xml"
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, rel)
		if fileExists(cand) {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	// sibling of resource/de → rules/de/grammar.xml
	if root := DiscoverGermanResourceDir(); root != "" {
		p := filepath.Clean(filepath.Join(root, "..", "rules", "de", "grammar.xml"))
		if fileExists(p) {
			return p
		}
	}
	return ""
}

// DiscoverGermanDEATGrammarXML finds rules/de/de-DE-AT/grammar.xml.
// Java GermanyGerman / AustrianGerman / NonSwissGerman.getRuleFileNames add this
// (rules for de-DE and de-AT but not de-CH). Empty if not present.
func DiscoverGermanDEATGrammarXML() string {
	return discoverGermanRulesFile("de-DE-AT", "grammar.xml")
}

// DiscoverGermanCHGrammarXML finds rules/de/de-CH/grammar.xml.
// Java Language.getRuleFileNames loads shortCode/shortCodeWithCountryAndVariant/grammar.xml
// when the variant short code is longer than 2 (SwissGerman → de/de-CH/grammar.xml).
func DiscoverGermanCHGrammarXML() string {
	return discoverGermanRulesFile("de-CH", "grammar.xml")
}

// discoverGermanRulesFile finds rules/de/<subdir>/<name> under inspiration or resource sibling.
func discoverGermanRulesFile(subdir, name string) string {
	rel := filepath.Join(
		"inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/rules/de",
		subdir, name,
	)
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, rel)
		if fileExists(cand) {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if root := DiscoverGermanResourceDir(); root != "" {
		// resource/de → rules/de/<subdir>/<name>
		p := filepath.Clean(filepath.Join(root, "..", "rules", "de", subdir, name))
		if fileExists(p) {
			return p
		}
	}
	return ""
}

// DiscoverGermanStyleXML finds official rules/de/style.xml when present.
func DiscoverGermanStyleXML() string {
	const rel = "inspiration/languagetool/languagetool-language-modules/de/src/main/resources/org/languagetool/rules/de/style.xml"
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, rel)
		if fileExists(cand) {
			return cand
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if root := DiscoverGermanResourceDir(); root != "" {
		p := filepath.Clean(filepath.Join(root, "..", "rules", "de", "style.xml"))
		if fileExists(p) {
			return p
		}
	}
	return ""
}

// DiscoverSpellingGlobal finds resource/spelling_global.txt (all-language accepted words).
func DiscoverSpellingGlobal() string {
	const rel = "inspiration/languagetool/languagetool-core/src/main/resources/org/languagetool/resource/spelling_global.txt"
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for {
		cand := filepath.Join(dir, rel)
		if fileExists(cand) {
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

// InitFromDiscoveredResources ports GermanSpellerRule / SpellingCheckRule.init()
// resource loads when the official resource tree is present on disk.
// Missing paths are skipped (partial init). Wires de_DE.dict (or AT/CH) into FilterDict.
func (r *GermanSpellerRule) InitFromDiscoveredResources() error {
	if r == nil {
		return nil
	}
	root := DiscoverGermanResourceDir()
	if root == "" {
		return nil
	}
	hun := filepath.Join(root, "hunspell")

	// Dict path by variant (Java default DE; AT/CH have their own .dict)
	dictName := "de_DE.dict"
	switch r.LanguageVariant {
	case "AT":
		dictName = "de_AT.dict"
	case "CH":
		dictName = "de_CH.dict"
	}
	dict := filepath.Join(hun, dictName)
	if st, err := os.Stat(dict); err == nil && st.Mode().IsRegular() {
		_ = WireGermanFilterSpeller(dict)
	} else if dictName != "de_DE.dict" {
		// fall back to de_DE if variant dict missing
		if p := filepath.Join(hun, "de_DE.dict"); fileExists(p) {
			_ = WireGermanFilterSpeller(p)
		}
	}

	// SpellingCheckRule.init order-ish: ignore, spelling, additional spelling, prohibit, additional prohibit
	if p := filepath.Join(hun, "ignore.txt"); fileExists(p) {
		_ = r.InitIgnoreFile(p)
	}
	if p := filepath.Join(hun, "spelling.txt"); fileExists(p) {
		_ = r.InitBaseSpellingIgnoreWords(p)
	}
	// getAdditionalSpellingFileNames: spelling_custom + spelling_global
	if p := filepath.Join(hun, "spelling_custom.txt"); fileExists(p) {
		_ = r.LoadIgnoreWordsFromFile(p)
	}
	if p := DiscoverSpellingGlobal(); p != "" {
		_ = r.LoadIgnoreWordsFromFile(p)
	}
	// CompoundAware paths also include multitoken-suggest + German spelling_recommendation
	if p := filepath.Join(root, "multitoken-suggest.txt"); fileExists(p) {
		_ = r.LoadIgnoreWordsFromFile(p)
	}
	if p := filepath.Join(hun, "spelling_recommendation.txt"); fileExists(p) {
		_ = r.LoadIgnoreWordsFromFile(p)
	}
	// Language-specific plain-text dict (AT/CH)
	switch r.LanguageVariant {
	case "AT":
		if p := filepath.Join(hun, "spelling-de-AT.txt"); fileExists(p) {
			_ = r.LoadIgnoreWordsFromFile(p)
		}
	case "CH":
		if p := filepath.Join(hun, "spelling-de-CH.txt"); fileExists(p) {
			_ = r.LoadIgnoreWordsFromFile(p)
		}
	}

	if p := filepath.Join(hun, "prohibit.txt"); fileExists(p) {
		_ = r.InitProhibitFile(p)
	}
	if p := filepath.Join(hun, "prohibit_custom.txt"); fileExists(p) {
		_ = r.LoadProhibitWordsFromFile(p)
	}

	// compound resource lists
	infix := filepath.Join(root, "words_infix_s.txt")
	stems := filepath.Join(root, "verb_stems.txt")
	vpref := filepath.Join(root, "verb_prefixes.txt")
	opref := filepath.Join(root, "other_prefixes.txt")
	if fileExists(infix) && fileExists(stems) && fileExists(vpref) && fileExists(opref) {
		_ = r.InitCompoundResourceFiles(infix, stems, vpref, opref)
	}
	if p := filepath.Join(root, "alt_neu.csv"); fileExists(p) {
		_ = r.InitOldSpellingFile(p)
	}
	// Compound tokenizers: common_words + hunspell .dic surfaces + LT exceptions/extended list
	r.wireCompoundTokenizers(root, hun, dictName)
	// TagPOS / LemmaOf from german.dict (optional) + added/removed + spelling /A /P
	r.wireTagPOSFromResources(root)
	// Synthesize when german_synth.dict is present (optional, fail-closed if missing)
	r.wireSynthesizeFromResources(root)
	return nil
}

// wireSynthesizeFromResources ports GermanSpellerRule synthesizer field when
// german_synth.dict exists (GermanSynthesizer with case/REMOVE/compound filters).
func (r *GermanSpellerRule) wireSynthesizeFromResources(resourceRoot string) {
	if r == nil || resourceRoot == "" || r.Synthesize != nil {
		return
	}
	// Ensure discovery uses same tree; openDiscovered* walks from cwd.
	_ = resourceRoot
	if openDiscoveredGermanSynthesizer() == nil && openDiscoveredGermanSynthBase() == nil {
		return
	}
	// GermanSpeller past-tense/participle: synthesize(token, posTagRE, true)
	r.Synthesize = synthesizeGermanRE
}

// wireTagPOSFromResources loads POS via german.dict (when present) + ManualTagger
// added/removed files — same CombiningTagger stack as Java BaseTagger/GermanTagger.
// Missing german.dict: manual-only (incomplete, not invented).
func (r *GermanSpellerRule) wireTagPOSFromResources(resourceRoot string) {
	if r == nil {
		return
	}
	// Do not overwrite test-injected hooks
	if r.TagPOS != nil || r.LemmaOf != nil {
		return
	}
	loadManual := func(name string) tagging.WordTagger {
		if resourceRoot == "" {
			return nil
		}
		p := filepath.Join(resourceRoot, name)
		if !fileExists(p) {
			return nil
		}
		f, err := os.Open(p)
		if err != nil {
			return nil
		}
		defer f.Close()
		mt, err := tagging.NewManualTagger(f)
		if err != nil || mt == nil {
			return nil
		}
		return mt
	}
	// Binary POS dict (optional)
	var morfo tagging.WordTagger
	if p := DiscoverGermanPOSDict(); p != "" {
		if mt := tagging.OpenMorfologikTagger(p); mt != nil {
			morfo = mt
		}
	} else if resourceRoot != "" {
		if p := filepath.Join(resourceRoot, "german.dict"); fileExists(p) {
			if mt := tagging.OpenMorfologikTagger(p); mt != nil {
				morfo = mt
			}
		}
	}
	// Manual additions (tagger2 in CombiningTagger)
	var manuals tagging.WordTagger
	added := loadManual("added.txt")
	addedCustom := loadManual("added_custom.txt")
	switch {
	case added != nil && addedCustom != nil:
		manuals = tagging.NewCombiningTagger(added, addedCustom, false)
	case added != nil:
		manuals = added
	case addedCustom != nil:
		manuals = addedCustom
	}
	// spelling.txt expansions: /A /P adj + underscore verb nominalized/zu
	var adjExp, verbExp tagging.WordTagger
	if resourceRoot != "" {
		spellingPath := filepath.Join(resourceRoot, "hunspell", "spelling.txt")
		if fileExists(spellingPath) {
			if ex, err := detag.LoadSpellingAdjExpansionFromFile(spellingPath); err == nil && ex != nil {
				adjExp = ex
			}
			if vex, err := detag.LoadSpellingVerbExpansionFromFile(spellingPath); err == nil && vex != nil {
				verbExp = vex
			}
		}
	}
	if morfo == nil && manuals == nil && adjExp == nil && verbExp == nil {
		return
	}
	// merge expansions into manuals side (additional surface readings)
	merge := func(extra tagging.WordTagger) {
		if extra == nil {
			return
		}
		if manuals != nil {
			manuals = tagging.NewCombiningTagger(manuals, extra, false)
		} else {
			manuals = extra
		}
	}
	merge(adjExp)
	merge(verbExp)
	removal := loadManual("removed.txt")
	if remCustom := loadManual("removed_custom.txt"); remCustom != nil {
		if removal != nil {
			removal = tagging.NewCombiningTagger(removal, remCustom, false)
		} else {
			removal = remCustom
		}
	}
	// Java BaseTagger CombiningTagger: tagger1=morfo dict, tagger2=manual additions,
	// removalTagger filters. Tag order: manual first, then dict, then remove.
	tagger1 := tagging.WordTagger(tagging.MapWordTagger{})
	if morfo != nil {
		tagger1 = morfo
	}
	tagger2 := tagging.WordTagger(tagging.MapWordTagger{})
	if manuals != nil {
		tagger2 = manuals
	}
	var wt tagging.WordTagger
	if removal != nil {
		wt = tagging.NewCombiningTaggerWithRemoval(tagger1, tagger2, removal, false)
	} else {
		wt = tagging.NewCombiningTagger(tagger1, tagger2, false)
	}
	r.TagPOS = func(word string) []string {
		if word == "" {
			return nil
		}
		tags := wt.Tag(word)
		if len(tags) == 0 && !startsWithUppercase(word) {
			tags = wt.Tag(uppercaseFirstChar(word))
		}
		if len(tags) == 0 {
			return nil
		}
		out := make([]string, 0, len(tags))
		for _, tw := range tags {
			if tw.PosTag != "" {
				out = append(out, tw.PosTag)
			}
		}
		return out
	}
	r.LemmaOf = func(word string) string {
		if word == "" {
			return ""
		}
		// Java findLemmaForNoun: tag(uppercaseFirstChar(word)), first SUB lemma
		probe := word
		tags := wt.Tag(probe)
		if len(tags) == 0 {
			probe = uppercaseFirstChar(word)
			tags = wt.Tag(probe)
		}
		for _, tw := range tags {
			if strings.HasPrefix(tw.PosTag, "SUB") && tw.Lemma != "" {
				return tw.Lemma
			}
		}
		return ""
	}
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.Mode().IsRegular()
}

// wireCompoundTokenizers builds strict + non-strict GermanCompoundTokenizer from:
// common_words.txt, hunspell .dic surfaces, and Java LT exceptions/extendedList
// (ApplyLanguageToolExtras inside NewGermanCompoundTokenizer).
func (r *GermanSpellerRule) wireCompoundTokenizers(resourceRoot, hunDir, dictName string) {
	if r == nil {
		return
	}
	strict := compoundtok.NewGermanCompoundTokenizer(true)
	nonStrict := compoundtok.NewGermanCompoundTokenizer(false)
	loadList := func(path string) {
		if !fileExists(path) {
			return
		}
		words, err := LoadSpellingWordList(path)
		if err != nil {
			return
		}
		for _, w := range words {
			strict.AddWord(w)
			nonStrict.AddWord(w)
		}
	}
	loadList(filepath.Join(resourceRoot, "common_words.txt"))
	// hunspell dic surfaces as part lexicon (real LT resource, not invent)
	dicPath := filepath.Join(hunDir, strings.TrimSuffix(dictName, ".dict")+".dic")
	if !fileExists(dicPath) {
		dicPath = filepath.Join(hunDir, "de_DE.dic")
	}
	if fileExists(dicPath) {
		if f, err := os.Open(dicPath); err == nil {
			_ = strict.LoadHunspellDic(f)
			f.Close()
		}
		if f, err := os.Open(dicPath); err == nil {
			_ = nonStrict.LoadHunspellDic(f)
			f.Close()
		}
	}
	r.CompoundTokenize = func(word string) []string {
		return strict.Tokenize(word)
	}
	r.CompoundTokenizeNonStrict = func(word string) []string {
		return nonStrict.Tokenize(word)
	}
	// getAllSplits stand-in: union of strict + non-strict dictionary partitions
	r.CompoundTokenizeAll = func(word string) [][]string {
		seen := map[string]struct{}{}
		var out [][]string
		add := func(splits [][]string) {
			for _, parts := range splits {
				key := strings.Join(parts, "\x00")
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				out = append(out, parts)
			}
		}
		add(strict.AllSplits(word))
		add(nonStrict.AllSplits(word))
		return out
	}
}

// wireCompoundTokenizersFromFile is retained for tests that load a single word list.
func (r *GermanSpellerRule) wireCompoundTokenizersFromFile(path string) {
	if r == nil {
		return
	}
	strict := compoundtok.NewGermanCompoundTokenizer(true)
	nonStrict := compoundtok.NewGermanCompoundTokenizer(false)
	words, err := LoadSpellingWordList(path)
	if err == nil {
		for _, w := range words {
			strict.AddWord(w)
			nonStrict.AddWord(w)
		}
	}
	r.CompoundTokenize = func(word string) []string {
		return strict.Tokenize(word)
	}
	r.CompoundTokenizeNonStrict = func(word string) []string {
		return nonStrict.Tokenize(word)
	}
	r.CompoundTokenizeAll = func(word string) [][]string {
		seen := map[string]struct{}{}
		var out [][]string
		for _, parts := range append(strict.AllSplits(word), nonStrict.AllSplits(word)...) {
			key := strings.Join(parts, "\x00")
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, parts)
		}
		return out
	}
}
