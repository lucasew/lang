package uk

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// DisambiguateSt keeps "ст." as abbr noun when followed by a number (ст. 208).
// Soft green: drop non-noun/non-abbr noise if present; ensure abbr:xp readings.
func DisambiguateSt(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		surface := strings.ToLower(tok.GetToken())
		if surface != "ст." && surface != "ст" {
			continue
		}
		// look ahead for number
		hasNum := false
		for j := i + 1; j < len(tokens) && j <= i+2; j++ {
			if tokens[j] == nil {
				continue
			}
			if isNumberish(tokens[j]) {
				hasNum = true
				break
			}
		}
		// also "ст. ст. 208"
		if !hasNum && i+1 < len(tokens) {
			next := strings.ToLower(tokens[i+1].GetToken())
			if next == "ст." || next == "ст" {
				for j := i + 2; j < len(tokens) && j <= i+3; j++ {
					if tokens[j] != nil && isNumberish(tokens[j]) {
						hasNum = true
						break
					}
				}
			}
		}
		if !hasNum {
			continue
		}
		// ensure we have noun abbr readings; strip verb-like if any
		readings := append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...)
		for _, r := range readings {
			if r == nil || r.GetPOSTag() == nil {
				continue
			}
			pos := *r.GetPOSTag()
			if strings.HasPrefix(pos, "verb") || strings.HasPrefix(pos, "adj") {
				tok.RemoveReading(r, "dis_st")
			}
		}
		// if untagged after strip, inject soft abbr noun
		if !tok.IsTagged() {
			p := "noun:inanim:f:v_naz:nv:abbr:xp1"
			l := "ст."
			tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &p, &l), "dis_st")
		}
	}
}

func isNumberish(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok.HasPosTag("number") || tok.HasPartialPosTag("number") {
		return true
	}
	s := tok.GetToken()
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// DisambiguatePronPos: "його/її/їх" before noun → keep poss adj; before verb → keep pers.
func DisambiguatePronPos(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens)-1; i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		low := strings.ToLower(tok.GetToken())
		if low != "його" && low != "її" && low != "їх" {
			continue
		}
		next := tokens[i+1]
		if next == nil {
			continue
		}
		nextNoun := hasPOSPrefix(next, "noun")
		nextVerb := hasPOSPrefix(next, "verb")
		readings := append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...)
		if nextNoun && !nextVerb {
			// keep adj:…:pron:pos; drop pure pers if poss present
			hasPos := false
			for _, r := range readings {
				if r != nil && r.GetPOSTag() != nil && strings.Contains(*r.GetPOSTag(), "pron:pos") {
					hasPos = true
					break
				}
			}
			if hasPos {
				for _, r := range readings {
					if r == nil || r.GetPOSTag() == nil {
						continue
					}
					pos := *r.GetPOSTag()
					if strings.Contains(pos, "pron:pers") && !strings.Contains(pos, "pron:pos") {
						tok.RemoveReading(r, "dis_pron_pos")
					}
				}
			}
		}
		if nextVerb && !nextNoun {
			// keep pers; drop pos adj
			for _, r := range readings {
				if r == nil || r.GetPOSTag() == nil {
					continue
				}
				if strings.Contains(*r.GetPOSTag(), "pron:pos") {
					tok.RemoveReading(r, "dis_pron_pos")
				}
			}
		}
	}
}

// DisambiguateYih: "їх" + noun → object/gen pers (drop pos if any leftover).
func DisambiguateYih(input *languagetool.AnalyzedSentence) {
	// same surface family as PronPos; reuse
	DisambiguatePronPos(input)
}

// RetagInitials tags single Cyrillic letter + "." as fname abbr when next is capitalized name-like.
func RetagInitials(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens)-1; i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		s := tok.GetToken()
		if !isInitialSurface(s) {
			continue
		}
		next := tokens[i+1]
		if next == nil {
			continue
		}
		ns := next.GetToken()
		if ns == "" {
			continue
		}
		r0, _ := utf8.DecodeRuneInString(ns)
		if !unicode.IsUpper(r0) {
			continue
		}
		// if already has abbr prop fname, ok; else inject soft
		if tok.HasPartialPosTag("abbr") && tok.HasPartialPosTag("fname") {
			continue
		}
		if !tok.IsTagged() || !tok.HasPartialPosTag("fname") {
			// inject dual gender fname abbr readings
			base := strings.TrimSuffix(s, ".")
			if base == "" {
				base = s
			}
			lemma := base + "."
			for _, pos := range []string{
				"noun:anim:f:v_naz:nv:abbr:prop:fname",
				"noun:anim:m:v_rod:nv:abbr:prop:fname",
				"noun:anim:m:v_zna:nv:abbr:prop:fname",
			} {
				p, l := pos, lemma
				tok.AddReading(languagetool.NewAnalyzedToken(s, &p, &l), "dis_initials")
			}
		}
	}
}

func isInitialSurface(s string) bool {
	if !strings.HasSuffix(s, ".") {
		return false
	}
	core := strings.TrimSuffix(s, ".")
	rs := []rune(core)
	if len(rs) != 1 {
		return false
	}
	return unicode.Is(unicode.Cyrillic, rs[0]) || unicode.IsLetter(rs[0])
}

func hasPOSPrefix(tok *languagetool.AnalyzedTokenReadings, prefix string) bool {
	if tok == nil {
		return false
	}
	for _, r := range tok.GetReadings() {
		if r != nil && r.GetPOSTag() != nil && strings.HasPrefix(*r.GetPOSTag(), prefix) {
			return true
		}
	}
	return false
}
