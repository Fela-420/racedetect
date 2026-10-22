	case *ast.DeferStmt:
		// Direct form: defer mu.Unlock()
		if sel, ok := x.Call.Fun.(*ast.SelectorExpr); ok {
			method := sel.Sel.Name
			if method == "Unlock" || method == "RUnlock" {
				defers = append(defers, deferSite{
					receiver: exprToString(sel.X),
					method:   method,
					pos:      x.Pos(),
				})
			}
		}
		// Closure form: defer func() { mu.Unlock(); ... }()
		if fl, ok := x.Call.Fun.(*ast.FuncLit); ok {
			ast.Inspect(fl.Body, func(m ast.Node) bool {
				es, ok := m.(*ast.ExprStmt)
				if !ok {
					return true
				}
				call, ok := es.X.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				if sel.Sel.Name == "Unlock" || sel.Sel.Name == "RUnlock" {
					defers = append(defers, deferSite{
						receiver: exprToString(sel.X),
						method:   sel.Sel.Name,
						pos:      x.Pos(),
					})
				}
				return true
			})
		}
		return false // don't descend into the deferred call's own args
