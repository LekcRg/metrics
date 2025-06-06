package analyzers

import (
	"github.com/timakin/bodyclose/passes/bodyclose"
	"go.uber.org/nilaway"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/gofix"
	"golang.org/x/tools/go/analysis/passes/hostport"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/slog"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

// GetPassesAnalyzers возвращает набор стандартных анализаторов из пакета go/analysis/passes.
func GetPassesAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		appends.Analyzer,
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		defers.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		gofix.Analyzer,
		hostport.Analyzer,
		httpmux.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		slog.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stdversion.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
		waitgroup.Analyzer,
	}
}

// GetStaticAnalyzers возвращает анализаторы класса SA (staticcheck).
func GetStaticAnalyzers() []*analysis.Analyzer {
	checks := make([]*analysis.Analyzer, 0, len(staticcheck.Analyzers))
	for _, v := range staticcheck.Analyzers {
		checks = append(checks, v.Analyzer)
	}

	return checks
}

// GetStyleAnalyzers возвращает выбранные анализаторы класса ST (stylecheck).
func GetStyleAnalyzers() []*analysis.Analyzer {
	styles := map[string]bool{
		// ST1012 - Poorly chosen name for error variable.
		"ST1012": true,
		// ST1013 - Should use constants for HTTP error codes, not magic numbers.
		"ST1013": true,
		// ST1020 - The documentation of an exported function should start with the function’s name.
		"ST1020": true,
		// ST1022 - The documentation of an exported variable or constant should start with variable’s name
		"ST1022": true,
	}

	checks := make([]*analysis.Analyzer, 0, len(styles))
	for _, v := range stylecheck.Analyzers {
		if styles[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}

	return checks
}

// GetQuickfixAnalyzers возвращает выбранные анализаторы из класса QF (quickfix).
func GetQuickfixAnalyzers() []*analysis.Analyzer {
	// TODO: autofix?
	quick := map[string]bool{
		// QF1002 - Convert untagged switch to tagged switch
		"QF1002": true,
		// QF1003 - Convert if/else-if chain to tagged switch
		"QF1003": true,
	}

	checks := make([]*analysis.Analyzer, 0, len(quick))

	for _, v := range quickfix.Analyzers {
		if quick[v.Analyzer.Name] {
			checks = append(checks, v.Analyzer)
		}
	}

	return checks
}

// GetOtherAnalyzers возвращает остальные публичные анализаторы
func GetOtherAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		// checks whether HTTP response body is closed successfully
		bodyclose.Analyzer,
		// to detect potential nil panics in Go code
		nilaway.Analyzer,
	}
}
