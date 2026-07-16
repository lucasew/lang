package rules

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ConfusionPairEntry is one replacement reading for a confusion form.
type ConfusionPairEntry struct {
	Token string // correct form
	POS   string // POS tag of the correct form
}

// ConfusionPairs maps lowercase wrong-form → possible correct readings.
type ConfusionPairs map[string][]ConfusionPairEntry

// LoadConfusionPairs parses confusion_pairs.txt (form;replacement;POS per line).
func LoadConfusionPairs(r io.Reader) (ConfusionPairs, error) {
	m := ConfusionPairs{}
	sc := bufio.NewScanner(r)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || line[0] == '#' {
			continue
		}
		parts := strings.Split(line, ";")
		if len(parts) != 3 {
			return nil, fmt.Errorf("format error line %d: expected 3 parts, got %d", lineNo, len(parts))
		}
		form := strings.TrimSpace(parts[0])
		repl := strings.TrimSpace(parts[1])
		pos := strings.TrimSpace(parts[2])
		m[form] = append(m[form], ConfusionPairEntry{Token: repl, POS: pos})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return m, nil
}

// ConfusionCheckFilter ports the ES/CA/PT ConfusionCheckFilter / DiacriticsCheckFilter surface.
type ConfusionCheckFilter struct {
	Pairs ConfusionPairs
	// MessageNoDiacritic replaces MessageDiacritic fragment when replacement lacks accent gain.
	MessageDiacritic   string // e.g. "se escribe con tilde"
	MessageNoDiacritic string // e.g. "se escribe de otra manera"
}

// ConfusionResult is the outcome of Suggest.
type ConfusionResult struct {
	Replacement string
	Message     string // possibly rewritten
	OK          bool
}

// Suggest looks up form (case-insensitive) matching postag regex.
// desiredPOS, if non-empty, must fully match the entry's POS (gender/number filter).
// template is the original suggestion template with {suggestion}/{Suggestion}/{SUGGESTION}.
func (f *ConfusionCheckFilter) Suggest(form, postag, desiredPOS, message, template string) ConfusionResult {
	if f.Pairs == nil {
		return ConfusionResult{}
	}
	original := form
	lower := strings.ToLower(form)
	entries, ok := f.Pairs[lower]
	if !ok {
		return ConfusionResult{}
	}
	var postagRE *regexp.Regexp
	if postag != "" {
		var err error
		postagRE, err = regexp.Compile(postag)
		if err != nil {
			return ConfusionResult{}
		}
	}
	var desiredRE *regexp.Regexp
	if desiredPOS != "" {
		var err error
		desiredRE, err = regexp.Compile("^" + desiredPOS + "$")
		if err != nil {
			// treat as raw match
			desiredRE = regexp.MustCompile(desiredPOS)
		}
	}
	var replacement string
	for _, e := range entries {
		if postagRE != nil && !postagRE.MatchString(e.POS) {
			continue
		}
		if desiredRE != nil && !desiredRE.MatchString(e.POS) {
			continue
		}
		replacement = e.Token
		break
	}
	// When desiredPOS was requested but none matched, suppress (Java returns null).
	// When desiredPOS empty, first postag match wins.
	if replacement == "" {
		return ConfusionResult{}
	}
	msg := message
	if f.MessageDiacritic != "" && f.MessageNoDiacritic != "" {
		if !(HasDiacritics(replacement) && !HasDiacritics(lower)) {
			msg = strings.ReplaceAll(msg, f.MessageDiacritic, f.MessageNoDiacritic)
		}
	}
	if tools.IsAllUppercase(original) {
		replacement = strings.ToUpper(replacement)
	} else if tools.IsCapitalizedWord(original) {
		replacement = tools.UppercaseFirstChar(replacement)
	}
	sugg := template
	if sugg != "" {
		sugg = strings.ReplaceAll(sugg, "{suggestion}", replacement)
		sugg = strings.ReplaceAll(sugg, "{Suggestion}", tools.UppercaseFirstChar(replacement))
		sugg = strings.ReplaceAll(sugg, "{SUGGESTION}", strings.ToUpper(replacement))
	} else {
		sugg = replacement
	}
	return ConfusionResult{Replacement: sugg, Message: msg, OK: true}
}

// HasDiacritics reports common Latin diacritic marks (Spanish/Catalan/Portuguese).
func HasDiacritics(s string) bool {
	for _, r := range s {
		switch r {
		case 'á', 'à', 'â', 'ã', 'ä', 'é', 'è', 'ê', 'ë', 'í', 'ì', 'î', 'ï',
			'ó', 'ò', 'ô', 'õ', 'ö', 'ú', 'ù', 'û', 'ü', 'ý', 'ÿ', 'ñ', 'ç',
			'Á', 'À', 'Â', 'Ã', 'Ä', 'É', 'È', 'Ê', 'Ë', 'Í', 'Ì', 'Î', 'Ï',
			'Ó', 'Ò', 'Ô', 'Õ', 'Ö', 'Ú', 'Ù', 'Û', 'Ü', 'Ý', 'Ñ', 'Ç':
			return true
		}
		// combining marks
		if unicode.Is(unicode.Mn, r) {
			return true
		}
	}
	return false
}
