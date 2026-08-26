package markup

import "fmt"

// MappingValue ports org.languagetool.markup.MappingValue.
type MappingValue struct {
	TotalPosition    int
	FakeMarkupLength int
}

// NewMappingValue ports MappingValue(int totalPosition).
func NewMappingValue(totalPosition int) MappingValue {
	return NewMappingValueFull(totalPosition, 0)
}

// NewMappingValueFull ports MappingValue(int totalPosition, int fakeMarkupLength).
func NewMappingValueFull(totalPosition, fakeMarkupLength int) MappingValue {
	return MappingValue{TotalPosition: totalPosition, FakeMarkupLength: fakeMarkupLength}
}

func (m MappingValue) GetTotalPosition() int    { return m.TotalPosition }
func (m MappingValue) GetFakeMarkupLength() int { return m.FakeMarkupLength }

// String ports MappingValue.toString.
func (m MappingValue) String() string {
	return fmt.Sprintf("totalPos:%d,fakeMarkupLen=%d", m.TotalPosition, m.FakeMarkupLength)
}
