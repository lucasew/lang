package rules

import (
	"fmt"
	"strconv"
	"strings"
)

// YMDDateHelper ports org.languagetool.rules.YMDDateHelper.
type YMDDateHelper struct{}

func NewYMDDateHelper() *YMDDateHelper { return &YMDDateHelper{} }

// ParseDate ports YMDDateHelper.parseDate.
// Java mutates the same args map: expands args["date"] = "yyyy-mm-dd" into year/month/day keys and returns args.
// Missing key "date" → IllegalArgumentException; not three '-' parts → RuntimeException.
func (h *YMDDateHelper) ParseDate(args map[string]string) (map[string]string, error) {
	dateString, ok := args["date"]
	if !ok {
		// Java: args.get("date") == null → IllegalArgumentException("Missing key 'date'")
		return nil, fmt.Errorf("Missing key 'date'")
	}
	// Java String.split("-", limit omitted): trailing empty segments discarded; empty string → {""} length 1.
	parts := strings.Split(dateString, "-")
	if len(parts) != 3 {
		// Java: RuntimeException("Expected date in format 'yyyy-mm-dd': '" + dateString + "'")
		return nil, fmt.Errorf("Expected date in format 'yyyy-mm-dd': '%s'", dateString)
	}
	args["year"] = parts[0]
	args["month"] = parts[1]
	args["day"] = parts[2]
	return args, nil
}

// CorrectDate ports YMDDateHelper.correctDate: replaces {realDate} with year+1-mm-dd,
// copies shortMessage and type onto a new RuleMatch.
func (h *YMDDateHelper) CorrectDate(match *RuleMatch, args map[string]string) *RuleMatch {
	// Java: Integer.parseInt(year) / match getters — NPE or NumberFormatException on bad input.
	year := args["year"]
	month := args["month"]
	day := args["day"]
	y, err := strconv.Atoi(year)
	if err != nil {
		// Java Integer.parseInt throws NumberFormatException.
		panic(err.Error())
	}
	// Java: String.format("%d-%s-%s", correctYear, month, day)
	correctDate := fmt.Sprintf("%d-%s-%s", y+1, month, day)
	// Java String.replace(CharSequence, CharSequence) replaces all non-overlapping occurrences.
	msg := strings.ReplaceAll(match.GetMessage(), "{realDate}", correctDate)
	// Java: new RuleMatch(rule, sentence, from, to, message, shortMessage); setType(match.getType())
	out := NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), msg)
	out.ShortMessage = match.GetShortMessage()
	out.SetType(match.GetType())
	return out
}
