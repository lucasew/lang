package bigdata

// Twin of ConfusionFileIndenterTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of ConfusionFileIndenterTest.indentWithCommentsTest
func TestConfusionFileIndenter_IndentWithCommentsTest(t *testing.T) {
	lines := []string{
		"mir; mit; 1.50 # p=0.994, r=0.658, tp=775, tn=1173, fp=5, fn=403, 178+1000, 2017-10-23",
		"nach; noch; 1.75 # p=0.990, r=0.504, tp=1009, tn=1990, fp=10, fn=991, 1000+1000, 2017-10-23",
	}
	result := IndentConfusionFile(lines)
	expected := "mir; mit; 1.50                                                                    # p=0.994, r=0.658, tp=775, tn=1173, fp=5, fn=403, 178+1000, 2017-10-23\n" +
		"nach; noch; 1.75                                                                  # p=0.990, r=0.504, tp=1009, tn=1990, fp=10, fn=991, 1000+1000, 2017-10-23\n"
	require.Equal(t, expected, result)
}

// Port of ConfusionFileIndenterTest.indentWithoutCommentsTest
func TestConfusionFileIndenter_IndentWithoutCommentsTest(t *testing.T) {
	lines := []string{"mir; mit; 1.50", "nach; noch; 1.75"}
	result := IndentConfusionFile(lines)
	expected := "mir; mit; 1.50\nnach; noch; 1.75\n"
	require.Equal(t, expected, result)
}

// Port of ConfusionFileIndenterTest.indentCommentedLineTest
func TestConfusionFileIndenter_IndentCommentedLineTest(t *testing.T) {
	lines := []string{
		"#mir; mit; 1.50 # p=0.994, r=0.658, tp=775, tn=1173, fp=5, fn=403, 178+1000, 2017-10-23",
		"nach; noch; 1.75# p=0.990, r=0.504, tp=1009, tn=1990, fp=10, fn=991, 1000+1000, 2017-10-23",
	}
	result := IndentConfusionFile(lines)
	expected := "#mir; mit; 1.50                                                                   # p=0.994, r=0.658, tp=775, tn=1173, fp=5, fn=403, 178+1000, 2017-10-23\n" +
		"nach; noch; 1.75                                                                  # p=0.990, r=0.504, tp=1009, tn=1990, fp=10, fn=991, 1000+1000, 2017-10-23\n"
	require.Equal(t, expected, result)
}

// Port of ConfusionFileIndenterTest.indentLongLineTest
func TestConfusionFileIndenter_IndentLongLineTest(t *testing.T) {
	lines := []string{
		"fielen|wie in 'Die Kinder fielen hin.'; vielen|wie in 'Wir helfen vielen Menschen.'; 0.50 # p=0.994, r=0.715, tp=805, tn=1121, fp=5, fn=321, 126+1000, 2017-09-24\n",
	}
	result := IndentConfusionFile(lines)
	expected := "fielen|wie in 'Die Kinder fielen hin.'; vielen|wie in 'Wir helfen vielen Menschen.'; 0.50 # p=0.994, r=0.715, tp=805, tn=1121, fp=5, fn=321, 126+1000, 2017-09-24\n\n"
	require.Equal(t, expected, result)
}
