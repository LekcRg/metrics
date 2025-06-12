// Package main кастомный статический анализатор для го.
//
// Пример использования:
//   go build .
//   go vet -vettool=./staticlint ./...
//
// Included analyzers:
// - Standard go/analysis/passes
// - SA analyzers from staticcheck.io
// - Stylecheck: ST1012, ST1013, ST1020, ST1022
// - Quickfix: QF1002, QF1003
// - bodyclose, nilaway
// - Custom analyzer: noOsExitInMain (disallows os.Exit in main.main)

package main

import (
	"github.com/LekcRg/metrics/internal/analyzers"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {

	checks := append(analyzers.GetPassesAnalyzers(), analyzers.GetStaticAnalyzers()...)
	checks = append(checks, analyzers.GetStyleAnalyzers()...)
	checks = append(checks, analyzers.GetQuickfixAnalyzers()...)
	checks = append(checks, analyzers.GetOtherAnalyzers()...)
	checks = append(checks, analyzers.NoOSExitMainAnalyzer)
	// checks := []*analysis.Analyzer{analyzers.NoOSExitMainAnalyzer}

	multichecker.Main(checks...)
}
