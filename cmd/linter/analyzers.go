package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

var PanicAnalyzer = &analysis.Analyzer{
	Name: "nopanic",
	Doc:  "reports usage of the in-build panic function",
	Run:  checkPanicUsage,
}

var builtinPanic = types.Universe.Lookup("panic")

func checkPanicUsage(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			id, ok := call.Fun.(*ast.Ident)
			if ok && pass.TypesInfo.Uses[id] == builtinPanic {
				pass.Reportf(call.Pos(), "usage of panic function")
			}
			return true
		})
	}
	return nil, nil
}

var LogFatalAnalyzer = &analysis.Analyzer{
	Name: "nologfatal",
	Doc:  "reports log.Fatal calls outside the main function of the main package",
	Run:  checkLogFatalUsage,
}

func checkLogFatalUsage(pass *analysis.Pass) (any, error) {
	isMainPkg := pass.Pkg.Name() == "main"

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && isMainPkg && fn.Recv == nil && fn.Name.Name == "main" {
				continue
			}
			ast.Inspect(decl, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				fn, ok := pass.TypesInfo.Uses[sel.Sel].(*types.Func)
				if !ok || fn.Pkg() == nil || fn.Pkg().Path() != "log" {
					return true
				}
				switch fn.Name() {
				case "Fatal", "Fatalf", "Fatalln":
					pass.Reportf(call.Pos(), "log.%s outside the main function of the main package", fn.Name())
				}
				return true
			})
		}
	}
	return nil, nil
}

var OsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "reports os.Exit calls outside the main function of the main package",
	Run:  checkOsExitUsage,
}

func checkOsExitUsage(pass *analysis.Pass) (any, error) {
	isMainPkg := pass.Pkg.Name() == "main"

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && isMainPkg && fn.Recv == nil && fn.Name.Name == "main" {
				continue
			}
			ast.Inspect(decl, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				fn, ok := pass.TypesInfo.Uses[sel.Sel].(*types.Func)
				if !ok || fn.Pkg() == nil || fn.Pkg().Path() != "os" {
					return true
				}
				if fn.Name() == "Exit" {
					pass.Reportf(call.Pos(), "os.Exit outside the main function of the main package")
				}
				return true
			})
		}
	}
	return nil, nil
}
