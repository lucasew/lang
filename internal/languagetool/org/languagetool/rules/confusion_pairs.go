package rules

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
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

// GenderNumberProbe maps gendernumberFrom token POS → desired replacement POS pattern
// (Java MS/FS/MP/… Pattern pairs in ConfusionCheckFilter / DiacriticsCheckFilter).
type GenderNumberProbe struct {
	// Probe matches patternTokens[i] via MatchesPosTagRegex.
	Probe string
	// Desired fully matches the confusion-pair entry POS (Matcher.matches).
	Desired string
}

// ESGenderNumberProbes ports org.languagetool.rules.es.ConfusionCheckFilter patterns.
var ESGenderNumberProbes = []GenderNumberProbe{
	{`[NAPD].+MS.*|V.P..SM`, `NC[MC][SN]000|A..[MC][SN].|V.P..SM`},
	{`[NAPD].+MP.*|V.P..PM`, `NC[MC][PN]000|A..[MC][PN].|V.P..PM`},
	{`[NAPD].+FS.*|V.P..SF`, `NC[FC][SN]000|A..[FC][SN].|V.P..SF`},
	{`[NAPD].+FP.*|V.P..PF`, `NC[FC][PN]000|A..[FC][PN].|V.P..PF`},
	{`[NAPD].+CP.*|V.P..P.`, `NC[MFC][PN]000|A..[MFC][PN].|V.P..P.`},
	{`[NAPD].+CS.*|V.P..S.`, `NC[MFC][SN]000|A..[MFC][SN].|V.P..S.`},
}

// CAGenderNumberProbes ports org.languagetool.rules.ca.DiacriticsCheckFilter patterns.
var CAGenderNumberProbes = []GenderNumberProbe{
	{`[NAPD].+MS.*|V.P..SM.`, `NC[MC][SN]000|A..[MC][SN].|V.P..SM.`},
	{`[NAPD].+MP.*|V.P..PM.`, `NC[MC][PN]000|A..[MC][PN].|V.P..PM.`},
	{`[NAPD].+FS.*|V.P..SF.`, `NC[FC][SN]000|A..[FC][SN].|V.P..SF.`},
	{`[NAPD].+FP.*|V.P..PF.`, `NC[FC][PN]000|A..[FC][PN].|V.P..PF.`},
	{`[NAPD].+CP.*|V.P..P..`, `NC[MFC][PN]000|A..[MFC][PN].|V.P..P..`},
	{`[NAPD].+CS.*|V.P..S..`, `NC[MFC][SN]000|A..[MFC][SN].|V.P..S..`},
}

// ConfusionCheckFilter ports the ES/CA/PT ConfusionCheckFilter / DiacriticsCheckFilter surface.
type ConfusionCheckFilter struct {
	Pairs ConfusionPairs
	// MessageNoDiacritic replaces MessageDiacritic fragment when replacement lacks accent gain.
	MessageDiacritic   string // e.g. "se escribe con tilde"
	MessageNoDiacritic string // e.g. "se escribe de otra manera"
	// GenderProbes nil → ESGenderNumberProbes.
	GenderProbes []GenderNumberProbe
	// ExpandAllSuggestions: CA applies template to every match suggestion; ES uses first only.
	ExpandAllSuggestions bool
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
		// Java Matcher.matches() on the gender Pattern (full string).
		var err error
		desiredRE, err = regexp.Compile("^(?:" + desiredPOS + ")$")
		if err != nil {
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

// AcceptRuleMatch ports ConfusionCheckFilter / DiacriticsCheckFilter.acceptRuleMatch.
func (f *ConfusionCheckFilter) AcceptRuleMatch(match *RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	postag, ok := arguments["postag"]
	if !ok {
		panic("Missing key 'postag'")
	}
	originalForm, ok := arguments["form"]
	if !ok {
		panic("Missing key 'form'")
	}
	desiredPOS := ""
	if gn, ok := arguments["gendernumberFrom"]; ok && gn != "" {
		i, err := strconv.Atoi(gn)
		if err != nil || i < 1 || i > len(patternTokens) {
			panic(fmt.Sprintf("ConfusionCheckFilter: Index out of bounds, value: %s", gn))
		}
		desiredPOS = f.desiredPOSFromToken(patternTokens[i-1])
		// Java: gendernumberFrom set but no probe match → leave desired null and skip
		// replacement (only assign when pattern non-null or gendernumberFrom null).
		if desiredPOS == "" {
			return nil
		}
	}
	template := ""
	reps := match.GetSuggestedReplacements()
	if len(reps) > 0 {
		template = reps[0]
	}
	res := f.Suggest(originalForm, postag, desiredPOS, match.GetMessage(), template)
	if !res.OK {
		return nil
	}
	out := NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), res.Message)
	out.ShortMessage = match.ShortMessage
	if f.ExpandAllSuggestions && len(reps) > 0 {
		// CA: rewrite every suggestion template with the same replacement token.
		// Re-run casing on base replacement (Suggest already cased into res.Replacement
		// when template is empty; with templates, extract bare token from first result).
		base := res.Replacement
		// When template substituted, base may be full phrase — recompute bare form:
		bare := f.bareReplacement(originalForm, postag, desiredPOS)
		if bare == "" {
			bare = base
		}
		var all []string
		for _, t := range reps {
			s := strings.ReplaceAll(t, "{suggestion}", bare)
			s = strings.ReplaceAll(s, "{Suggestion}", tools.UppercaseFirstChar(bare))
			s = strings.ReplaceAll(s, "{SUGGESTION}", strings.ToUpper(bare))
			all = append(all, s)
		}
		out.SetSuggestedReplacements(all)
	} else {
		out.SetSuggestedReplacement(res.Replacement)
	}
	return out
}

func (f *ConfusionCheckFilter) bareReplacement(form, postag, desiredPOS string) string {
	// Suggest with empty template returns cased replacement token.
	r := f.Suggest(form, postag, desiredPOS, "", "")
	if !r.OK {
		return ""
	}
	return r.Replacement
}

func (f *ConfusionCheckFilter) desiredPOSFromToken(atr *languagetool.AnalyzedTokenReadings) string {
	if atr == nil {
		return ""
	}
	probes := f.GenderProbes
	if len(probes) == 0 {
		probes = ESGenderNumberProbes
	}
	for _, p := range probes {
		if atr.MatchesPosTagRegex(p.Probe) {
			return p.Desired
		}
	}
	return ""
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
