package server

import (
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ForeignScriptIgnoreRanges is retained only as an incomplete diagnostic helper.
// Java ignoreRanges are produced from RuleMatch.getNewLanguageMatches via
// CheckResults (see TextChecker.CheckWithOptionsAndIgnore), not from script
// heuristics. Do not use this for /v2/check production responses.
//
// When called, maps non-Latin script runs to matching altLanguages codes only
// (no invent language IDs outside the provided alt list).
func ForeignScriptIgnoreRanges(text, primaryLang string, altLangs []string) []IgnoreRangeInfo {
	if text == "" || len(altLangs) == 0 {
		return nil
	}
	_ = primaryLang

	type span struct {
		from, to int
		kind     string
	}
	var spans []span
	i := 0
	for i < len(text) {
		r, size := utf8.DecodeRuneInString(text[i:])
		if r == utf8.RuneError && size == 1 {
			i++
			continue
		}
		kind := scriptKind(r)
		if kind == "" || kind == "latin" {
			i += size
			continue
		}
		start := i
		j := i + size
		for j < len(text) {
			r2, sz := utf8.DecodeRuneInString(text[j:])
			k2 := scriptKind(r2)
			// Character.isWhitespace for "soft" separators inside a foreign span —
			// not Go unicode.IsSpace alone (NBSP differs from isWhitespace, but
			// here we allow Zs separators inside spans via IsSpaceChar as well).
			isSep := tools.CharacterIsWhitespace(r2) || tools.CharacterIsSpaceChar(r2) ||
				unicode.IsPunct(r2) || unicode.IsDigit(r2)
			if k2 != kind && !isSep {
				break
			}
			if k2 == kind {
				j += sz
				continue
			}
			if isSep {
				k := j + sz
				found := false
				for k < len(text) {
					r3, sz3 := utf8.DecodeRuneInString(text[k:])
					k3 := scriptKind(r3)
					if k3 == kind {
						found = true
						break
					}
					if k3 != "" && k3 != kind {
						break
					}
					isSep3 := tools.CharacterIsWhitespace(r3) || tools.CharacterIsSpaceChar(r3) ||
						unicode.IsPunct(r3) || unicode.IsDigit(r3)
					if !isSep3 {
						break
					}
					k += sz3
				}
				if !found {
					break
				}
				j += sz
				continue
			}
			break
		}
		// trim trailing whitespace (Character.isWhitespace / isSpaceChar)
		end := j
		for end > start {
			r3, sz3 := utf8.DecodeLastRuneInString(text[start:end])
			if tools.CharacterIsWhitespace(r3) || tools.CharacterIsSpaceChar(r3) {
				end -= sz3
				continue
			}
			break
		}
		if end > start {
			hasLetter := false
			for p := start; p < end; {
				r3, sz3 := utf8.DecodeRuneInString(text[p:end])
				if scriptKind(r3) == kind {
					hasLetter = true
					break
				}
				p += sz3
			}
			if hasLetter {
				spans = append(spans, span{start, end, kind})
			}
		}
		i = j
	}
	if len(spans) == 0 {
		return nil
	}
	out := make([]IgnoreRangeInfo, 0, len(spans))
	for _, s := range spans {
		lang := pickAltForScript(s.kind, altLangs)
		if lang == "" {
			continue
		}
		out = append(out, IgnoreRangeInfo{From: s.from, To: s.to, Lang: lang})
	}
	return out
}

func scriptKind(r rune) string {
	switch {
	case unicode.Is(unicode.Cyrillic, r):
		return "cyrillic"
	case unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r):
		return "cjk"
	case unicode.Is(unicode.Arabic, r):
		return "arabic"
	case unicode.Is(unicode.Greek, r):
		return "greek"
	case unicode.Is(unicode.Latin, r):
		return "latin"
	default:
		return ""
	}
}

func pickAltForScript(kind string, alts []string) string {
	prefer := map[string][]string{
		"cyrillic": {"ru", "uk", "be", "sr", "bg", "mk"},
		"cjk":      {"zh", "ja", "ko"},
		"arabic":   {"ar", "fa", "ur"},
		"greek":    {"el"},
	}
	prefs := prefer[kind]
	for _, a := range alts {
		base := baseLangCode(a)
		for _, p := range prefs {
			if base == p {
				return a
			}
		}
	}
	// No invent: only return an alt when its code matches the script preference list.
	return ""
}

func baseLangCode(code string) string {
	for i := 0; i < len(code); i++ {
		if code[i] == '-' {
			return code[:i]
		}
	}
	return code
}

// filterRemoteByIgnoreRanges drops matches whose span is fully inside any ignore range.
// Java applies ignore ranges from CheckResults when serializing / filtering multi-lang noise.
func filterRemoteByIgnoreRanges(ms []RemoteRuleMatch, ranges []IgnoreRangeInfo) []RemoteRuleMatch {
	if len(ms) == 0 || len(ranges) == 0 {
		return ms
	}
	out := make([]RemoteRuleMatch, 0, len(ms))
	for _, m := range ms {
		end := m.Offset + m.ErrorLength
		drop := false
		for _, r := range ranges {
			if m.Offset >= r.From && end <= r.To {
				drop = true
				break
			}
		}
		if !drop {
			out = append(out, m)
		}
	}
	return out
}
