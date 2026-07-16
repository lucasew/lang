package rules

import (
	"fmt"
	"strconv"
	"strings"
)

// RuleOption ports org.languagetool.rules.RuleOption configuration metadata.
type RuleOption struct {
	DefaultValue         any
	ConfigureText        string
	MinConfigurableValue any
	MaxConfigurableValue any
}

// NewRuleOption builds an option; min/max default to 0/100 when not numeric-ranged.
func NewRuleOption(defaultValue any, configureText string, min, max any) *RuleOption {
	ro := &RuleOption{
		DefaultValue:  defaultValue,
		ConfigureText: configureText,
	}
	if min != nil && max != nil && isNumericLike(defaultValue) {
		ro.MinConfigurableValue = min
		ro.MaxConfigurableValue = max
	} else {
		ro.MinConfigurableValue = 0
		ro.MaxConfigurableValue = 100
	}
	return ro
}

func isNumericLike(o any) bool {
	switch o.(type) {
	case int, int32, int64, float32, float64:
		return true
	default:
		return false
	}
}

// ObjectToString encodes a typed value with a type prefix (i/b/f/d/c/s).
func ObjectToString(o any) string {
	switch v := o.(type) {
	case int:
		return "i" + strconv.Itoa(v)
	case int32:
		// distinguish character vs integer by convention: int32 used as int
		return "i" + strconv.Itoa(int(v))
	case int64:
		return "i" + strconv.FormatInt(v, 10)
	case bool:
		return "b" + strconv.FormatBool(v)
	case float32:
		return "f" + strconv.FormatFloat(float64(v), 'g', -1, 32)
	case float64:
		return "d" + strconv.FormatFloat(v, 'g', -1, 64)
	case string:
		if len([]rune(v)) == 1 {
			// still string encoding 's' unless callers want 'c' — use CharToString helper
		}
		return "s" + v
	default:
		return "s" + fmt.Sprint(o)
	}
}

// CharToString encodes a single character with 'c' prefix.
func CharToString(c rune) string {
	return "c" + string(c)
}

// ObjectsToString joins encoded objects with ';'.
func ObjectsToString(objs []any) string {
	parts := make([]string, len(objs))
	for i, o := range objs {
		parts[i] = ObjectToString(o)
	}
	return strings.Join(parts, ";")
}

// StringToObject decodes a value produced by ObjectToString.
// Bare integers without a type prefix are accepted for LT compatibility.
func StringToObject(s string) (any, error) {
	if s == "" {
		return nil, fmt.Errorf("empty")
	}
	c, str := s[0], s[1:]
	switch c {
	case 's':
		return str, nil
	case 'b':
		return strconv.ParseBool(str)
	case 'f':
		f, err := strconv.ParseFloat(str, 32)
		return float32(f), err
	case 'd':
		return strconv.ParseFloat(str, 64)
	case 'c':
		if str == "" {
			return nil, fmt.Errorf("empty char")
		}
		return []rune(str)[0], nil
	case 'i':
		return strconv.Atoi(str)
	default:
		// old version: plain integer
		return strconv.Atoi(s)
	}
}

// StringToObjects splits on ';' and decodes each part.
func StringToObjects(s string) ([]any, error) {
	if s == "" {
		return nil, nil
	}
	parts := strings.Split(s, ";")
	out := make([]any, len(parts))
	for i, p := range parts {
		o, err := StringToObject(strings.TrimSpace(p))
		if err != nil {
			return nil, err
		}
		out[i] = o
	}
	return out, nil
}
