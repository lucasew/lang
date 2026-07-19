package de

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

const GermanDictPath = "/de/german.dict"

// GermanTagger ports org.languagetool.tagging.de.GermanTagger.
// Binary german.dict + manual added/removed via WordTagger; optional
// SpellingAdjExpansion (/A /P), SpellingVerbExpansion (prefix_verb), compound split.
type GermanTagger struct {
	*tagging.BaseTagger
	// RemovalTagger ports GermanTagger.removalTagger (removed.txt via CombiningTagger).
	// Used so imperative short forms do not overwrite manually removed tags.
	RemovalTagger tagging.WordTagger
	// SplitCompound optional compound splitter for unknown tokens.
	SplitCompound func(word string) []string
	// AdjExpansion optional spelling.txt /A /P forms (ExpansionInfos.adjInfos).
	AdjExpansion *SpellingAdjExpansion
	// VerbExpansion optional spelling.txt prefix_verb maps (ExpansionInfos.verbInfos).
	VerbExpansion *SpellingVerbExpansion
}

func NewGermanTagger(wt tagging.WordTagger) *GermanTagger {
	t := &GermanTagger{
		BaseTagger: tagging.NewBaseTagger(wt, GermanDictPath, "de", true),
	}
	// Java: removalTagger = (ManualTagger) ((CombiningTagger) getWordTagger()).getRemovalTagger()
	if ct, ok := wt.(*tagging.CombiningTagger); ok {
		t.RemovalTagger = ct.GetRemovalTagger()
	}
	return t
}

// SetSpellingAdjExpansion attaches /A /P expansions (Java expansionInfos.adjInfos).
func (t *GermanTagger) SetSpellingAdjExpansion(ex *SpellingAdjExpansion) {
	if t != nil {
		t.AdjExpansion = ex
	}
}

// SetSpellingVerbExpansion attaches prefix_verb expansions (Java verbInfos/nominalized).
func (t *GermanTagger) SetSpellingVerbExpansion(ex *SpellingVerbExpansion) {
	if t != nil {
		t.VerbExpansion = ex
	}
}

// DefaultGermanTagger is a process-level tagger (empty dict until loaded).
var DefaultGermanTagger = NewGermanTagger(tagging.MapWordTagger{})

// Tag tags tokens with German case/gender-aware retries and sentence-context
// arms (imperative short form, substantivated -er, gender gap, mitarbeitend).
// Java: tag(sentenceTokens) → tag(sentenceTokens, true) with ignoreCase=true.
func (t *GermanTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	return t.TagIgnoreCase(sentenceTokens, true)
}

// TagIgnoreCase ports GermanTagger.tag(List, boolean ignoreCase).
func (t *GermanTagger) TagIgnoreCase(sentenceTokens []string, ignoreCase bool) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	prevWord := ""
	firstWord := true
	for i, word := range sentenceTokens {
		readings := t.tagOneInSentence(word, sentenceTokens, i, firstWord, prevWord, ignoreCase)
		out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
		pos += len([]rune(word))
		if strings.TrimSpace(word) != "" {
			// Java: firstWord = !isAlphanumeric after first successful first-word handling
			if firstWord {
				firstWord = !isAlphanumericDE(word)
			}
			prevWord = word
		}
	}
	return out
}

func isAlphanumericDE(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

// tagOne tags a single word without sentence context.
// Lookup uses ignoreCase=false (Java tag(singleton, false)).
func (t *GermanTagger) tagOne(word string) []*languagetool.AnalyzedToken {
	return t.tagOneInSentence(word, nil, 0, false, "", false)
}

func (t *GermanTagger) tagOneInSentence(word string, sentenceTokens []string, idxPos int, firstWord bool, prevWord string, ignoreCase bool) []*languagetool.AnalyzedToken {
	w := word
	var readings []*languagetool.AnalyzedToken

	// Gender star: jede * r
	var taggerTokens []tagging.TaggedWord
	if gap := t.genderGapTaggerTokens(sentenceTokens, idxPos, word); gap != nil {
		taggerTokens = gap
	} else {
		taggerTokens = t.TagWordExact(w)
	}

	// Case retries only when ignoreCase (Java tag(..., ignoreCase)).
	if ignoreCase {
		// sentence start / after colon: lowercase retry
		if (firstWord || prevWord == ":") && len(taggerTokens) == 0 {
			taggerTokens = t.TagWordExact(strings.ToLower(w))
		} else if idxPos == 0 {
			// Java pos==0: also merge lowercase readings (Haben, Sollen…)
			taggerTokens = append(taggerTokens, t.TagWordExact(strings.ToLower(w))...)
		} else if len(taggerTokens) == 0 && idxPos > 0 {
			// Java: empty → direct speech after : „ Word (indexOf first occurrence)
			idx := indexOfToken(sentenceTokens, word)
			if idx > 2 && sentenceTokens[idx-1] == "„" && sentenceTokens[idx-3] == ":" {
				taggerTokens = append(taggerTokens, t.TagWordExact(strings.ToLower(w))...)
			}
		}
	}

	for _, tw := range taggerTokens {
		readings = append(readings, toToken(word, tw))
	}
	// known-word: IMP:SIN:SFT ↔ 1:SIN:PRÄ:SFT mutual tags (non-separable / bare verbs)
	if len(readings) > 0 {
		readings = t.addImpPraesSFTMutual(word, sentenceTokens, idxPos, readings)
	}

	lower := strings.ToLower(w)
	// No mid-sentence full-lowercase invent: Java only lowercases at sentence start,
	// after ":", after ": „" direct speech, or inside specific unknown-word branches.

	// known word path done; unknown → expansions + sentence-context forms
	if len(readings) == 0 {
		readings = t.tagFromExpansions(word, w, lower)
	}
	if len(readings) == 0 {
		if m := t.tagMitarbeitenden(word); len(m) > 0 {
			readings = m
		}
	}
	if len(readings) == 0 && sentenceTokens != nil {
		if imp := t.getImperativeForm(word, sentenceTokens, idxPos); len(imp) > 0 {
			readings = imp
		}
	}
	if len(readings) == 0 && sentenceTokens != nil {
		if sub := t.getSubstantivatedForms(word, sentenceTokens); len(sub) > 0 {
			readings = sub
		}
	}
	// elative intensifier compounds (supergut, uralt…)
	if len(readings) == 0 {
		if el := t.tagElativeUnknown(word); len(el) > 0 {
			readings = el
		}
	}
	// dash-linked sanitizeWord + separable prefix verbs (:NEB / EIZ)
	if len(readings) == 0 {
		if d := t.tagUnknownDashAndPrefix(word, sentenceTokens, idxPos); len(d) > 0 {
			readings = d
		}
	}

	// compound split fallback (with lemma stem rebuild when multi-part)
	// skip domain-like sequences: example . com
	if len(readings) == 0 && t.SplitCompound != nil && !isDomainLikeSequence(sentenceTokens, idxPos) {
		parts := t.SplitCompound(w)
		if len(parts) > 1 {
			last := parts[len(parts)-1]
			// Java: uppercase last part when word starts upper (except *freie*)
			if tools.StartsWithUppercase(w) && !strings.Contains(last, "freie") &&
				!strings.Contains(last, "freier") && !strings.Contains(last, "freien") &&
				!strings.Contains(last, "freies") && !strings.Contains(last, "freiem") {
				last = tools.UppercaseFirstChar(last)
			}
			lastTags := t.TagWordExact(last)
			if len(lastTags) == 0 && t.AdjExpansion != nil {
				lastTags = t.AdjExpansion.Tag(last)
			}
			// rebuild lemma: part0 + lowercase(part1…) + lowercase(lemma)
			stem := ""
			for i, p := range parts[:len(parts)-1] {
				if i == 0 {
					stem += p
				} else {
					stem += tools.LowercaseFirstChar(p)
				}
			}
			for _, tw := range lastTags {
				if strings.HasPrefix(tw.PosTag, "VER:IMP") {
					continue
				}
				lem := tw.Lemma
				if lem == "" {
					lem = last
				}
				lem = tools.LowercaseFirstChar(lem)
				// Java prfxs.contains(firstPart) + VER:1/2/3 → :NEB lemma prefix+lemma
				pos := tw.PosTag
				if len(parts) >= 2 && isExactSeparablePrefix(strings.ToLower(parts[0])) {
					if (strings.HasPrefix(pos, "VER:1") || strings.HasPrefix(pos, "VER:2") || strings.HasPrefix(pos, "VER:3")) &&
						(idxPos == 0 || word == strings.ToLower(word) || isTitleCaseWord(word)) {
						if !strings.HasSuffix(pos, ":NEB") {
							pos = pos + ":NEB"
						}
					}
				}
				readings = append(readings, toToken(word, tagging.NewTaggedWord(stem+lem, pos)))
			}
		}
	}
	if len(readings) == 0 {
		readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
	}
	return readings
}

// tagFromExpansions ports the expansionInfos branch of GermanTagger for unknown words:
// prefix verbs, nominalized verbs, adj /A /P, *erweise.
func (t *GermanTagger) tagFromExpansions(surface, w, lower string) []*languagetool.AnalyzedToken {
	var readings []*languagetool.AnalyzedToken
	if t.VerbExpansion != nil {
		// Java maps are case-sensitive: verbInfos keys are lowercase; nominalized are Title case.
		// 1) exact surface verbInfos (lowercase infinitives / zu-forms)
		if vi, ok := t.VerbExpansion.LookupVerb(w); ok {
			if vi.Infix == "zu" {
				lemma := vi.Prefix + vi.VerbBaseform
				readings = append(readings, toToken(surface, tagging.NewTaggedWord(lemma, "VER:EIZ:NON")))
				return readings
			}
			// Java: only when prefix starts lowercase
			if tools.StartsWithLowercase(vi.Prefix) {
				pref := vi.Prefix + vi.Infix
				rest := ""
				if strings.HasPrefix(w, pref) {
					rest = w[len(pref):]
				} else if strings.HasPrefix(lower, pref) {
					rest = lower[len(pref):]
				}
				if rest != "" && !isNotAVerb(w) {
					sep := startsWithAnyPrefix(strings.ToLower(vi.Prefix), prefixesSeparableVerbsLongestList)
					for _, tw := range t.TagWordExact(rest) {
						if tw.PosTag == "" {
							continue
						}
						if !(strings.HasPrefix(tw.PosTag, "VER:") || strings.HasPrefix(tw.PosTag, "PA1:") || strings.HasPrefix(tw.PosTag, "PA2:")) {
							continue
						}
						if strings.HasPrefix(tw.PosTag, "VER:MOD") || strings.HasPrefix(tw.PosTag, "VER:AUX") {
							continue
						}
						lemma := vi.Prefix + tw.Lemma
						if tw.Lemma == "" {
							lemma = vi.Prefix + vi.VerbBaseform
						}
						pos := tw.PosTag
						// separable finite → :NEB
						if sep && (strings.HasPrefix(pos, "VER:1") || strings.HasPrefix(pos, "VER:2") || strings.HasPrefix(pos, "VER:3")) {
							pos = pos + ":NEB"
						}
						if sep && strings.HasPrefix(pos, "VER:IMP") {
							continue // separable: no bare IMP on fused form
						}
						readings = append(readings, toToken(surface, tagging.NewTaggedWord(lemma, pos)))
					}
				}
			}
			if len(readings) > 0 {
				return readings
			}
		}
		// 2) nominalized / genitive fixed tags (exact surface, e.g. Herumgeben)
		for _, tw := range t.VerbExpansion.Tag(w) {
			readings = append(readings, toToken(surface, tw))
		}
		if len(readings) > 0 {
			return readings
		}
	}
	if t.AdjExpansion != nil {
		for _, tw := range t.AdjExpansion.Tag(w) {
			readings = append(readings, toToken(surface, tw))
		}
		if len(readings) == 0 && w != lower {
			for _, tw := range t.AdjExpansion.Tag(lower) {
				readings = append(readings, toToken(surface, tw))
			}
		}
		if len(readings) > 0 {
			return readings
		}
	}
	if t.isWeiseException(w) || (w != lower && t.isWeiseException(lower)) {
		for _, tag := range tagsForWeise {
			readings = append(readings, toToken(surface, tagging.NewTaggedWord(surface, tag)))
		}
	}
	return readings
}

// isWeiseException ports GermanTagger.isWeiseException: ends with "erweise" and
// stem has ADJ reading.
func (t *GermanTagger) isWeiseException(word string) bool {
	if !strings.HasSuffix(word, "erweise") {
		return false
	}
	stem := strings.TrimSuffix(word, "erweise")
	if stem == "" {
		return false
	}
	for _, tw := range t.TagWordExact(stem) {
		if strings.HasPrefix(tw.PosTag, "ADJ") {
			return true
		}
	}
	// also try adj expansion for stem
	if t.AdjExpansion != nil {
		for _, tw := range t.AdjExpansion.Tag(stem) {
			if strings.HasPrefix(tw.PosTag, "ADJ") {
				return true
			}
		}
	}
	return false
}

func toToken(surface string, tw tagging.TaggedWord) *languagetool.AnalyzedToken {
	var pos, lemma *string
	if tw.PosTag != "" {
		p := tw.PosTag
		pos = &p
	}
	if tw.Lemma != "" {
		l := tw.Lemma
		lemma = &l
	}
	return languagetool.NewAnalyzedToken(surface, pos, lemma)
}

// SwissGermanTagger ports tagging.de.SwissGermanTagger as GermanTagger with ss↔ß retry.
type SwissGermanTagger struct {
	*GermanTagger
}

func NewSwissGermanTagger(wt tagging.WordTagger) *SwissGermanTagger {
	return &SwissGermanTagger{GermanTagger: NewGermanTagger(wt)}
}

// Tag ports SwissGermanTagger.tag(list) → tag(list, true).
func (t *SwissGermanTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	return t.TagIgnoreCase(sentenceTokens, true)
}

// TagIgnoreCase ports SwissGermanTagger.tag(List, ignoreCase):
// super.tag then for untagged tokens with "ss", Lookup(ss→ß) and addReading.
func (t *SwissGermanTagger) TagIgnoreCase(sentenceTokens []string, ignoreCase bool) []*languagetool.AnalyzedTokenReadings {
	if t == nil || t.GermanTagger == nil {
		return nil
	}
	out := t.GermanTagger.TagIgnoreCase(sentenceTokens, ignoreCase)
	if out == nil {
		return nil
	}
	for i, reading := range out {
		if reading == nil || i >= len(sentenceTokens) {
			continue
		}
		word := sentenceTokens[i]
		if !strings.Contains(word, "ss") || reading.IsTagged() {
			continue
		}
		alt := strings.ReplaceAll(word, "ss", "ß")
		if alt == word {
			continue
		}
		// Java: replacementReading = lookup(ss→ß) — recursive Lookup (ignoreCase=false)
		repl := t.GermanTagger.Lookup(alt)
		if repl == nil {
			continue
		}
		for _, at := range repl.GetReadings() {
			if at == nil || at.GetPOSTag() == nil {
				continue
			}
			// keep Swiss surface (ss)
			reading.AddReading(languagetool.NewAnalyzedToken(word, at.GetPOSTag(), at.GetLemma()), "SwissGermanTagger")
		}
	}
	return out
}

// Lookup ports GermanTagger.lookup via Swiss.tag(singleton, false) polymorphism:
// Swiss ss→ß runs inside TagIgnoreCase; then null if first POS still null.
func (t *SwissGermanTagger) Lookup(word string) *languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := t.TagIgnoreCase([]string{word}, false)
	if len(out) == 0 || out[0] == nil {
		return nil
	}
	rds := out[0].GetReadings()
	if len(rds) == 0 || rds[0].GetPOSTag() == nil {
		return nil
	}
	return out[0]
}

// Lookup ports GermanTagger.lookup:
// tag(singleton, ignoreCase=false); null if first reading has null POS.
func (t *GermanTagger) Lookup(word string) *languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	// Java: tag(Collections.singletonList(word), false)
	out := t.TagIgnoreCase([]string{word}, false)
	if len(out) == 0 || out[0] == nil {
		return nil
	}
	rds := out[0].GetReadings()
	if len(rds) == 0 || rds[0].GetPOSTag() == nil {
		return nil
	}
	return out[0]
}
