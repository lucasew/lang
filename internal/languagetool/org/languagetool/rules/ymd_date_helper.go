package rules

import (
	"fmt"
	"strings"
)

// YMDDateHelper ports org.languagetool.rules.YMDDateHelper.
type YMDDateHelper struct{}

func NewYMDDateHelper() *YMDDateHelper { return &YMDDateHelper{} }

// ParseDate expands args["date"] = "yyyy-mm-dd" into year/month/day keys.
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
