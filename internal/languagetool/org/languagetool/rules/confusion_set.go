package rules

import (
	"fmt"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ConfusionSet ports org.languagetool.rules.ConfusionSet.
type ConfusionSet struct {
	set    map[string]*ConfusionString // key: str + "\x00" + desc
	factor int64
}

func confusionKey(c *ConfusionString) string {
	d := ""
	if c.description != nil {
		d = *c.description
	}
	return c.str + "\x00" + d
}

func NewConfusionSet(factor int64, words ...string) *ConfusionSet {
	if factor < 1 {
		panic(fmt.Sprintf("factor must be >= 1: %d", factor))
	}
	cs := &ConfusionSet{set: map[string]*ConfusionString{}, factor: factor}
	for _, w := range words {
		c := NewConfusionString(w, nil)
		cs.set[confusionKey(c)] = c
	}
	return cs
}

func NewConfusionSetFromList(factor int64, confusionStrings []*ConfusionString) *ConfusionSet {
	if factor < 1 {
		panic(fmt.Sprintf("factor must be >= 1: %d", factor))
	}
	cs := &ConfusionSet{set: map[string]*ConfusionString{}, factor: factor}
	for _, c := range confusionStrings {
		cs.set[confusionKey(c)] = c
	}
	return cs
}

func (c *ConfusionSet) GetFactor() int64 { return c.factor }

func (c *ConfusionSet) GetSet() []*ConfusionString {
	out := make([]*ConfusionString, 0, len(c.set))
	for _, v := range c.set {
		out = append(out, v)
	}
	return out
}

func (c *ConfusionSet) GetUppercaseFirstCharSet() []*ConfusionString {
	var out []*ConfusionString
	for _, s := range c.set {
		out = append(out, NewConfusionString(tools.UppercaseFirstChar(s.GetString()), s.GetDescription()))
	}
	return out
}

func (c *ConfusionSet) String() string {
	var parts []string
	for _, s := range c.set {
		parts = append(parts, s.String())
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func (c *ConfusionSet) Equals(o *ConfusionSet) bool {
	if c == o {
		return true
	}
	if c == nil || o == nil {
		return false
	}
	if c.factor != o.factor || len(c.set) != len(o.set) {
		return false
	}
	for k, v := range c.set {
		ov, ok := o.set[k]
		if !ok || !v.Equal(ov) {
			return false
		}
	}
	return true
}
