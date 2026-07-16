package en

import (
	"fmt"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// YMDDateCheckFilter ports org.languagetool.rules.en.YMDDateCheckFilter argument checks.
type YMDDateCheckFilter struct {
	dateCheck *DateCheckFilter
	ymd       *rules.YMDDateHelper
}

func NewYMDDateCheckFilter() *YMDDateCheckFilter {
	return &YMDDateCheckFilter{
		dateCheck: NewDateCheckFilter(),
		ymd:       rules.NewYMDDateHelper(),
	}
}

func (f *YMDDateCheckFilter) PrepareArgs(args map[string]string) (map[string]string, error) {
	if _, ok := args["year"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	if _, ok := args["month"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	if _, ok := args["day"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	return f.ymd.ParseDate(args)
}
