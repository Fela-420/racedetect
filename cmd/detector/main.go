// Stage 2: CFG-based mutex reachability analysis.
//
// Rule: for every *.Lock()/*.RLock() call not covered by an immediate
// defer of the matching Unlock()/RUnlock(), walk the function's CFG
// from that call to every exit point. Any exit reached without first
// passing through the matching unlock call is reported as a violation.
//
// Each FuncDecl AND each FuncLit (closure) is analyzed as its own
// independent unit -- a return inside a closure exits the closure, not
// its enclosing function, so they cannot share one CFG.
//
// Usage:
//
//	go run ./cmd/detector <path-to-.go-file-or-dir>
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/cfg"
)

type lockSite struct {
	stmt         *ast.ExprStmt
	receiver     string
	method       string // "Lock" or "RLock"
	unlockMethod string // "Unlock" or "RUnlock"
	pos          token.Pos
}

type deferSite struct {
	receiver string
	method   string // "Unlock" or "RUnlock"
	pos      token.Pos
}

type funcUnit struct {
	name string
	body *ast.BlockStmt
}

// find is one problem found while walking forward from a single Lock
// site. kind is "missing-unlock" (a path reaches a function exit with
// no Unlock in between) or "double-lock" (a path reaches ANOTHER
// Lock/RLock on the same receiver before any Unlock in between --
// this deadlocks before the program ever gets near a later Unlock,
// so it is reported instead of, not in addition to, missing-unlock).
type find struct {
	kind          string
	line          int    // for missing-unlock: the exit line
	reacquireLine int    // for double-lock: the line of the SECOND Lock/RLock call
	message       string // for double-lock: a ready-to-print description
}

type labeledFind struct {
	unit       string
	lockLine   int
	lockMethod string
	find       find
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: detector <file-or-dir>")
		os.Exit(1)
	}

	files, err := collectGoFiles(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error collecting files: %v\n", err)
		os.Exit(1)
	}
	if len(files) == 0 {
		fmt.Println("no .go files found")
		return
	}

	totalViolations, totalCovered, totalClean := 0, 0, 0

	for _, f := range files {
		v, covered, clean, err := analyzeFile(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[PARSE ERROR] %s: %v\n", f, err)
			continue
		}
		totalViolations += v
		totalCovered += covered
		totalClean += clean
	}

	fmt.Printf("\n=== SUMMARY ===\n")
	fmt.Printf("violations: %d | covered-by-defer: %d | clean-explicit: %d\n",
		totalViolations, totalCovered, totalClean)
}

func collectGoFiles(root string) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	var out []string
	if !info.IsDir() {
		out = append(out, root)
		return out, nil
	}
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			out = append(out, path)
		}
		return nil
	})
	return out, err
}

func analyzeFile(path string) (violations, covered, clean int, err error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return 0, 0, 0, err
	}

	fmt.Printf("\n=== %s ===\n", path)

	units := collectFuncUnits(fset, node)
	if len(units) == 0 {
		fmt.Println("  no functions found")
		return 0, 0, 0, nil
	}

	anyOutput := false
	var allFinds []labeledFind
	for _, unit := range units {
		locks, defers := collectLockAndDeferSites(unit.body)

		for _, lock := range locks {
			deferCovers := false
			for _, d := range defers {
				if d.receiver == lock.receiver && d.method == lock.unlockMethod && d.pos > lock.pos {
					deferCovers = true
					break
				}
			}

			line := fset.Position(lock.pos).Line

			if deferCovers {
				fmt.Printf("  [OK-DEFER]   %-18s line %-4d %-7s -> covered by defer\n", unit.name, line, lock.method)
				covered++
				anyOutput = true
				continue
			}

			finds := checkReachability(fset, unit.body, lock)
			if len(finds) == 0 {
				fmt.Printf("  [OK]         %-18s line %-4d %-7s -> %s on every path\n", unit.name, line, lock.method, lock.unlockMethod)
				clean++
			}
			for _, f := range finds {
				allFinds = append(allFinds, labeledFind{unit: unit.name, lockLine: line, lockMethod: lock.method, find: f})
			}
			anyOutput = true
		}
	}

	// A lock call that was itself identified as "the second lock" in a
	// double-lock finding will, in reality, never execute its own
	// Unlock -- the program freezes AT that second Lock() call, so it
	// never gets the chance to reach anywhere past it. Reporting a
	// separate "missing unlock" for that same call site would just be
	// a confusing restatement of the same underlying bug. Suppress it.
	doubleLockLines := map[int]bool{}
	for _, lf := range allFinds {
		if lf.find.kind == "double-lock" {
			doubleLockLines[lf.find.reacquireLine] = true
		}
	}
	for _, lf := range allFinds {
		if lf.find.kind == "missing-unlock" && doubleLockLines[lf.lockLine] {
			continue // subsumed by the double-lock finding already reported for this same call site
		}
		switch lf.find.kind {
		case "double-lock":
			fmt.Printf("  [VIOLATION]  %-18s line %-4d %-7s -> %s\n", lf.unit, lf.lockLine, lf.lockMethod, lf.find.message)
		case "missing-unlock":
			fmt.Printf("  [VIOLATION]  %-18s line %-4d %-7s -> exit at line %d reached WITHOUT unlock\n", lf.unit, lf.lockLine, lf.lockMethod, lf.find.line)
		}
		violations++
	}

	if !anyOutput {
		fmt.Println("  no lock calls found")
	}

	return violations, covered, clean, nil
}

// collectFuncUnits finds every FuncDecl and FuncLit in the file, each
// treated as an independent CFG unit.
func collectFuncUnits(fset *token.FileSet, file *ast.File) []funcUnit {
	var units []funcUnit
	ast.Inspect(file, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Body != nil {
				units = append(units, funcUnit{name: x.Name.Name, body: x.Body})
			}
		case *ast.FuncLit:
			line := fset.Position(x.Pos()).Line
			units = append(units, funcUnit{name: fmt.Sprintf("closure@L%d", line), body: x.Body})
		}
		return true
	})
	return units
}

// collectLockAndDeferSites scans a single unit's body for Lock/RLock
// call sites and Unlock/RUnlock defer sites, WITHOUT descending into
// any nested FuncLit (that's a separate unit, analyzed on its own).
func collectLockAndDeferSites(body *ast.BlockStmt) ([]lockSite, []deferSite) {
	var locks []lockSite
	var defers []deferSite

	var visit func(n ast.Node) bool
	visit = func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncLit:
			return false // nested closure is its own unit; don't descend

		case *ast.ExprStmt:
			if call, ok := x.X.(*ast.CallExpr); ok {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					method := sel.Sel.Name
					if method == "Lock" || method == "RLock" {
						unlockMethod := "Unlock"
						if method == "RLock" {
							unlockMethod = "RUnlock"
						}
						locks = append(locks, lockSite{
							stmt:         x,
							receiver:     exprToString(sel.X),
							method:       method,
							unlockMethod: unlockMethod,
							pos:          x.Pos(),
						})
					}
				}
			}

		case *ast.DeferStmt:
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
			return false // don't descend into the deferred call's own args
		}
		return true
	}
	ast.Inspect(body, visit)
	return locks, defers
}

func exprToString(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return exprToString(x.X) + "." + x.Sel.Name
	case *ast.StarExpr:
		return "*" + exprToString(x.X)
	default:
		return "?"
	}
}

// checkReachability builds the CFG for `body`, locates the block
// containing lock.stmt, and walks forward through all successors.
// Each step of each path asks, in order:
//  1. Does this node release the lock (matching Unlock/RUnlock)?
//     -> path is safe, stop walking it.
//  2. Does this node ACQUIRE the lock again (matching Lock/RLock),
//     before any release was seen on this path?
//     -> double-lock: the program deadlocks right here, report it,
//        and stop walking this path (nothing past this point can
//        ever actually execute).
//  3. Neither, and this is a function exit with no successors?
//     -> missing-unlock: this path genuinely returns without ever
//        releasing the lock.
//
// This ordering is what lets the walk correctly follow a loop
// back-edge into the SAME Lock() statement it started from (as in a
// `for { Lock(); ...; continue }` bug) -- re-scanning that block from
// its start finds the Lock() call again before the visited-check
// would otherwise silently stop the walk.
func checkReachability(fset *token.FileSet, body *ast.BlockStmt, lock lockSite) []find {
	mayReturn := func(call *ast.CallExpr) bool { return true } // v1: no special-casing of panic/os.Exit
	g := cfg.New(body, mayReturn)

	startBlock, startIdx := findContainingBlock(g, lock.stmt)
	if startBlock == nil {
		return nil // couldn't locate -- don't false-positive on a tool gap
	}

	var finds []find
	seenExit := map[int]bool{}
	seenDoubleLock := map[int]bool{}
	visited := map[*cfg.Block]bool{}

	var walk func(b *cfg.Block, fromIdx int)
	walk = func(b *cfg.Block, fromIdx int) {
		for i := fromIdx; i < len(b.Nodes); i++ {
			n := b.Nodes[i]

			if isMatchingUnlock(n, lock) {
				return // released -- this path is safe, stop walking it
			}

			if isMatchingAcquire(n, lock) {
				line := fset.Position(n.Pos()).Line
				if seenDoubleLock[line] {
					return // already reported this exact re-acquisition point
				}
				seenDoubleLock[line] = true

				msg := fmt.Sprintf("re-acquires %s at line %d before any %s -- deadlocks here, never reaches a later %s",
					lock.method, line, lock.unlockMethod, lock.unlockMethod)
				if n == ast.Node(lock.stmt) {
					msg = fmt.Sprintf("loop re-enters this same %s call (line %d) before any %s -- deadlocks here",
						lock.method, line, lock.unlockMethod)
				}
				finds = append(finds, find{kind: "double-lock", reacquireLine: line, message: msg})
				return // this path is now dead in practice -- stop walking it
			}
		}

		if visited[b] {
			return
		}
		visited[b] = true

		if len(b.Succs) == 0 {
			el := exitLine(fset, b)
			if !seenExit[el] {
				seenExit[el] = true
				finds = append(finds, find{kind: "missing-unlock", line: el})
			}
			return
		}
		for _, succ := range b.Succs {
			walk(succ, 0)
		}
	}
	walk(startBlock, startIdx+1)

	return finds
}

func findContainingBlock(g *cfg.CFG, target *ast.ExprStmt) (*cfg.Block, int) {
	for _, b := range g.Blocks {
		for i, n := range b.Nodes {
			if n == ast.Node(target) {
				return b, i
			}
		}
	}
	return nil, -1
}

func isMatchingUnlock(n ast.Node, lock lockSite) bool {
	stmt, ok := n.(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := stmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == lock.unlockMethod && exprToString(sel.X) == lock.receiver
}

// isMatchingAcquire reports whether n is a Lock/RLock call on the
// same receiver as `lock` -- i.e. the lock being grabbed AGAIN.
func isMatchingAcquire(n ast.Node, lock lockSite) bool {
	stmt, ok := n.(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := stmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	method := sel.Sel.Name
	return (method == "Lock" || method == "RLock") && exprToString(sel.X) == lock.receiver
}

func exitLine(fset *token.FileSet, b *cfg.Block) int {
	if ret := b.Return(); ret != nil {
		return fset.Position(ret.Pos()).Line
	}
	if len(b.Nodes) > 0 {
		return fset.Position(b.Nodes[len(b.Nodes)-1].End()).Line
	}
	return 0
}

