package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"strings"
)

// UkrainianDictPath is the Java resource path (UkrainianTagger super).
// Official bytes: ua.net.nlp morfologik-ukrainian-lt → /uk/ukrainian.dict
const UkrainianDictPath = "/uk/ukrainian.dict"

type UkrainianTagger struct{ *tagging.BaseTagger }

func NewUkrainianTagger(wt tagging.WordTagger) *UkrainianTagger {
	return &UkrainianTagger{BaseTagger: tagging.NewBaseTagger(wt, UkrainianDictPath, "uk", false)}
}

// exactWordTag is CompoundTagger's wordTagger.tag — raw dict, no BaseTagger case-merge.
func (t *UkrainianTagger) exactWordTag(word string) []tagging.TaggedWord {
	return t.TagWordExact(word)
}

func (t *UkrainianTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		// Java getAnalyzedTokens: ` → ' when not at index 0; surface mutates with word.
		w := strings.ReplaceAll(word, "’", "'")
		surface := word
		if i := strings.Index(w, "`"); i > 0 {
			w = strings.ReplaceAll(w, "`", "'")
			surface = w
		}
		// Java additionalTags strips IGNORED_CHARS (soft hyphen + combining acute)
		// before compound arms; strip early so NAME_SUFFIX / RICCHA see clean forms.
		if clean := languagetool.UkrainianIgnoredCharactersRegex.ReplaceAllString(w, ""); clean != w {
			w = clean
			// Java AnalyzedToken surfaces after strip use cleaned word in additionalTags.
			surface = clean
		}
		// exact lookup for CompoundTagger / additionalTags (Java wordTagger, not case-merge)
		exact := t.exactWordTag
		var readings []*languagetool.AnalyzedToken
		// Java CompoundTagger.generateEntities from official /uk/entities.txt
		if ents := EntityReadings(w); len(ents) > 0 {
			readings = ents
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		if sp := SpecialPOSTag(w); sp != "" {
			p := sp
			lemma := w
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(surface, &p, &lemma)}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		// Java X-подібний / X-вмісний: right adj tags from wordTagger (no invent endings).
		if dyn := DynamicAdjReadings(w, exact); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(surface, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		// Java CompoundTagger left "по" + poAdvMatch (по-сибірськи / по-свинячому).
		if dyn := DynamicPoAdvReadings(w, exact); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(surface, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		// Java rightPartsWithLeftTagMap (гей-но, стривай-бо, …) — left POS from dict.
		if dyn := DynamicRightParticleReadings(w, exact); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(surface, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		// Java doGuessMultiHyphens only when dashIdx != firstDashIdx (2+ hyphens).
		// intj redup (а-а is single-dash tagMatch; гей-гей-гей is multi).
		// а-а single dash is handled via FullTagMatch / equal redup paths.
		if strings.Count(strings.ReplaceAll(strings.ReplaceAll(w, "–", "-"), "—", "-"), "-") >= 2 {
			if dyn := DynamicIntjRedupReadings(w, exact); len(dyn) > 0 {
				for _, d := range dyn {
					p, l := d.POS, d.Lemma
					readings = append(readings, languagetool.NewAnalyzedToken(surface, &p, &l))
				}
				out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
				pos += tagging.UTF16Len(word)
				continue
			}
			// merge/stretch (ва-ре-ни-ки, Та-а-ак) + 3-part entities.
			if stretch := DynamicMultiHyphenStretchReadings(w, exact); len(stretch) > 0 {
				readings = stretch
				out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
				pos += tagging.UTF16Len(word)
				continue
			}
		}
		// Java equalParts redup for весь/усе (Усе-усе, всього-всього).
		if dyn := DynamicEqualRedupReadings(w, exact); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(surface, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		// Java numrAdjMatch runs inside tryOWithAdj after tagMatch (not early).
		// Java NAME_SUFFIX (Мустафа-ага) — left name POS + fixed suffix list.
		if dyn := DynamicNameSuffixReadings(w, exact); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(surface, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		// Java BAD_SUFFIX (був-би) — left dict + :bad.
		if dyn := DynamicBadSuffixReadings(w, exact); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(surface, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		// Java left "аль" → tag Аль-right with :bad (surface becomes Аль-…).
		if dyn := DynamicAlPrefixReadings(w, exact); len(dyn) > 0 {
			for _, d := range dyn {
				p, l, s := d.POS, d.Lemma, d.Surface
				readings = append(readings, languagetool.NewAnalyzedToken(s, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		// Java Вибори-2014 / WORDS_WITH_YEAR.
		if dyn := DynamicYearCompoundReadings(w, exact); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(surface, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		// Java Формула-1 / WORDS_WITH_NUM.
		if dyn := DynamicNumSuffixCompoundReadings(w, exact); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(surface, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		// Java гірко-прегірко / гіркий-прегіркий (right пре+left).
		if dyn := DynamicPreRedupReadings(w, exact); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(surface, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		// Java напівX-напівY dual compounds.
		if dyn := DynamicNapivDualReadings(w, exact); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(surface, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		// Java matchDigitCompound: short endings from LetterEndingForNumericHelper;
		// longer right halves need wordTagger (pass TagWord; fail-closed without hits).
		if dyn := DynamicNumericReadings(w, exact); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(surface, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		if dyn := FixedPartReadings(w, exact); len(dyn) > 0 {
			readings = dyn
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += tagging.UTF16Len(word)
			continue
		}
		for _, tw := range t.TagWord(w) {
			readings = append(readings, toTok(surface, tw))
		}
		lower := strings.ToLower(w)
		if len(readings) == 0 && w != lower && !tools.IsMixedCase(w) {
			for _, tw := range t.TagWord(lower) {
				readings = append(readings, toTok(surface, tw))
			}
		}
		// Java getAnalyzedTokens: en-dash U+2013 → hyphen, then super.getAnalyzedTokens(hyphen)
		// (full tag path incl. compounds). TagWord-only misses numr-numr like "два-три".
		if len(readings) == 0 && strings.Contains(w, "\u2013") && altDashesInWordRE.MatchString(w) {
			hyphenated := strings.ReplaceAll(w, "\u2013", "-")
			// Recurse on hyphen form only (no en-dash → no re-entry here).
			rs := t.Tag([]string{hyphenated})
			if len(rs) > 0 && rs[0] != nil {
				var tagged []*languagetool.AnalyzedToken
				for _, rd := range rs[0].GetReadings() {
					if rd == nil {
						continue
					}
					if rd.GetPOSTag() == nil {
						continue
					}
					// Keep hyphenated token surfaces like Java super.getAnalyzedTokens result.
					tagged = append(tagged, rd)
				}
				if len(tagged) > 0 {
					// Java: newTokens.add(new AnalyzedToken(origWord, null, null));
					// origWord is pre-backtick surface for en-dash (Java origWord before replace).
					tagged = append(tagged, languagetool.NewAnalyzedToken(word, nil, nil))
					readings = tagged
				}
			}
		}
		// Java additionalTags / getAnalyzedTokens alt rewrites (CAPS_INSIDE, з→с, ї→і, convertTokens).
		if len(readings) == 0 {
			if alt := AltTagAdjustReadings(w, t.TagWord); len(alt) > 0 {
				readings = alt
			}
		}
		// Java solid LEFT_O_ADJ_INVALID_PATTERN (len≥9, no hyphen) → adj rest + prefix lemma.
		if len(readings) == 0 {
			if sol := SolidLeftOAdjInvalidReadings(w, t.TagWord); len(sol) > 0 {
				readings = sol
			}
		}
		// Java strip [] → :alt when WORDS_WITH_BRACKETS-like.
		if len(readings) == 0 {
			if br := BracketAltReadings(w, t.TagWord); len(br) > 0 {
				readings = br
			}
		}
		// Java additionalTags RICCHA / OTYI (стодвадцятиріччя, …мільйонний).
		// note: analyzeAllCapitamizedAdj runs after compounds (Java getAnalyzedTokens order).
		if len(readings) == 0 {
			if num := NumericLongFormReadings(w, exact); len(num) > 0 {
				readings = num
			}
		}
		// Java dual capitalized prop compounds (Київ-Прага, lname pairs).
		if len(readings) == 0 {
			if dyn := DynamicDualPropReadings(w, exact); len(dyn) > 0 {
				for _, d := range dyn {
					p, l := d.POS, d.Lemma
					readings = append(readings, languagetool.NewAnalyzedToken(surface, &p, &l))
				}
			}
		}
		// Java MISSING_APO in additionalTags — before guessOtherTags / no-dash.
		// напівяпонка → dict напів'японка + :bad (lemma keeps apostrophe).
		if len(readings) == 0 {
			for _, cand := range MissingApostropheCandidates(w) {
				// Java filter2: drop bad|arch|alt|abbr|slang|subst|short|long then add :bad
				for _, tw := range filterMissingApoTags(exact(cand)) {
					tw2 := tw
					tw2.PosTag = AddIfNotContains(tw2.PosTag, ":bad")
					readings = append(readings, toTok(surface, tw2))
				}
				if len(readings) > 0 {
					break
				}
			}
		}
		// Java MISSING_HYPHEN only: …небудь (not inventing dash after anti/mini/…).
		if len(readings) == 0 {
			for _, cand := range MissingHyphenCandidates(w) {
				if strings.HasSuffix(strings.ToLower(cand), "-небудь") {
					if nebud := NebudMissingHyphenReadings(surface, cand, exact); len(nebud) > 0 {
						readings = nebud
						break
					}
				}
			}
		}
		// Java guessOtherTags: capitalized *штрассе / *дзе / *швілі / *іані paradigms.
		if len(readings) == 0 {
			if other := GuessOtherTagsReadings(w); len(other) > 0 {
				readings = other
			}
		}
		// Java guessOtherTagsInternal no-dash prefixes (експрес-style solid compounds).
		if len(readings) == 0 {
			if nd := DynamicNoDashPrefixReadings(w, exact); len(nd) > 0 {
				readings = nd
			}
		}
		// Java elongated-vowel collapse after untagged (гаааа → га + :alt); needs dict.
		if len(readings) == 0 {
			if alt := ElongatedAltReadings(w, t.TagWord); len(alt) > 0 {
				readings = alt
			}
		}
		// Java ALLCAPS → capitalizeProperName + getAdjustedAnalyzedTokens (full re-tag
		// including compounds: БАЛОГИ-КОЛЕСНІКОВА → Балоги-Колеснікова dual lname).
		// Filter noun.*?:prop.*|noninfl.*; remap surface to original.
		// Java tokens.addAll(newTokens) without de-dupe (УКРАЇНА keeps duplicate :alt).
		if prop := allCapsPropFullReadings(w, surface, t); len(prop) > 0 {
			if len(readings) == 0 {
				readings = prop
			} else {
				readings = append(readings, prop...)
			}
		}
		if len(readings) == 0 {
			if pref := TryNoDashPrefixTags(w, func(right string) []*languagetool.AnalyzedToken {
				var rs []*languagetool.AnalyzedToken
				for _, tw := range exact(right) {
					rs = append(rs, toTok(right, tw))
				}
				low := strings.ToLower(right)
				if len(rs) == 0 && right != low {
					for _, tw := range exact(low) {
						rs = append(rs, toTok(right, tw))
					}
				}
				return rs
			}); len(pref) > 0 {
				readings = pref
			}
		}
		if len(readings) == 0 {
			if np := CompoundNumrPOS(w); np != "" {
				p, l := np, w
				readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(surface, &p, &l)}
			}
		}
		// Java additionalTags COMPOUND_WITH_QUOTES (екс-«депутат») before guessCompoundTag.
		if len(readings) == 0 && (strings.Contains(w, "-") || strings.Contains(w, "\u2013")) {
			if cq := CompoundWithQuotesReadings(w, func(adj string) []*languagetool.AnalyzedToken {
				// Re-enter Tag on quote-stripped form (no quotes → no recursion).
				rs := t.Tag([]string{adj})
				if len(rs) == 0 || rs[0] == nil {
					return nil
				}
				return rs[0].GetReadings()
			}); len(cq) > 0 {
				readings = cq
			}
		}
		// Java doGuessTwoHyphens (exactly two dashes) before/alongside single-dash tagMatch.
		if len(readings) == 0 && strings.Count(strings.ReplaceAll(strings.ReplaceAll(w, "–", "-"), "—", "-"), "-") == 2 {
			if two := DynamicTwoHyphenReadings(w, exact); len(two) > 0 {
				readings = two
			}
		}
		// Java single-dash: з-за… alt, invalid prefixes, dash_prefixes / LAT / top-numr
		// (before tagMatch).
		if len(readings) == 0 && strings.Contains(w, "-") {
			if one := DynamicSingleLetterRedupReadings(w, exact); len(one) > 0 {
				readings = one
			}
		}
		if len(readings) == 0 && strings.Contains(w, "-") {
			if inv := DynamicInvalidDashPrefixReadings(w, exact); len(inv) > 0 {
				readings = inv
			}
		}
		if len(readings) == 0 && strings.Contains(w, "-") {
			if dp := DynamicDashPrefixReadings(w, exact); len(dp) > 0 {
				readings = dp
			}
		}
		// Java півгодини-годину (left p:v_*:nv + right noun:inanim → :p:)
		if len(readings) == 0 && strings.Contains(w, "-") {
			if piv := DynamicPivNvDualReadings(w, exact); len(piv) > 0 {
				readings = piv
			}
		}
		// Java upper-right arm (adj:bad solid lower, tryOWithAdj)
		if len(readings) == 0 && strings.Contains(w, "-") {
			if ur := DynamicUpperRightCompoundReadings(w, exact); len(ur) > 0 {
				readings = ur
			}
		}
		// Java tagMatch after dash prefixes / upper-right (before final tryOWithAdj).
		if len(readings) == 0 && strings.Contains(w, "-") && AllowFullTagMatch(w, exact) {
			if ft := FullTagMatchReadings(w, exact); len(ft) > 0 {
				readings = ft
			}
		}
		// Java final tryOWithAdj after failed tagMatch (includes numrAdjMatch + oAdjMatch).
		if len(readings) == 0 && strings.Contains(w, "-") {
			if oadj := DynamicFinalOAdjReadings(w, exact); len(oadj) > 0 {
				readings = oadj
			}
		}
		// Directional oAdj as last resort (same tryOWithAdj family).
		if len(readings) == 0 {
			if dyn := DynamicDirectionalAdjReadings(w, exact); len(dyn) > 0 {
				for _, d := range dyn {
					p, l := d.POS, d.Lemma
					readings = append(readings, languagetool.NewAnalyzedToken(surface, &p, &l))
				}
			}
		}
		// Java analyzeAllCapitamizedAdj after compounds (getAnalyzedTokens order):
		// always attempted; merges into existing readings when already tagged.
		if adj := AnalyzeAllCapitalizedAdj(w, exact); len(adj) > 0 {
			if len(readings) == 0 {
				readings = adj
			} else {
				readings = mergeUniqueAnalyzedTokens(readings, adj)
			}
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(surface, nil, nil)}
		}
		out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
		pos += tagging.UTF16Len(word)
	}
	return out
}

func toTok(surface string, tw tagging.TaggedWord) *languagetool.AnalyzedToken {
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

// allCapsPropFullReadings ports UkrainianTagger ALLCAPS arm:
// capitalizeProperName + getAdjustedAnalyzedTokens (full re-analysis for hyphenated
// forms; dict-only for solid tokens). Filters noun.*?:prop.*|noninfl.* full match.
func allCapsPropFullReadings(word, surface string, t *UkrainianTagger) []*languagetool.AnalyzedToken {
	if t == nil || tagging.UTF16Len(word) <= 2 || !isAllUppercaseUk(word) {
		return nil
	}
	// pure latin romans handled as number:latin
	if reLatinNum.MatchString(word) {
		return nil
	}
	adjusted := capitalizeProperName(word)
	if adjusted == "" || adjusted == word {
		return nil
	}
	// Solid ALLCAPS: dict re-tag only (УКРАЇНА → Україна).
	if !strings.Contains(adjusted, "-") {
		return AllCapsPropReadings(word, t.TagWord)
	}
	// Hyphenated ALLCAPS: full Tag on title-case form (БАЛОГИ-КОЛЕСНІКОВА → dual lname).
	// Recursion-safe: adjusted is not ALLCAPS.
	rs := t.Tag([]string{adjusted})
	if len(rs) == 0 || rs[0] == nil {
		return nil
	}
	var out []*languagetool.AnalyzedToken
	for _, rd := range rs[0].GetReadings() {
		if rd == nil || rd.GetPOSTag() == nil {
			continue
		}
		pos := *rd.GetPOSTag()
		// Java Pattern.compile("noun.*?:prop.*|noninfl.*") Matcher.matches()
		if !allCapsPropPOSRE.MatchString(pos) {
			continue
		}
		lem := ""
		if rd.GetLemma() != nil {
			lem = *rd.GetLemma()
		}
		p, l := pos, lem
		out = append(out, languagetool.NewAnalyzedToken(surface, &p, &l))
	}
	return out
}

// mergeUniqueAnalyzedTokens appends extras whose (POS, lemma) are not already in base
// (Java analyzeAllCapitamizedAdj / ALLCAPS: !tokens.contains(token)).
func mergeUniqueAnalyzedTokens(base, extras []*languagetool.AnalyzedToken) []*languagetool.AnalyzedToken {
	if len(extras) == 0 {
		return base
	}
	seen := make(map[string]struct{}, len(base)+len(extras))
	key := func(t *languagetool.AnalyzedToken) string {
		if t == nil {
			return ""
		}
		pos, lem := "", ""
		if t.GetPOSTag() != nil {
			pos = *t.GetPOSTag()
		}
		if t.GetLemma() != nil {
			lem = *t.GetLemma()
		}
		return pos + "\x00" + lem
	}
	for _, t := range base {
		seen[key(t)] = struct{}{}
	}
	out := base
	for _, t := range extras {
		k := key(t)
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}
		out = append(out, t)
	}
	return out
}
