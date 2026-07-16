package de

import (
	"fmt"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// YMDNewYearDateFilter combines YMD date parsing with NewYearDateFilter.
type YMDNewYearDateFilter struct {
	newYear *NewYearDateFilter
	ymd     *rules.YMDDateHelper
}

func NewYMDNewYearDateFilter() *YMDNewYearDateFilter {
	return &YMDNewYearDateFilter{
		newYear: NewNewYearDateFilter(),
		ymd:     rules.NewYMDDateHelper(),
	}
}

// PrepareArgs parses date=yyyy-mm-dd into year/month/day; rejects direct keys.
func (f *YMDNewYearDateFilter) PrepareArgs(args map[string]string) (map[string]string, error) {
	if _, ok := args["year"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for YMDNewYearDateFilter")
	}
	if _, ok := args["month"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for YMDNewYearDateFilter")
	}
	if _, ok := args["day"]; ok {
		return nil, fmt.Errorf("set only 'weekDay' and 'date' for YMDNewYearDateFilter")
	}
	return f.ymd.ParseDate(args)
}

// ShouldFlag after PrepareArgs.
func (f *YMDNewYearDateFilter) ShouldFlagFromArgs(args map[string]string) (bool, error) {
	parsed, err := f.PrepareArgs(args)
	if err != nil {
		return false, err
	}
	var y, m int
	fmt.Sscanf(parsed["year"], "%d", &y)
	fmt.Sscanf(parsed["month"], "%d", &m)
	return f.newYear.ShouldFlag(y, m), nil
}
