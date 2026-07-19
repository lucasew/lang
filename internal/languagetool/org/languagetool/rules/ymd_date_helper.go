package rules

import (
	"fmt"
	"strconv"
	"strings"
)

// YMDDateHelper ports org.languagetool.rules.YMDDateHelper.
type YMDDateHelper struct{}

func NewYMDDateHelper() *YMDDateHelper { return &YMDDateHelper{} }

// ParseDate expands args["date"] = "yyyy-mm-dd" into year/month/day keys.
// Mutates a copy (Go maps are refs; we still copy keys so callers can pass shared maps safely).
func (h *YMDDateHelper) ParseDate(args map[string]string) (map[string]string, error) {
	dateString, ok := args["date"]
	if !ok || dateString == "" {
		return nil, fmt.Errorf("missing key 'date'")
	}
	parts := strings.Split(dateString, "-")
	if len(parts) != 3 {
		return nil, fmt.Errorf("expected date in format 'yyyy-mm-dd': %q", dateString)
	}
	out := map[string]string{}
	for k, v := range args {
		out[k] = v
	}
	out["year"] = parts[0]
	out["month"] = parts[1]
	out["day"] = parts[2]
	return out, nil
}

// CorrectDate ports YMDDateHelper.correctDate: replaces {realDate} with year+1-mm-dd.
func (h *YMDDateHelper) CorrectDate(match *RuleMatch, args map[string]string) *RuleMatch {
	if match == nil {
		return nil
	}
	year := args["year"]
	month := args["month"]
	day := args["day"]
	y, err := strconv.Atoi(year)
	if err != nil {
		// Java Integer.parseInt throws; surface as panic for parity with other date filters.
		panic(err.Error())
	}
	correctDate := fmt.Sprintf("%d-%s-%s", y+1, month, day)
	msg := strings.ReplaceAll(match.GetMessage(), "{realDate}", correctDate)
	out := NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), msg)
	out.ShortMessage = match.ShortMessage
	out.IssueType = match.IssueType
	return out
}
