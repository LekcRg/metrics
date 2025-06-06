package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var NoOSExitMainAnalyzer = &analysis.Analyzer{
	Name:     "noOsExitInMain",
	Doc:      "checks that os.Exit is not used directly in main() function of main package",
	Run:      NoOSExitMain,
	Requires: []*analysis.Analyzer{},
}

func NoOSExitMain(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if file.Name.Name != "main" {
			continue
		}

		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if funcDecl.Name.Name != "main" {
				continue
			}

			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				funcCall, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				method, ok := funcCall.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := method.X.(*ast.Ident)

				if !ok {
					return true
				}

				if ident.Name != "os" || method.Sel.Name != "Exit" {
					return true
				}

				message := "calling os.Exit in main.main is not allowed"
				pass.Report(analysis.Diagnostic{
					Pos:     funcCall.Pos(),
					End:     funcCall.End(),
					Message: message,
				})

				return true
			})
		}
	}

	return nil, nil
}
