package linter

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "urlcheck",
	Doc:      "checks that Url fields use Getter type and GetterRoutePath",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func containsUrl(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "url")
}

func isGetterType(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Name() == "Getter" && strings.HasSuffix(obj.Pkg().Path(), "getters")
}

func isGetterRoutePathCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	// Check for lago.GetterRoutePath(...)
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == "GetterRoutePath"
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CompositeLit)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		lit := n.(*ast.CompositeLit)

		litType := pass.TypesInfo.TypeOf(lit)
		if litType == nil {
			return
		}

		// Unwrap pointer
		if ptr, ok := litType.(*types.Pointer); ok {
			litType = ptr.Elem()
		}

		st, ok := litType.Underlying().(*types.Struct)
		if !ok {
			return
		}

		for _, elt := range lit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			ident, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}

			if !containsUrl(ident.Name) {
				continue
			}

			// Find field type in the struct
			var fieldType types.Type
			for field := range st.Fields() {
				if field.Name() == ident.Name {
					fieldType = field.Type()
					break
				}
			}
			if fieldType == nil {
				continue
			}

			if !isGetterType(fieldType) {
				pass.Reportf(kv.Pos(), "field %q contains 'Url' but its type is not Getter; consider changing the field type to getters.Getter", ident.Name)
			} else if !isGetterRoutePathCall(kv.Value) {
				pass.Reportf(kv.Value.Pos(), "field %q is a Url Getter; consider using lago.GetterRoutePath() instead", ident.Name)
			}
		}
	})

	return nil, nil
}
