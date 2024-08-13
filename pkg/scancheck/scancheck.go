package scancheck

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "scancheck",
	Doc:  "Checks that bufio scanner errors are checked outside a Scan() loop",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	spector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{ // filter needed nodes: visit only them
		(*ast.ForStmt)(nil),
	}

	spector.Preorder(nodeFilter, func(node ast.Node) {
		forNode, ok := node.(*ast.ForStmt)
		if !ok {
			panic("node filter for ast.ForStmt failed")
		}

		cond, ok := forNode.Cond.(*ast.CallExpr)
		if !ok {
			return
		}

		if !isCallToScannerScan(cond) {
			return
		}

		// now that we've found a for loop, recursively walk the entire AST within it to find if a scanner.Err() call is made
		ast.Inspect(forNode.Body, func(node ast.Node) bool {
			callExpr, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			if isCallToScannerErr(callExpr) {
				pass.Reportf(callExpr.Pos(), "scanner.Err() called inside a Scan() loop")
				return false
			}
			return true
		})
		return
	})

	return nil, nil
}

func isCallToScannerScan(node *ast.CallExpr) bool {
	selx, ok := node.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if selx.Sel.Name != "Scan" {
		return false
	}

	ident, ok := selx.X.(*ast.Ident)
	if !ok {
		return false
	}

	if ident.Obj == nil {
		return false
	}

	assignment, ok := ident.Obj.Decl.(*ast.AssignStmt)
	if !ok {
		return false
	}

	if !isAssignmentScannerCreation(assignment) {
		return false
	}
	return true
}

func isCallToScannerErr(node *ast.CallExpr) bool {
	selx, ok := node.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if selx.Sel.Name != "Err" {
		return false
	}

	ident, ok := selx.X.(*ast.Ident)
	if !ok {
		return false
	}

	if ident.Obj == nil {
		return false
	}

	assignment, ok := ident.Obj.Decl.(*ast.AssignStmt)
	if !ok {
		return false
	}

	if !isAssignmentScannerCreation(assignment) {
		return false
	}
	return true
}

func isAssignmentScannerCreation(assignStmt *ast.AssignStmt) bool {
	if len(assignStmt.Rhs) != 1 {
		return false
	}

	rhs, ok := assignStmt.Rhs[0].(*ast.CallExpr)
	if !ok {
		return false
	}

	selx, ok := rhs.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if !isName(selx.X, "bufio") {
		return false
	}

	if selx.Sel.Name != "NewScanner" {
		return false
	}

	return true
}

func isName(n ast.Node, name string) bool {
	if n, ok := n.(*ast.Ident); ok {
		return n.Name == name
	}
	return false
}
