package analyzers

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis"
)

func TestGetPassesAnalyzers(t *testing.T) {
	res := GetPassesAnalyzers()
	someItems := []string{
		"appends",
		"asmdecl",
		"assign",
		"atomic",
	}
	compare(t, res, someItems)
}

func TestGetStaticAnalyzers(t *testing.T) {
	res := GetStaticAnalyzers()
	someItems := []string{
		"SA1000",
		"SA1021",
		"SA4004",
		"SA5009",
	}
	compare(t, res, someItems)
}
func TestGetStyleAnalyzers(t *testing.T) {
	res := GetStyleAnalyzers()
	someItems := []string{
		"ST1012",
		"ST1013",
		"ST1020",
		"ST1022",
	}

	compare(t, res, someItems)
}
func TestGetQuickfixAnalyzers(t *testing.T) {
	res := GetQuickfixAnalyzers()
	someItems := []string{
		"QF1002",
		"QF1003",
	}

	compare(t, res, someItems)
}

func TestGetOtherAnalyzers(t *testing.T) {
	res := GetOtherAnalyzers()
	someItems := []string{
		"bodyclose",
		"nilaway",
	}

	compare(t, res, someItems)
}

func compare(t *testing.T, analysis []*analysis.Analyzer, list []string) {
	containsLength := 0
	for _, item := range analysis {
		if slices.Contains(list, item.Name) {
			containsLength++
		}
	}

	assert.Equal(t, containsLength, len(list))
}
