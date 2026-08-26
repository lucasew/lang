package de

import (
	"fmt"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// YMDNewYearDateFilter ports org.languagetool.rules.de.YMDNewYearDateFilter.
// Expects 'date' as yyyy-mm-dd; correctDate then super NewYearDateFilter.
type YMDNewYearDateFilter struct {
	// newYear holds Force* for tests / ShouldFlag helpers.
	newYear *NewYearDateFilter
	core    *rules.NewYearDateFilterCore
	ymd     *rules.YMDDateHelper
}

func NewYMDNewYearDateFilter() *YMDNewYearDateFilter {
	core := deNewYearDateCore()
	return &YMDNewYearDateFilter{
		newYear: NewNewYearDateFilter(),
		core:    core,
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
	if f == nil || f.ymd == nil {
		return nil, fmt.Errorf("missing key 'date'")
	}
	return f.ymd.ParseDate(args)
}

// ShouldFlag after PrepareArgs (test helper).
func (f *YMDNewYearDateFilter) ShouldFlagFromArgs(args map[string]string) (bool, error) {
	parsed, err := f.PrepareArgs(args)
	if err != nil {
		return false, err
	}
	var y, m int
	fmt.Sscanf(parsed["year"], "%d", &y)
	fmt.Sscanf(parsed["month"], "%d", &m)
	// Prefer Force* from newYear on the core for tests.
	core := f.effectiveCore()
	return core.ShouldFlag(y, m), nil
}

func (f *YMDNewYearDateFilter) effectiveCore() *rules.NewYearDateFilterCore {
	core := f.core
	if core == nil {
		core = deNewYearDateCore()
	}
	// Mirror Force* from language filter onto core (Java TestHackHelper / calendar).
	if f.newYear != nil {
		c := *core
		c.ForceJanuary = f.newYear.ForceJanuary
		c.ForceYear = f.newYear.ForceYear
		// Keep GetMonth from original
		if c.GetMonth == nil {
			c.GetMonth = core.GetMonth
		}
		return &c
	}
	return core
}

// AcceptRuleMatch ports YMDNewYearDateFilter.acceptRuleMatch.
// Java: reject year/month/day; parseDate; super.acceptRuleMatch(correctDate(match), args).
func (f *YMDNewYearDateFilter) AcceptRuleMatch(match *rules.RuleMatch, args map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	// Java message uses YMDDateCheckFilter.class.getSimpleName() (upstream quirk).
	if _, ok := args["year"]; ok {
		panic("Set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	if _, ok := args["month"]; ok {
		panic("Set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	if _, ok := args["day"]; ok {
		panic("Set only 'weekDay' and 'date' for YMDDateCheckFilter")
	}
	if f.ymd == nil {
		return nil
	}
	parsed, err := f.ymd.ParseDate(args)
	if err != nil {
		panic(err.Error())
	}
	corrected := f.ymd.CorrectDate(match, parsed)
	core := f.effectiveCore()
	if core == nil {
		return nil
	}
	msg := core.AcceptFromArgs(parsed, corrected.GetMessage())
	if msg == "" {
		return nil
	}
	out := rules.NewRuleMatch(match.GetRule(), match.Sentence, match.GetFromPos(), match.GetToPos(), msg)
	out.ShortMessage = match.ShortMessage
	out.IssueType = match.IssueType
	return out
}
