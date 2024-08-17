package scancheck

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: Name,
	Doc:  Description,
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

const (
	Name               = "scancheck"
	Description        = "Checks that bufio scanner errors are checked outside a Scan() loop"
	funcNameScan       = "Scan"
	funcNameErr        = "Err"
	funcNameNew        = "New"
	funcNameNewScanner = "NewScanner"
	pkgNameBufio       = "bufio"
	structNameScanner  = "Scanner"
	linterErrMessage   = "scanner.Err() called inside a Scan() loop"
)

func run(pass *analysis.Pass) (interface{}, error) {
	spector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{ // filter needed nodes: visit only them
		(*ast.ForStmt)(nil),
	}

	spector.Preorder(nodeFilter, func(node ast.Node) {
		forNode := node.(*ast.ForStmt)

		cond, ok := forNode.Cond.(*ast.CallExpr)
		if !ok {
			return
		}

		if !isCallToScannerFunc(cond, funcNameScan) {
			return
		}

		inspectBlockForFunc(pass, forNode.Body, funcNameErr)
		return
	})

	return nil, nil
}

func inspectBlockForFunc(pass *analysis.Pass, block *ast.BlockStmt, funcName string) {
	ast.Inspect(block, func(node ast.Node) bool {
		callExpr, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}

		if isCallToScannerFunc(callExpr, funcName) {
			pass.Reportf(callExpr.Pos(), linterErrMessage)
			return false
		}
		return true
	})
}

func isCallToScannerFunc(node *ast.CallExpr, funcName string) bool {
	selx, ok := node.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if selx.Sel.Name != funcName {
		return false
	}

	ident, ok := selx.X.(*ast.Ident)
	if !ok || ident.Obj == nil {
		return false
	}

	switch decl := ident.Obj.Decl.(type) {
	case *ast.AssignStmt:
		return isAssignmentScannerCreation(ident.Obj.Name, decl)
	case *ast.ValueSpec:
		for _, value := range decl.Values {
			lit, ok := value.(*ast.CompositeLit)
			if !ok {
				continue
			}
			if isSelxBufioScanner(lit.Type) {
				return true
			}
		}
	}
	return false
}

func isAssignmentScannerCreation(name string, assignStmt *ast.AssignStmt) bool {
	assignedIdx := -1
	for i, lhs := range assignStmt.Lhs {
		if ident, ok := lhs.(*ast.Ident); ok {
			if ident.Name == name {
				assignedIdx = i
				break
			}
		}
	}

	// we didn't find the thing we were looking for, or the LHS value is a multi-assignment from a single call
	// this is a false-negative scenario.
	if assignedIdx == -1 || assignedIdx >= len(assignStmt.Rhs) {
		return false
	}

	rhs, ok := assignStmt.Rhs[assignedIdx].(*ast.CallExpr)
	if !ok {
		return false
	}

	switch fun := rhs.Fun.(type) {
	case *ast.SelectorExpr: // call to bufio.NewScanner
		return isSelxBufioScanner(fun)
	case *ast.Ident: // call to new(bufio.Scanner)
		if fun.Name != funcNameNew {
			return false
		}
		// not possible: `new` builtin requires exactly 1 argument
		if len(rhs.Args) != 1 {
			return false
		}
		return isSelxBufioScanner(rhs.Args[0])
	default:
		return false
	}
}

func isSelxBufioScanner(node ast.Node) bool {
	selx, ok := node.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selx.X.(*ast.Ident)
	if !ok {
		return false
	}

	// handle shadowed bufio package
	if ident.Obj != nil {
		return false
	}

	if ident.Name != pkgNameBufio {
		return false
	}

	if selx.Sel.Name != funcNameNewScanner && selx.Sel.Name != structNameScanner {
		return false
	}

	return true
}
