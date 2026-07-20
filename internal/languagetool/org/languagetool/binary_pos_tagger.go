package languagetool

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	catok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/ca"
	estok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/es"
	frtok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/fr"
	pttok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/pt"
)

var (
	glAdjPartFSTagWord        = regexp.MustCompile(`^V.P..SF.|A[QO].[FC][SN].$`)
	glVerbTagWord             = regexp.MustCompile(`^V.+`)
	glPrefixesForVerbsTagWord = regexp.MustCompile(`(?i)^(auto|re)(...+)$`)
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
	switch langBase {
	case "pl":
		// Java PolishTagger.tag (exact WordTagger lookups + case merge).
		// Inline to avoid import cycle: languagetool → tagging/pl → languagetool.
		tw = polishTaggerCaseTagWord(wordTagger)
	case "ru":
		// Java RussianTagger: accent strip then BaseTagger.getAnalyzedTokens.
		// MayMissingYO is chunk-level (full Tagger.tag); TagWord inject is POS only.
		tw = russianTaggerTagWord(wordTagger, dictPath)
	case "gl":
		// Java GalicianTagger: exact lookups + mente/auto|re prefixes (not BaseTagger alone).
		tw = galicianTaggerTagWord(wordTagger)
	case "pt":
		// Java PortugueseTagger: exact lookups + ordinals/mente/soto- (not BaseTagger alone).
		tw = portugueseTaggerTagWord(wordTagger)
	case "es":
		// Java SpanishTagger: exact + all-upper title + mente/auto prefixes.
		tw = spanishTaggerTagWord(wordTagger)
	case "fr":
		// Java FrenchTagger: capitalized/all-upper/hyphen-title + oe + prefixes.
		tw = frenchTaggerTagWord(wordTagger)
	case "ca":
		// Java CatalanTagger: all-upper + ment/auto/ela + Valencian POS filter (non-val strips 0*).
		tw = catalanTaggerTagWord(wordTagger, false)
	default:
		// Java BaseTagger: tagLowercaseWithUppercase=true by default (most language taggers).
		base := tagging.NewBaseTagger(wordTagger, dictPath, langBase, true)
		tw = baseTaggerToTagWord(base)
	}
	lt.TagWord = tw
	wireTokenizerIsTaggedFromPOS(lt.GetLanguageCode(), tw)
	return true
}

// spanishTaggerTagWord ports Java SpanishTagger.tag for TagWord inject.
func spanishTaggerTagWord(wt tagging.WordTagger) func(token string) []TokenTag {
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
	adjPartFS := regexp.MustCompile(`^VMP00SF|A[QO].[FC]S.$`)
	verbRE := regexp.MustCompile(`^V.+`)
	prefVerb := regexp.MustCompile(`(?i)^(auto)([^r].{3,})$`)
	prefVerb2 := regexp.MustCompile(`(?i)^(autor)(r.{3,})$`)
	prefAdjSuper := regexp.MustCompile(`(?i)^(super)(.*[aeiouáéèíòóïü].+[aeiouáéèíòóïü].*)$`)
	prefAdjHyphen := regexp.MustCompile(`(?i)^(.+)-(.+)$`)
	adjRE := regexp.MustCompile(`^AQ.+`)
	adjMS := regexp.MustCompile(`^AQ.MS.|AQ.CS.|AQ.MN.$`)
	noPrefAdj := regexp.MustCompile(`(?i)^(anti|pre|ex|pro|afro|ultra|super|súper)$`)
	adjVP := regexp.MustCompile(`^AQ.*|V.P.*$`)

	return func(token string) []TokenTag {
		if token == "" {
			return nil
		}
		w := strings.ReplaceAll(token, "’", "'")
		lower := strings.ToLower(w)
		isLower := w == lower
		isMixed := tools.IsMixedCase(w)
		isAllUpper := tools.IsAllUppercase(w)
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
		add(lookup(w))
		if !isLower && !isMixed {
			add(lookup(lower))
		}
		if isAllUpper {
			add(lookup(tools.UppercaseFirstChar(lower)))
		}
		if len(out) == 0 && !isMixed {
			if strings.HasSuffix(lower, "mente") {
				possibleAdj := strings.TrimSuffix(lower, "mente")
				for _, tw := range lookup(possibleAdj) {
					if tw.POS != "" && adjPartFS.MatchString(tw.POS) {
						add([]TokenTag{{POS: "RG", Lemma: lower}})
						break
					}
				}
			}
			if len(out) == 0 {
				if m := prefVerb.FindStringSubmatch(w); m != nil {
					for _, tw := range lookup(strings.ToLower(m[2])) {
						if tw.POS != "" && verbRE.MatchString(tw.POS) {
							add([]TokenTag{{POS: tw.POS, Lemma: strings.ToLower(m[1]) + tw.Lemma}})
						}
					}
				}
			}
			if len(out) == 0 {
				if m := prefVerb2.FindStringSubmatch(w); m != nil {
					for _, tw := range lookup(strings.ToLower(m[2])) {
						if tw.POS != "" && verbRE.MatchString(tw.POS) {
							add([]TokenTag{{POS: tw.POS, Lemma: strings.ToLower(m[1]) + tw.Lemma}})
						}
					}
				}
			}
			if len(out) == 0 {
				if m := prefAdjSuper.FindStringSubmatch(w); m != nil {
					for _, tw := range lookup(strings.ToLower(m[2])) {
						if tw.POS != "" && adjVP.MatchString(tw.POS) {
							add([]TokenTag{{POS: tw.POS, Lemma: strings.ToLower(m[1]) + tw.Lemma}})
						}
					}
				}
			}
			if len(out) == 0 {
				if m := prefAdjHyphen.FindStringSubmatch(w); m != nil {
					pref := strings.ToLower(m[1])
					if !noPrefAdj.MatchString(pref) {
						adj := strings.ToLower(m[2])
						prefixOK := false
						for _, tw := range lookup(pref) {
							if tw.POS != "" && adjMS.MatchString(tw.POS) {
								prefixOK = true
								break
							}
						}
						if prefixOK {
							for _, tw := range lookup(adj) {
								if tw.POS != "" && adjRE.MatchString(tw.POS) {
									add([]TokenTag{{POS: tw.POS, Lemma: pref + "-" + tw.Lemma}})
									break
								}
							}
						}
					}
				}
			}
		}
		if len(out) == 0 && tools.IsEmoji(token) {
			add([]TokenTag{{POS: "_emoji_", Lemma: "_emoji_"}})
		}
		return out
	}
}

// frenchTaggerTagWord ports Java FrenchTagger.tagWord (+ oe fallback) for TagWord inject.
func frenchTaggerTagWord(wt tagging.WordTagger) func(token string) []TokenTag {
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
	verbRE := regexp.MustCompile(`^V .+$`)
	prefVerbs := regexp.MustCompile(`(?i)^(auto|auto-|re-|sur-)([^-].*[aeiouêàéèíòóïü].+[aeiouêàéèíòóïü].*)$`)
	nounAdj := regexp.MustCompile(`^[NJ] .+|V ppa.*$`)
	prefNounAdjHyphen := regexp.MustCompile(`(?i)^(post-|sur-|mini-|méga-|demi-|péri-|anti-|géo-|nord-|sud-|néo-|méga-|ultra-|pro-|inter-|micro-|macro-|sous-|haut-|auto-|ré-|pré-|super-|vice-|hyper-|proto-|grand-|pseudo-)(.+)$`)
	prefNounAdj := regexp.MustCompile(`(?i)^(mini|méga)([^-].*[aeiouêàéèíòóïü].+[aeiouêàéèíòóïü].*)$`)
	ambig := map[string]struct{}{
		"-Le": {}, "-Les": {}, "-La": {}, "-Elle": {}, "-Elles": {}, "-On": {},
		"-Tu": {}, "-Vous": {}, "-Il": {}, "-Ils": {}, "-Ce": {},
	}

	return func(token string) []TokenTag {
		if token == "" {
			return nil
		}
		w := token
		if len(w) > 1 && strings.Contains(w, "’") {
			w = strings.ReplaceAll(w, "’", "'")
		}
		tagOne := func(word, originalWord string) []TokenTag {
			lower := strings.ToLower(word)
			isStartUpper := tools.IsCapitalizedWord(word)
			isAllUpper := tools.IsAllUppercase(word)
			_, isAmbig := ambig[originalWord]
			isHyphenTitle := !isAmbig && strings.Contains(originalWord, "-") &&
				originalWord == tools.ConvertToTitleCaseIteratingChars(lower)
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
			if isAllUpper || isStartUpper || isHyphenTitle {
				add(lookup(lower))
			}
			if len(out) == 0 && isAllUpper {
				add(lookup(tools.ConvertToTitleCaseIteratingChars(lower)))
			}
			if len(out) == 0 {
				if m := prefVerbs.FindStringSubmatch(word); m != nil {
					for _, tw := range lookup(strings.ToLower(m[2])) {
						if tw.POS != "" && verbRE.MatchString(tw.POS) {
							add([]TokenTag{{POS: tw.POS, Lemma: strings.ToLower(m[1]) + tw.Lemma}})
						}
					}
				}
			}
			if len(out) == 0 {
				if m := prefNounAdj.FindStringSubmatch(word); m != nil {
					for _, tw := range lookup(strings.ToLower(m[2])) {
						if tw.POS != "" && nounAdj.MatchString(tw.POS) {
							add([]TokenTag{{POS: tw.POS, Lemma: strings.ToLower(m[1]) + tw.Lemma}})
						}
					}
				}
			}
			if len(out) == 0 {
				if m := prefNounAdjHyphen.FindStringSubmatch(word); m != nil {
					possible := strings.ToLower(m[2])
					for _, tw := range lookup(possible) {
						if tw.POS != "" && nounAdj.MatchString(tw.POS) {
							add([]TokenTag{{POS: tw.POS, Lemma: strings.ToLower(m[1]) + tw.Lemma}})
						}
					}
				}
			}
			return out
		}
		out := tagOne(w, w)
		if len(out) == 0 && strings.Contains(strings.ToLower(w), "oe") {
			alt := strings.ReplaceAll(strings.ReplaceAll(w, "oe", "œ"), "OE", "Œ")
			out = tagOne(alt, w)
		}
		if len(out) == 0 && tools.IsEmoji(token) {
			return []TokenTag{{POS: "_emoji_", Lemma: "_emoji_"}}
		}
		return out
	}
}

// catalanTaggerTagWord ports Java CatalanTagger.tag case + additionalTags for TagWord inject.
// isValencian: strip leading '0' from POS (val) vs drop 0* tags (central).
func catalanTaggerTagWord(wt tagging.WordTagger, isValencian bool) func(token string) []TokenTag {
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
	adjPartFS := regexp.MustCompile(`^VMP00SF.|A[QO].[FC]S.$`)
	verbRE := regexp.MustCompile(`^V.+$`)
	prefVerbs := regexp.MustCompile(`(?i)^(auto)(.*[aeiouàéèíòóïü].+[aeiouàéèíòóïü].*)$`)
	adjCompost := regexp.MustCompile(`(?i)^(.*)o-(.*.*)$`)
	tresAdj := regexp.MustCompile(`(?i)^(.*)o-(.*)o-(.*.*)$`)
	altresPref := map[string]struct{}{
		"greco": {}, "sino": {}, "italo": {}, "franco": {}, "gal·lo": {}, "luso": {},
		"germano": {}, "hispano": {}, "anglo": {}, "àrabo": {}, "austro": {}, "belgo": {},
	}
	noAltresPref := map[string]struct{}{
		"grego": {}, "xineso": {}, "italiano": {}, "franceso": {},
		"portugueso": {}, "angleso": {}, "espanyolo": {}, "alemanyo": {}, "arabo": {}, "austríaco": {}, "bèlgico": {},
	}
	allUpperExc := map[string]struct{}{"ARNAU": {}, "CRISTIAN": {}, "TOMÀS": {}}
	wordformHasPostag := func(wordform, postag string) bool {
		for _, tw := range lookup(wordform) {
			if tw.POS == postag {
				return true
			}
		}
		return false
	}
	isValidAdjForm := func(stem string) bool {
		if _, bad := noAltresPref[stem+"o"]; bad {
			return false
		}
		if wordformHasPostag(stem+"a", "AQ0FS0") {
			return true
		}
		_, ok := altresPref[stem+"o"]
		return ok
	}
	filterPOS := func(tags []TokenTag) []TokenTag {
		if len(tags) == 0 {
			return tags
		}
		var out []TokenTag
		for _, t := range tags {
			if t.POS != "" && t.POS[0] == '0' {
				if isValencian {
					out = append(out, TokenTag{POS: t.POS[1:], Lemma: t.Lemma})
				}
				// non-valencian: drop 0* tags
				continue
			}
			out = append(out, t)
		}
		return out
	}

	return func(token string) []TokenTag {
		if token == "" {
			return nil
		}
		w := token
		if len(w) > 1 && strings.Contains(w, "’") {
			w = strings.ReplaceAll(w, "’", "'")
		}
		w = tools.NormalizeNFC(w)
		lower := strings.ToLower(w)
		isLower := w == lower
		isMixed := tools.IsMixedCase(w)
		isAllUpper := tools.IsAllUppercase(w)
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
		add(lookup(w))
		if !isLower && !isMixed {
			add(lookup(lower))
		}
		_, exc := allUpperExc[w]
		if (len(out) == 0 || exc) && isAllUpper {
			add(lookup(tools.UppercaseFirstChar(lower)))
		}
		if len(out) == 0 && !isMixed {
			if strings.HasSuffix(lower, "ment") {
				possibleAdj := strings.TrimSuffix(lower, "ment")
				for _, tw := range lookup(possibleAdj) {
					if tw.POS != "" && adjPartFS.MatchString(tw.POS) {
						add([]TokenTag{{POS: "RG", Lemma: lower}})
						break
					}
				}
			}
			if len(out) == 0 {
				if m := prefVerbs.FindStringSubmatch(w); m != nil {
					possibleVerb := tools.NormalizeNFC(strings.ToLower(m[2]))
					for _, tw := range lookup(possibleVerb) {
						if tw.Lemma == "nòmer" {
							continue
						}
						if tw.POS != "" && verbRE.MatchString(tw.POS) {
							add([]TokenTag{{POS: tw.POS, Lemma: strings.ToLower(m[1]) + tw.Lemma}})
						}
					}
				}
			}
			if len(out) == 0 {
				if m := adjCompost.FindStringSubmatch(w); m != nil {
					adj1 := strings.ToLower(m[1])
					if isValidAdjForm(adj1) {
						adj2 := strings.ToLower(m[2])
						for _, tw := range lookup(adj2) {
							if tw.POS != "" && strings.HasPrefix(tw.POS, "A") {
								add([]TokenTag{{POS: tw.POS, Lemma: adj1 + "o-" + tw.Lemma}})
								break
							}
						}
					}
				}
			}
			if len(out) == 0 {
				if m := tresAdj.FindStringSubmatch(w); m != nil {
					adj1, adj2 := strings.ToLower(m[1]), strings.ToLower(m[2])
					if isValidAdjForm(adj1) && isValidAdjForm(adj2) {
						adj3 := strings.ToLower(m[3])
						for _, tw := range lookup(adj3) {
							if tw.POS != "" && strings.HasPrefix(tw.POS, "A") {
								add([]TokenTag{{POS: tw.POS, Lemma: adj1 + "o-" + adj2 + "o-" + tw.Lemma}})
								break
							}
						}
					}
				}
			}
			if len(out) == 0 && (strings.Contains(w, "\u0140") || strings.Contains(w, "\u013f")) {
				possible := strings.ReplaceAll(lower, "\u0140", "l·")
				add(lookup(possible))
			}
			if len(out) == 0 && isValencian && strings.HasSuffix(lower, "iste") {
				possible := strings.TrimSuffix(lower, "iste") + "ista"
				for _, tw := range lookup(possible) {
					switch tw.POS {
					case "NCCS000":
						add([]TokenTag{{POS: "NCMS000", Lemma: possible}})
					case "AQ0CS0":
						add([]TokenTag{{POS: "AQ0MS0", Lemma: possible}})
					}
				}
			}
		}
		out = filterPOS(out)
		if len(out) == 0 && tools.IsEmoji(token) {
			return []TokenTag{{POS: "_emoji_", Lemma: "_emoji_"}}
		}
		return out
	}
}

// portugueseTaggerTagWord ports Java PortugueseTagger.tag for TagWord inject
// (exact case merge + number expressions + mente + soto- prefixes).
func portugueseTaggerTagWord(wt tagging.WordTagger) func(token string) []TokenTag {
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
	// Patterns aligned with PortugueseTagger.java / tagging/pt.
	adjPartFS := regexp.MustCompile(`^V.P..SF.|A[QO].[FC][SN].$`)
	verbRE := regexp.MustCompile(`^V.+`)
	prefixes := regexp.MustCompile(`(?i)^(soto-)(...+)$`)
	const (
		ordMasc = "oºᵒ"
		ordFem  = "aªᵃ"
		ordPl   = "sˢ"
	)
	ordSuf := "[" + ordMasc + ordFem + "][" + ordPl + "]?"
	ordPat := regexp.MustCompile(`^\d+[\d,.]*\.?` + ordSuf + `$`)
	ordMascSg := regexp.MustCompile("[" + ordMasc + "]$")
	ordFemSg := regexp.MustCompile("[" + ordFem + "]$")
	ordMascPl := regexp.MustCompile("[" + ordMasc + "][" + ordPl + "]$")
	ordFemPl := regexp.MustCompile("[" + ordFem + "][" + ordPl + "]$")
	ordReplace := regexp.MustCompile(ordSuf)
	percentPat := regexp.MustCompile(`^−?\d+[\d,.]*%$`)
	degreePat := regexp.MustCompile(`^−?\d+[\d,.]*°$`)

	hyphenMixed := func(word string) bool {
		if strings.Contains(word, "-") {
			for _, part := range strings.Split(word, "-") {
				if tools.IsMixedCase(part) {
					return true
				}
			}
			return false
		}
		return tools.IsMixedCase(word)
	}

	return func(token string) []TokenTag {
		if token == "" {
			return nil
		}
		w := strings.ReplaceAll(token, "’", "'")
		lower := strings.ToLower(w)
		isLower := w == lower
		isMixed := hyphenMixed(w)
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
		add(lookup(w))
		if !isLower && !isMixed {
			add(lookup(lower))
		}
		if len(out) == 0 {
			if ordPat.MatchString(w) {
				lemma := ordReplace.ReplaceAllString(w, "º")
				ng := ""
				switch {
				case ordMascPl.MatchString(w):
					ng = "MP"
				case ordFemPl.MatchString(w):
					ng = "FP"
				case ordMascSg.MatchString(w):
					ng = "MS"
				case ordFemSg.MatchString(w):
					ng = "FS"
				}
				if ng != "" {
					add([]TokenTag{
						{POS: "NC" + ng + "000", Lemma: lemma},
						{POS: "AO0" + ng + "0", Lemma: lemma},
					})
				}
			} else if percentPat.MatchString(w) || degreePat.MatchString(w) {
				add([]TokenTag{{POS: "NCMP000", Lemma: w}})
			}
		}
		if len(out) == 0 && !isMixed && strings.HasSuffix(lower, "mente") {
			possibleAdj := strings.TrimSuffix(lower, "mente")
			for _, tw := range lookup(possibleAdj) {
				if tw.POS != "" && adjPartFS.MatchString(tw.POS) {
					add([]TokenTag{{POS: "RG", Lemma: lower}})
					break
				}
			}
		}
		if len(out) == 0 && !isMixed {
			if m := prefixes.FindStringSubmatch(w); m != nil {
				pref := strings.ToLower(m[1])
				verb := strings.ToLower(m[2])
				for _, tw := range lookup(verb) {
					if tw.POS == "" || !verbRE.MatchString(tw.POS) {
						continue
					}
					lemma := pref + tw.Lemma
					if len(lookup(lemma)) > 0 {
						continue
					}
					add([]TokenTag{{POS: tw.POS, Lemma: lemma}})
				}
			}
		}
		return out
	}
}

// galicianTaggerTagWord ports Java GalicianTagger.tag case + additionalTags for TagWord inject.
func galicianTaggerTagWord(wt tagging.WordTagger) func(token string) []TokenTag {
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
		w := strings.ReplaceAll(token, "’", "'")
		lower := strings.ToLower(w)
		isLower := w == lower
		isMixed := tools.IsMixedCase(w)
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
		add(lookup(w))
		if !isLower && !isMixed {
			add(lookup(lower))
		}
		if len(out) == 0 && !isMixed {
			// -mente adverbs of manner (RM)
			if strings.HasSuffix(lower, "mente") {
				possibleAdj := strings.TrimSuffix(lower, "mente")
				for _, tw := range lookup(possibleAdj) {
					if tw.POS != "" && glAdjPartFSTagWord.MatchString(tw.POS) {
						add([]TokenTag{{POS: "RM", Lemma: lower}})
						break
					}
				}
			}
			// auto|re + verb
			if m := glPrefixesForVerbsTagWord.FindStringSubmatch(w); m != nil && len(out) == 0 {
				pref := strings.ToLower(m[1])
				possibleVerb := strings.ToLower(m[2])
				for _, tw := range lookup(possibleVerb) {
					if tw.POS != "" && glVerbTagWord.MatchString(tw.POS) {
						add([]TokenTag{{POS: tw.POS, Lemma: pref + tw.Lemma}})
					}
				}
			}
		}
		return out
	}
}

// russianTaggerTagWord ports Java RussianTagger.tag for TagWord inject:
// NormalizeRussianSurface then BaseTagger case-merge on the normalized form.
func russianTaggerTagWord(wt tagging.WordTagger, dictPath string) func(token string) []TokenTag {
	if wt == nil {
		return nil
	}
	base := tagging.NewBaseTagger(wt, dictPath, "ru", true)
	inner := baseTaggerToTagWord(base)
	return func(token string) []TokenTag {
		if token == "" {
			return nil
		}
		norm, _ := tagging.NormalizeRussianSurface(token)
		return inner(norm)
	}
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
