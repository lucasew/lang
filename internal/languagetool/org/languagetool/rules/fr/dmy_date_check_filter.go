package fr

import (
	"fmt"
	"strings"
)

// DMYDateCheckFilter ports org.languagetool.rules.fr.DMYDateCheckFilter.
// Expects date argument as dd-mm-yyyy.
type DMYDateCheckFilter struct {
	dateCheck *DateCheckFilter
}

func NewDMYDateCheckFilter() *DMYDateCheckFilter {
	return &DMYDateCheckFilter{dateCheck: NewDateCheckFilter()}
}

// PrepareArgs expands date=dd-mm-yyyy into day/month/year; rejects direct keys.
func (f *DMYDateCheckFilter) PrepareArgs(args map[string]string) (map[string]string, error) {
	if _, ok := args["year"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for DMYDateCheckFilter")
	}
	if _, ok := args["month"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for DMYDateCheckFilter")
	}
	if _, ok := args["day"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for DMYDateCheckFilter")
	}
	dateString, ok := args["date"]
	if !ok || dateString == "" {
		return nil, fmt.Errorf("missing key 'date'")
	}
	parts := strings.Split(dateString, "-")
	if len(parts) != 3 {
		return nil, fmt.Errorf("expected date in format 'dd-mm-yyyy': %q", dateString)
	}
	out := map[string]string{}
	for k, v := range args {
		out[k] = v
	}
	out["day"] = parts[0]
	out["month"] = parts[1]
	out["year"] = parts[2]
	return out, nil
}
