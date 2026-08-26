package be

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBE_ExamplePairs(t *testing.T) {
	require.Equal(t, []string{"Збольшага"}, NewSimpleReplaceRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Вялікая Айчынная вайна"}, NewBelarusianSpecificCaseRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
