package uk

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
	"strings"
)

const UkrainianDictPath = "/uk/uk.dict"

type UkrainianTagger struct{ *tagging.BaseTagger }

func NewUkrainianTagger(wt tagging.WordTagger) *UkrainianTagger {
	return &UkrainianTagger{BaseTagger: tagging.NewBaseTagger(wt, UkrainianDictPath, "uk", false)}
}

func (t *UkrainianTagger) Tag(sentenceTokens []string) []*languagetool.AnalyzedTokenReadings {
	if t == nil {
		return nil
	}
	out := make([]*languagetool.AnalyzedTokenReadings, 0, len(sentenceTokens))
	pos := 0
	for _, word := range sentenceTokens {
		w := strings.ReplaceAll(word, "’", "'")
		var readings []*languagetool.AnalyzedToken
		// Java CompoundTagger.generateEntities from official /uk/entities.txt
		if ents := EntityReadings(w); len(ents) > 0 {
			readings = ents
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		if sp := SpecialPOSTag(w); sp != "" {
			p := sp
			lemma := w
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &lemma)}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java X-подібний / X-вмісний: right adj tags from wordTagger (no invent endings).
		if dyn := DynamicAdjReadings(w, t.TagWord); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java CompoundTagger left "по" + poAdvMatch (по-сибірськи / по-свинячому).
		if dyn := DynamicPoAdvReadings(w, t.TagWord); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java rightPartsWithLeftTagMap (гей-но, стривай-бо, …) — left POS from dict.
		if dyn := DynamicRightParticleReadings(w, t.TagWord); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java doGuessMultiHyphens intj redup (а-а, гей-гей-гей) — dict-gated.
		if dyn := DynamicIntjRedupReadings(w, t.TagWord); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java doGuessMultiHyphens merge/stretch (ва-ре-ни-ки, Та-а-ак) + 3-part entities.
		if stretch := DynamicMultiHyphenStretchReadings(w, t.TagWord); len(stretch) > 0 {
			readings = stretch
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java equalParts redup for весь/усе (Усе-усе, всього-всього).
		if dyn := DynamicEqualRedupReadings(w, t.TagWord); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java numrAdjMatch (дво-триметровий…) — left numr + right adj from dict.
		if dyn := DynamicNumrAdjReadings(w, t.TagWord); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java NAME_SUFFIX (Мустафа-ага) — left name POS + fixed suffix list.
		if dyn := DynamicNameSuffixReadings(w, t.TagWord); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java BAD_SUFFIX (був-би) — left dict + :bad.
		if dyn := DynamicBadSuffixReadings(w, t.TagWord); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java left "аль" → tag Аль-right with :bad.
		if dyn := DynamicAlPrefixReadings(w, t.TagWord); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java Вибори-2014 / WORDS_WITH_YEAR.
		if dyn := DynamicYearCompoundReadings(w, t.TagWord); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java Формула-1 / WORDS_WITH_NUM.
		if dyn := DynamicNumSuffixCompoundReadings(w, t.TagWord); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java гірко-прегірко / гіркий-прегіркий (right пре+left).
		if dyn := DynamicPreRedupReadings(w, t.TagWord); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java напівX-напівY dual compounds.
		if dyn := DynamicNapivDualReadings(w, t.TagWord); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		// Java matchDigitCompound: short endings from LetterEndingForNumericHelper;
		// longer right halves need wordTagger (pass TagWord; fail-closed without hits).
		if dyn := DynamicNumericReadings(w, t.TagWord); len(dyn) > 0 {
			for _, d := range dyn {
				p, l := d.POS, d.Lemma
				readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
			}
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		if dyn := FixedPartReadings(w, t.TagWord); len(dyn) > 0 {
			readings = dyn
			out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
			pos += len([]rune(word))
			continue
		}
		for _, tw := range t.TagWord(w) {
			readings = append(readings, toTok(word, tw))
		}
		lower := strings.ToLower(w)
		if len(readings) == 0 && w != lower && !tools.IsMixedCase(w) {
			for _, tw := range t.TagWord(lower) {
				readings = append(readings, toTok(word, tw))
			}
		}
		// Java getAnalyzedTokens: en-dash U+2013 → hyphen re-tag (+ null reading for surface).
		if len(readings) == 0 {
			if alt := AltDashReadings(w, t.TagWord); len(alt) > 0 {
				readings = alt
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
		// Java analyzeAllCapitamizedAdj (Івано-Франківська as adj):
		// always attempted; merges into existing readings when already tagged.
		if adj := AnalyzeAllCapitalizedAdj(w, t.TagWord); len(adj) > 0 {
			if len(readings) == 0 {
				readings = adj
			} else {
				readings = mergeUniqueAnalyzedTokens(readings, adj)
			}
		}
		// Java CompoundTagger.oAdjMatch: only after dict miss; right adj from wordTagger (no invent endings).
		if len(readings) == 0 {
			if dyn := DynamicDirectionalAdjReadings(w, t.TagWord); len(dyn) > 0 {
				for _, d := range dyn {
					p, l := d.POS, d.Lemma
					readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
				}
			}
		}
		// Java additionalTags RICCHA / OTYI (стодвадцятиріччя, …мільйонний).
		if len(readings) == 0 {
			if num := NumericLongFormReadings(w, t.TagWord); len(num) > 0 {
				readings = num
			}
		}
		// Java dual capitalized prop compounds (Київ-Прага, lname pairs).
		if len(readings) == 0 {
			if dyn := DynamicDualPropReadings(w, t.TagWord); len(dyn) > 0 {
				for _, d := range dyn {
					p, l := d.POS, d.Lemma
					readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
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
			if nd := DynamicNoDashPrefixReadings(w, t.TagWord); len(nd) > 0 {
				readings = nd
			}
		}
		// Java elongated-vowel collapse after untagged (гаааа → га + :alt); needs dict.
		if len(readings) == 0 {
			if alt := ElongatedAltReadings(w, t.TagWord); len(alt) > 0 {
				readings = alt
			}
		}
		// Java ALLCAPS → capitalizeProperName + noun.*:prop|noninfl re-tag (needs dict).
		// Merges when surface already has tags (Java tokens.addAll(newTokens)).
		if prop := AllCapsPropReadings(w, t.TagWord); len(prop) > 0 {
			if len(readings) == 0 {
				readings = prop
			} else {
				readings = mergeUniqueAnalyzedTokens(readings, prop)
			}
		}
		if len(readings) == 0 {
			if pref := TryNoDashPrefixTags(w, func(right string) []*languagetool.AnalyzedToken {
				var rs []*languagetool.AnalyzedToken
				for _, tw := range t.TagWord(right) {
					rs = append(rs, toTok(right, tw))
				}
				low := strings.ToLower(right)
				if len(rs) == 0 && right != low {
					for _, tw := range t.TagWord(low) {
						rs = append(rs, toTok(right, tw))
					}
				}
				return rs
			}); len(pref) > 0 {
				readings = pref
			}
		}
		if len(readings) == 0 {
			for _, cand := range MissingApostropheCandidates(w) {
				// Java filter2: drop bad|arch|alt|abbr|slang|subst|short|long then add :bad
				for _, tw := range filterMissingApoTags(t.TagWord(cand)) {
					tw2 := tw
					tw2.PosTag = AddIfNotContains(tw2.PosTag, ":bad")
					readings = append(readings, toTok(word, tw2))
				}
				if len(readings) > 0 {
					break
				}
			}
		}
		if len(readings) == 0 {
			for _, cand := range MissingHyphenCandidates(w) {
				// re-tag hyphenated candidate via compound path specials
				if dyn := DynamicAdjReadings(cand, t.TagWord); len(dyn) > 0 {
					for _, d := range dyn {
						p, l := d.POS, d.Lemma
						readings = append(readings, languagetool.NewAnalyzedToken(word, &p, &l))
					}
					break
				}
				if pref := TryNoDashPrefixTags(strings.ReplaceAll(cand, "-", ""), func(right string) []*languagetool.AnalyzedToken {
					var rs []*languagetool.AnalyzedToken
					for _, tw := range t.TagWord(right) {
						rs = append(rs, toTok(right, tw))
					}
					return rs
				}); len(pref) > 0 {
					readings = pref
					break
				}
				// -небудь indefinite pronouns (Java MISSING_HYPHEN + pronoun POS + :bad)
				if strings.HasSuffix(strings.ToLower(cand), "-небудь") {
					if nebud := NebudMissingHyphenReadings(word, cand, t.TagWord); len(nebud) > 0 {
						readings = nebud
						break
					}
				}
				// direct right-of-hyphen dict lookup (мінітест → міні-тест → тест)
				if i := strings.Index(cand, "-"); i > 0 {
					right := cand[i+1:]
					for _, tw := range t.TagWord(right) {
						readings = append(readings, toTok(word, tw))
					}
					if low := strings.ToLower(right); len(readings) == 0 && right != low {
						for _, tw := range t.TagWord(low) {
							readings = append(readings, toTok(word, tw))
						}
					}
					if len(readings) > 0 {
						break
					}
				}
			}
		}
		if len(readings) == 0 {
			if np := CompoundNumrPOS(w); np != "" {
				p, l := np, w
				readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, &p, &l)}
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
			if two := DynamicTwoHyphenReadings(w, t.TagWord); len(two) > 0 {
				readings = two
			}
		}
		// Java single-dash: з-за… alt, invalid prefixes, dash_prefixes / LAT / top-numr.
		if len(readings) == 0 && strings.Contains(w, "-") {
			if one := DynamicSingleLetterRedupReadings(w, t.TagWord); len(one) > 0 {
				readings = one
			}
		}
		if len(readings) == 0 && strings.Contains(w, "-") {
			if inv := DynamicInvalidDashPrefixReadings(w, t.TagWord); len(inv) > 0 {
				readings = inv
			}
		}
		if len(readings) == 0 && strings.Contains(w, "-") {
			if dp := DynamicDashPrefixReadings(w, t.TagWord); len(dp) > 0 {
				readings = dp
			}
		}
		// Java півгодини-годину (left p:v_*:nv + right noun:inanim → :p:)
		if len(readings) == 0 && strings.Contains(w, "-") {
			if piv := DynamicPivNvDualReadings(w, t.TagWord); len(piv) > 0 {
				readings = piv
			}
		}
		// Java upper-right arm (adj:bad solid lower, tryOWithAdj)
		if len(readings) == 0 && strings.Contains(w, "-") {
			if ur := DynamicUpperRightCompoundReadings(w, t.TagWord); len(ur) > 0 {
				readings = ur
			}
		}
		// Java tagMatch only after pron/part guards, left len, no-dash solid, upper-right.
		if len(readings) == 0 && strings.Contains(w, "-") && AllowFullTagMatch(w, t.TagWord) {
			if ft := FullTagMatchReadings(w, t.TagWord); len(ft) > 0 {
				readings = ft
			}
		}
		// Java final tryOWithAdj after failed tagMatch
		if len(readings) == 0 && strings.Contains(w, "-") {
			if oadj := DynamicFinalOAdjReadings(w, t.TagWord); len(oadj) > 0 {
				readings = oadj
			}
		}
		if len(readings) == 0 {
			readings = []*languagetool.AnalyzedToken{languagetool.NewAnalyzedToken(word, nil, nil)}
		}
		out = append(out, languagetool.NewAnalyzedTokenReadingsList(readings, pos))
		pos += len([]rune(word))
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
