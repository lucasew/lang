package languagetool

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	catok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ca"
	estok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/es"
	frtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/fr"
	pttok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/pt"
)

// RegisterBinaryPOSTagger installs lt.TagWord from a Morfologik POS dictionary
// (CFSA2 or FSA5), matching Java BaseTagger:
//   CombiningTagger(MorfologikTagger, ManualTagger(added*), ManualTagger(removed*), overwrite=false)
// plus BaseTagger.getAnalyzedTokens case-merge.
// Returns false if the dictionary cannot be opened.
//
// Also wires language word-tokenizer IsTagged* hooks (Java *Tagger.INSTANCE used
// by *WordTokenizer.wordsToAdd) for FR/ES/PT/CA when those modules are present.
func RegisterBinaryPOSTagger(lt *JLanguageTool, dictPath string) bool {
	if lt == nil || dictPath == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	var wordTagger tagging.WordTagger = morfologikPOSWordTagger{d: d}
	// Java BaseTagger.initWordTagger: only wrap when manual additions stream exists.
	if manual := loadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); manual != nil {
		removal := loadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"})
		wordTagger = tagging.NewCombiningTaggerWithRemoval(wordTagger, manual, removal, false)
	}
	langBase := languageBaseFromPath(dictPath, lt.GetLanguageCode())
	var tw func(token string) []TokenTag
	if langBase == "pl" {
		// Java PolishTagger.tag (exact WordTagger lookups + case merge).
		// Inline to avoid import cycle: languagetool → tagging/pl → languagetool.
		tw = polishTaggerCaseTagWord(wordTagger)
	} else {
		// Java BaseTagger: tagLowercaseWithUppercase=true by default (most language taggers).
		base := tagging.NewBaseTagger(wordTagger, dictPath, langBase, true)
		tw = baseTaggerToTagWord(base)
	}
	lt.TagWord = tw
	wireTokenizerIsTaggedFromPOS(lt.GetLanguageCode(), tw)
	return true
}

// polishTaggerCaseTagWord ports Java PolishTagger.tag case logic for TagWord inject:
// surface exact, then if non-lower also lower exact, then if both empty and surface
// is lower try UppercaseFirstChar. Always merges lower for non-lower (incl. mixed case).
func polishTaggerCaseTagWord(wt tagging.WordTagger) func(token string) []TokenTag {
	if wt == nil {
		return nil
	}
	lookup := func(w string) []TokenTag {
		tws := wt.Tag(w)
		if len(tws) == 0 {
			return nil
		}
		out := make([]TokenTag, 0, len(tws))
		for _, tw := range tws {
			out = append(out, TokenTag{POS: tw.PosTag, Lemma: tw.Lemma})
		}
		return out
	}
	return func(token string) []TokenTag {
		if token == "" {
			return nil
		}
		// Java: typewriter apostrophe normalisation in Polish ports often use ’ → '
		word := strings.ReplaceAll(token, "’", "'")
		low := strings.ToLower(word)
		var out []TokenTag
		seen := map[string]struct{}{}
		add := func(tags []TokenTag) {
			for _, t := range tags {
				key := t.POS + "\x00" + t.Lemma
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				out = append(out, t)
			}
		}
		add(lookup(word))
		if word != low {
			add(lookup(low))
		}
		if len(out) == 0 && word == low {
			title := tools.UppercaseFirstChar(word)
			if title != word {
				add(lookup(title))
			}
		}
		return out
	}
}

func languageBaseFromPath(dictPath, langCode string) string {
	base := langCode
	if i := strings.IndexByte(langCode, '-'); i > 0 {
		base = langCode[:i]
	}
	base = strings.ToLower(base)
	if base != "" {
		return base
	}
	// Fallback: …/resource/{code}/….dict
	parts := strings.Split(filepath.ToSlash(dictPath), "/")
	for i, p := range parts {
		if p == "resource" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "xx"
}

func baseTaggerToTagWord(bt *tagging.BaseTagger) func(token string) []TokenTag {
	if bt == nil {
		return nil
	}
	return func(token string) []TokenTag {
		if token == "" {
			return nil
		}
		tws := bt.TagWord(token)
		if len(tws) == 0 {
			return nil
		}
		out := make([]TokenTag, 0, len(tws))
		seen := map[string]struct{}{}
		for _, tw := range tws {
			key := tw.PosTag + "\x00" + tw.Lemma
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, TokenTag{POS: tw.PosTag, Lemma: tw.Lemma})
		}
		return out
	}
}

// wireTokenizerIsTaggedFromPOS ports Java *WordTokenizer → *Tagger.INSTANCE.isTagged.
func wireTokenizerIsTaggedFromPOS(langCode string, tw func(token string) []TokenTag) {
	if tw == nil {
		return
	}
	isTagged := func(s string) bool {
		for _, t := range tw(s) {
			if t.POS != "" {
				return true
			}
		}
		return false
	}
	base := langCode
	if i := strings.IndexByte(langCode, '-'); i > 0 {
		base = langCode[:i]
	}
	switch strings.ToLower(base) {
	case "fr":
		frtok.IsTaggedFR = isTagged
	case "es":
		estok.IsTaggedES = isTagged
	case "pt":
		pttok.IsTaggedPT = isTagged
	case "ca":
		catok.IsTaggedCA = isTagged
	}
}

// morfologikPOSWordTagger is MorfologikTagger + multi-reading '+' split for
// Morfeusz-style tags (subst:…+adj:…). Italian-style VER:part+past stays whole.
type morfologikPOSWordTagger struct {
	d *atticmorfo.Dictionary
}

func (w morfologikPOSWordTagger) Tag(word string) []tagging.TaggedWord {
	if w.d == nil || word == "" {
		return nil
	}
	forms, err := w.d.Lookup(word)
	if err != nil || len(forms) == 0 {
		return nil
	}
	out := make([]tagging.TaggedWord, 0, len(forms)*2)
	for _, f := range forms {
		if f.Tag == "" || !strings.Contains(f.Tag, "+") {
			out = append(out, tagging.NewTaggedWord(f.Stem, f.Tag))
			continue
		}
		parts := strings.Split(f.Tag, "+")
		splitMulti := len(parts) > 1
		if splitMulti {
			for _, part := range parts {
				if !strings.Contains(part, ":") {
					splitMulti = false
					break
				}
			}
		}
		if !splitMulti {
			out = append(out, tagging.NewTaggedWord(f.Stem, f.Tag))
			continue
		}
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			out = append(out, tagging.NewTaggedWord(f.Stem, part))
		}
	}
	return out
}

// LoadManualTaggerBesideDict loads Java BaseTagger manual files next to the POS
// dict (or one/two parents up for nested layouts like sr/dictionary/ekavian/).
// Concatenates all present names from the first resource root that has any file.
// Exported for language-specific tagger wiring (e.g. EnglishTagger).
func LoadManualTaggerBesideDict(dictPath string, names []string) tagging.WordTagger {
	return loadManualTaggerBesideDict(dictPath, names)
}

// LoadManualTaggerFromDirs tries each resource directory in order; returns the
// first non-nil ManualTagger built from names present in that directory.
func LoadManualTaggerFromDirs(dirs []string, names []string) tagging.WordTagger {
	for _, dir := range dirs {
		if dir == "" {
			continue
		}
		var paths []string
		for _, name := range names {
			p := filepath.Join(dir, name)
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				paths = append(paths, p)
			}
		}
		if len(paths) > 0 {
			return openManualTaggerConcat(paths)
		}
	}
	return nil
}

func loadManualTaggerBesideDict(dictPath string, names []string) tagging.WordTagger {
	if dictPath == "" || len(names) == 0 {
		return nil
	}
	dir := filepath.Dir(dictPath)
	for depth := 0; depth < 4; depth++ {
		var paths []string
		for _, name := range names {
			p := filepath.Join(dir, name)
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				paths = append(paths, p)
			}
		}
		if len(paths) > 0 {
			return openManualTaggerConcat(paths)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return nil
}

func openManualTaggerConcat(paths []string) tagging.WordTagger {
	var readers []io.Reader
	var files []*os.File
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			continue
		}
		files = append(files, f)
		readers = append(readers, f)
	}
	if len(readers) == 0 {
		return nil
	}
	mt, err := tagging.NewManualTagger(io.MultiReader(readers...))
	for _, f := range files {
		_ = f.Close()
	}
	if err != nil || mt == nil {
		return nil
	}
	return mt
}

// BinaryPOSTagWord returns a TagWord inject from an opened POS dictionary only
// (no manual added/removed). Prefer RegisterBinaryPOSTagger for engine wiring.
// Case logic follows Java BaseTagger (via TagWord on a plain morfologik tagger).
func BinaryPOSTagWord(d *atticmorfo.Dictionary) func(token string) []TokenTag {
	if d == nil {
		return nil
	}
	bt := tagging.NewBaseTagger(morfologikPOSWordTagger{d: d}, "", "xx", true)
	return baseTaggerToTagWord(bt)
}
