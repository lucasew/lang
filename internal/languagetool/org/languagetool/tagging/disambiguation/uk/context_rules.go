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

// fem/masc title lemmas for proper-name gender override
var femTitles = map[string]struct{}{
	"пані": {}, "місіс": {}, "місис": {}, "міс": {}, "леді": {}, "княгиня": {}, "німкеня": {},
}
var mascTitles = map[string]struct{}{
	"пан": {}, "містер": {}, "м-р": {}, "сер": {}, "князь": {}, "німець": {}, "поляк": {},
}

var likelyVklySurfaces = map[string]struct{}{
	"суде": {}, "роде": {}, "заходе": {}, "місяченьку": {}, "редакціє": {},
}

// RetagFemNames ports retagFemNames soft: title + name + past verb gender forces name gender.
func RetagFemNames(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens)-2; i++ {
		title := tokens[i]
		name := tokens[i+1]
		verb := tokens[i+2]
		if title == nil || name == nil || verb == nil {
			continue
		}
		// title lemma or surface
		gen := ""
		if titleHas(title, femTitles, "f") || hasPOSStart(title, "noun:anim:f:v_naz:prop:fname") {
			gen = "f"
		} else if titleHas(title, mascTitles, "m") || hasPOSStart(title, "noun:anim:m:v_naz:prop:fname") {
			gen = "m"
		} else {
			continue
		}
		// past verb of same gender
		if !hasPastGender(verb, gen) {
			continue
		}
		prefix := "noun:anim:" + gen + ":v_naz:prop"
		if hasPOSStart(name, prefix) {
			// drop non-matching gender prop readings
			for _, r := range append([]*languagetool.AnalyzedToken(nil), name.GetReadings()...) {
				if r == nil || r.GetPOSTag() == nil {
					continue
				}
				if !strings.HasPrefix(*r.GetPOSTag(), prefix) {
					name.RemoveReading(r, "proper_name_gender_override")
				}
			}
		} else if gen == "f" && hasPOSStart(name, "noun:anim:m:v_naz:prop") {
			// леді Черчилль → retag as fem lname
			for _, r := range append([]*languagetool.AnalyzedToken(nil), name.GetReadings()...) {
				name.RemoveReading(r, "proper_name_gender_override")
			}
			p := "noun:anim:f:v_naz:prop:lname"
			l := name.GetToken()
			name.AddReading(languagetool.NewAnalyzedToken(name.GetToken(), &p, &l), "proper_name_gender_override")
		}
		i++ // skip name
	}
}

func titleHas(tok *languagetool.AnalyzedTokenReadings, set map[string]struct{}, gen string) bool {
	low := strings.ToLower(tok.GetToken())
	if _, ok := set[low]; ok {
		return true
	}
	for _, r := range tok.GetReadings() {
		if r == nil {
			continue
		}
		if r.GetLemma() != nil {
			if _, ok := set[strings.ToLower(*r.GetLemma())]; ok {
				return true
			}
		}
	}
	// also require anim gender hint soft
	_ = gen
	return false
}

func hasPastGender(tok *languagetool.AnalyzedTokenReadings, gen string) bool {
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil {
			continue
		}
		pos := *r.GetPOSTag()
		if strings.HasPrefix(pos, "verb") && strings.Contains(pos, "past") && strings.Contains(pos, ":"+gen) {
			return true
		}
	}
	return false
}

func hasPOSStart(tok *languagetool.AnalyzedTokenReadings, prefix string) bool {
	for _, r := range tok.GetReadings() {
		if r != nil && r.GetPOSTag() != nil && strings.HasPrefix(*r.GetPOSTag(), prefix) {
			return true
		}
	}
	return false
}

// RemoveInanimVKly drops inanim vocative when other cases remain and context is not vocative.
func RemoveInanimVKly(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		if likelyVklyContext(tokens, i) {
			continue
		}
		readings := tok.GetReadings()
		var vkly []*languagetool.AnalyzedToken
		other := false
		for _, r := range readings {
			if r == nil || r.GetPOSTag() == nil {
				continue
			}
			pos := *r.GetPOSTag()
			if strings.HasSuffix(pos, "_END") || pos == "SENT_END" {
				continue
			}
			// inanim v_kly not geo
			if strings.Contains(pos, "noun:inanim:") && strings.Contains(pos, "v_kly") && !strings.Contains(pos, ":geo") {
				vkly = append(vkly, r)
			} else {
				other = true
			}
		}
		if len(vkly) == 0 || !other {
			continue
		}
		for _, r := range vkly {
			if r.GetLemma() != nil && *r.GetLemma() == "зоря" {
				continue
			}
			tok.RemoveReading(r, "inanim_v_kly")
		}
	}
}

func likelyVklyContext(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if tokens[i] == nil {
		return false
	}
	if _, ok := likelyVklySurfaces[strings.ToLower(tokens[i].GetToken())]; ok {
		return true
	}
	if i >= len(tokens)-1 {
		return false
	}
	next := tokens[i+1].GetToken()
	if !isPunctAfterKly(next) {
		return false
	}
	if i > 0 && tokens[i-1] != nil {
		prev := strings.ToLower(tokens[i-1].GetToken())
		if prev == "о" {
			return true
		}
		if hasPOSPrefix(tokens[i-1], "adj") && tokens[i-1].HasPartialPosTag("v_kly") {
			return true
		}
	}
	return false
}

func isPunctAfterKly(s string) bool {
	if s == "!" || s == "?" || s == "," || s == "»" || s == "\"" || s == "…" {
		return true
	}
	if strings.HasPrefix(s, "..") || strings.HasPrefix(s, "...") {
		return true
	}
	return false
}

// RemoveLowerCaseHomonymsForAbbreviations drops non-abbr readings on ALL-CAPS abbr tokens.
func RemoveLowerCaseHomonymsForAbbreviations(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	for _, tok := range input.GetTokensWithoutWhitespace() {
		if tok == nil {
			continue
		}
		s := tok.GetToken()
		if s == "" || !isAllUpperLetters(s) {
			continue
		}
		if !tok.HasPartialPosTag("abbr") {
			continue
		}
		for _, r := range append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...) {
			if r == nil || r.GetPOSTag() == nil {
				continue
			}
			pos := *r.GetPOSTag()
			if strings.HasSuffix(pos, "_END") {
				continue
			}
			if !strings.Contains(pos, ":abbr") {
				tok.RemoveReading(r, "lowercase_vs_abbr")
			}
		}
	}
}

func isAllUpperLetters(s string) bool {
	hasLetter := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return hasLetter
}

// RemovePluralForNames drops plural proper-name readings unless plural context.
func RemovePluralForNames(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		tok := tokens[i]
		if tok == nil {
			continue
		}
		// plural adj/numr/багато before → keep plural names
		if i > 1 && tokens[i-1] != nil {
			prev := tokens[i-1]
			if prev.HasPartialPosTag("adj:p") || prev.HasPartialPosTag("numr") || prev.HasPartialPosTag("number") ||
				prev.HasPartialPosTag(":num") {
				continue
			}
			switch strings.ToLower(prev.GetToken()) {
			case "багато", "мало", "сотня", "півсотня", "два", "дві", "три", "чотири":
				continue
			}
			// prep з/із before plural name (наймолодшого з Моцартів)
			if isPrepZ(prev) {
				continue
			}
		}
		// next is plural lname → keep
		if i+1 < len(tokens) && tokens[i+1] != nil && tokens[i+1].HasPartialPosTag(":lname") && tokens[i+1].HasPartialPosTag(":p:") {
			continue
		}
		var plurals []*languagetool.AnalyzedToken
		other := false
		for _, r := range tok.GetReadings() {
			if r == nil || r.GetPOSTag() == nil {
				continue
			}
			pos := *r.GetPOSTag()
			if strings.HasSuffix(pos, "_END") {
				continue
			}
			// plural prop fname/lname
			if strings.Contains(pos, ":prop") && (strings.Contains(pos, ":p:") || strings.Contains(pos, ":p:v_")) &&
				(strings.Contains(pos, "fname") || strings.Contains(pos, "lname") || strings.Contains(pos, "geo")) {
				plurals = append(plurals, r)
			} else {
				other = true
			}
		}
		if len(plurals) > 0 && other {
			for _, r := range plurals {
				tok.RemoveReading(r, "plural_for_names")
			}
		}
	}
}

func isPrepZ(tok *languagetool.AnalyzedTokenReadings) bool {
	low := strings.ToLower(tok.GetToken())
	if low == "з" || low == "із" || low == "зі" {
		return true
	}
	return tok.HasPartialPosTag("prep") && (low == "з" || low == "із" || low == "зі")
}

// RemoveLowerCaseBadForUpperCaseGood strips :bad readings when surface is capitalized prop.
func RemoveLowerCaseBadForUpperCaseGood(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	for _, tok := range input.GetTokensWithoutWhitespace() {
		if tok == nil || len(tok.GetReadings()) < 2 {
			continue
		}
		s := tok.GetToken()
		if s == "" {
			continue
		}
		rs := []rune(s)
		if !unicode.IsUpper(rs[0]) {
			continue
		}
		if !tok.HasPartialPosTag("prop") {
			continue
		}
		// drop readings with :bad whose lemma equals lowercased form of another lemma
		for _, r := range append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...) {
			if r == nil || r.GetPOSTag() == nil {
				continue
			}
			if strings.Contains(*r.GetPOSTag(), ":bad") {
				tok.RemoveReading(r, "lowercase_bad_vs_uppercase_good")
			}
		}
	}
}

// RemoveVerbImpr drops verb:impr when token is also noun and previous adj agrees in case/gender soft.
func RemoveVerbImpr(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 2; i < len(tokens); i++ {
		tok, prev := tokens[i], tokens[i-1]
		if tok == nil || prev == nil {
			continue
		}
		if !hasPOSPrefix(tok, "verb") || !hasPOSPrefix(tok, "noun") || !hasPOSPrefix(prev, "adj") {
			continue
		}
		hasImpr := false
		for _, r := range tok.GetReadings() {
			if r != nil && r.GetPOSTag() != nil && strings.Contains(*r.GetPOSTag(), "impr") {
				hasImpr = true
				break
			}
		}
		if !hasImpr {
			continue
		}
		// soft: if adj and noun share gender letter or both plural
		if adjNounSoftAgree(prev, tok) {
			for _, r := range append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...) {
				if r != nil && r.GetPOSTag() != nil && strings.HasPrefix(*r.GetPOSTag(), "verb") && strings.Contains(*r.GetPOSTag(), "impr") {
					tok.RemoveReading(r, "not_an_imperative_2")
				}
			}
		}
	}
}

func adjNounSoftAgree(adj, noun *languagetool.AnalyzedTokenReadings) bool {
	// share :p: or same :m/f/n:
	for _, a := range adj.GetReadings() {
		if a == nil || a.GetPOSTag() == nil {
			continue
		}
		ap := *a.GetPOSTag()
		for _, n := range noun.GetReadings() {
			if n == nil || n.GetPOSTag() == nil || !strings.HasPrefix(*n.GetPOSTag(), "noun") {
				continue
			}
			np := *n.GetPOSTag()
			for _, g := range []string{":p:", ":m:", ":f:", ":n:"} {
				if strings.Contains(ap, g) && strings.Contains(np, g) {
					return true
				}
			}
		}
	}
	return false
}

// PreferVocativeWhenBang keeps only v_kly on adj+noun immediately before "!" (звертання).
func PreferVocativeWhenBang(input *languagetool.AnalyzedSentence) {
	if input == nil {
		return
	}
	tokens := input.GetTokensWithoutWhitespace()
	for i := 1; i < len(tokens); i++ {
		if tokens[i] == nil || tokens[i].GetToken() != "!" {
			continue
		}
		// look back at noun then adj
		for j := i - 1; j >= 1 && j >= i-3; j-- {
			tok := tokens[j]
			if tok == nil {
				continue
			}
			if !tok.HasPartialPosTag("v_kly") {
				continue
			}
			// has other cases too → keep only v_kly
			for _, r := range append([]*languagetool.AnalyzedToken(nil), tok.GetReadings()...) {
				if r == nil || r.GetPOSTag() == nil {
					continue
				}
				pos := *r.GetPOSTag()
				if strings.HasSuffix(pos, "_END") {
					continue
				}
				if !strings.Contains(pos, "v_kly") {
					tok.RemoveReading(r, "vkly_zvert")
				}
			}
		}
	}
}
