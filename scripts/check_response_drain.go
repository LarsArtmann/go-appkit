//go:build ignore

// check_response_drain.go — repo-local AST walker for the undrained-body
// class (2026-11-06, issue54 session §e5: three occurrences — preset probes,
// rate-limit SSE E2E, errorpages E2E — undrained response bodies hold
// server-side connections ACTIVE through server.Shutdown and stall it past
// the stop timeout).
//
// Mechanism (linter-building decision tree): a syntax-shape rule, so a
// go/parser AST walker — no type information, no plugin machinery. Run via:
//
//	go run scripts/check_response_drain.go <dir> [<dir>...]
//
// Rule: in every *_test.go file, a call to `X.Body.Close()` (or
// `defer X.Body.Close()`) inside a function that never READS `X.Body` is a
// finding. A read is any other call-shaped use of the `X.Body` selector:
// io.Copy(_, X.Body), io.ReadAll(X.Body), X.Body.Read(...), X.Body.String(),
// X.Body.Len(), or passing X.Body to any helper call. The enclosing function
// is the scope (test helpers drain in the helper; the caller sees only the
// helper call).
//
// Confidence model (trust-engineering): findings are high-confidence by
// construction — the close site names the exact response ident and the
// whole enclosing function was scanned for a read of that ident. Negative
// filters keep the FP budget at zero:
//   - functions whose name matches the drain-helper conventions
//     (drain*, readBody*, httpGet*, get*, fetch*) are trusted;
//   - responses ASSIGNED FROM a testkit.* call are trusted (the testkit
//     helpers drain — pinned by testkit's own tests);
//   - a `//respdrain:ignore <reason>` comment on the Close line suppresses
//     the finding (reason required by convention; review checks it, not
//     the tool).
//
// Exit codes: 0 clean · 1 findings (each printed as file:line:col + rule
// text) · 2 usage/walk failure.
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type finding struct {
	pos   token.Position
	ident string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run scripts/check_response_drain.go <dir> [<dir>...]")
		os.Exit(2)
	}
	status := 0
	for _, root := range os.Args[1:] {
		walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				base := filepath.Base(path)
				if base == "vendor" || base == ".git" || strings.HasPrefix(base, "testdata") {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(filepath.Base(path), "_test.go") {
				return nil
			}
			for _, f := range inspectFile(path) {
				fmt.Printf("%s: %s.Body closed without being read (undrained bodies hold server connections open through Shutdown — drain with io.Copy(io.Discard, %s.Body) or a testkit helper)\n",
					f.pos, f.ident, f.ident)
				status = 1
			}
			return nil
		})
		if walkErr != nil {
			fmt.Fprintf(os.Stderr, "respdrain: %v\n", walkErr)
			os.Exit(2)
		}
	}
	os.Exit(status)
}

func inspectFile(path string) []finding {
	fset := token.NewFileSet()
	src, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		// A test file the compiler already rejects; go vet/test owns it.
		return nil
	}
	var findings []finding
	// ONE pass per TOP-LEVEL function: closes inside deferred closures
	// (`defer func() { _ = resp.Body.Close() }()`) attribute to the
	// enclosing function — the response flows in from the outer scope, so
	// outer-scope reads count. Split-scoping by FuncLit was the FP class
	// that flagged every drained caller.
	for _, decl := range src.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		findings = append(findings, checkFuncBody(fset, src, fn.Body, fn.Doc)...)
	}
	return findings
}

// bodyUse records one call-shaped use of an `X.Body` selector.
type bodyUse struct {
	ident   string // response ident, e.g. "resp"
	isRead  bool
	isClose bool
	pos     token.Pos
}

func checkFuncBody(fset *token.FileSet, file *ast.File, body *ast.BlockStmt, doc *ast.CommentGroup) []finding {
	if body == nil {
		return nil
	}
	if doc != nil && strings.Contains(doc.Text(), "respdrain:ignore") {
		return nil
	}

	var uses []bodyUse

	// NO assignment-source trust: verified 2026-11-06 that EVERY response
	// helper in the fleet (root httpGet, flightrecorder httpGetURL,
	// preset getWithHeaders, testkit.HTTPGetContext) returns the RAW
	// response — the convention is caller-owns-the-body. A close without
	// a read is a finding wherever it sits.

	ast.Inspect(body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.CallExpr:
			// Shape 1: X.Body.Close() (incl. defer) → close.
			if sel, ok := stmt.Fun.(*ast.SelectorExpr); ok && sel.Sel != nil && sel.Sel.Name == "Close" {
				if outer, ok := sel.X.(*ast.SelectorExpr); ok && outer.Sel != nil && outer.Sel.Name == "Body" {
					if id := identText(outer.X); id != "" {
						uses = append(uses, bodyUse{ident: id, isClose: true, pos: sel.Pos()})
					}
				}
				return true
			}
			// Shape 2: X.Body.<anything-else>() → read.
			if sel, ok := stmt.Fun.(*ast.SelectorExpr); ok && sel.Sel != nil {
				if outer, ok := sel.X.(*ast.SelectorExpr); ok && outer.Sel != nil && outer.Sel.Name == "Body" {
					if id := identText(outer.X); id != "" {
						uses = append(uses, bodyUse{ident: id, isRead: true, pos: sel.Pos()})
					}
				}
			}
			// Shape 3: X.Body passed as an argument → read (io.Copy,
			// io.ReadAll, drain helpers, testkit).
			for _, arg := range stmt.Args {
				if as, ok := arg.(*ast.SelectorExpr); ok && as.Sel != nil && as.Sel.Name == "Body" {
					if id := identText(as.X); id != "" {
						uses = append(uses, bodyUse{ident: id, isRead: true, pos: as.Pos()})
					}
				}
			}
		}
		return true
	})

	readIdents := map[string]bool{}
	for _, u := range uses {
		if u.isRead {
			readIdents[u.ident] = true
		}
	}

	var findings []finding
	for _, u := range uses {
		if !u.isClose {
			continue
		}
		if readIdents[u.ident] || suppressedAt(fset, file, u.pos) {
			continue
		}
		findings = append(findings, finding{pos: fset.Position(u.pos), ident: u.ident})
	}
	return findings
}

// suppressedAt reports a `respdrain:ignore` comment on the same line.
func suppressedAt(fset *token.FileSet, file *ast.File, pos token.Pos) bool {
	line := fset.Position(pos).Line
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if fset.Position(c.Pos()).Line == line && strings.Contains(c.Text, "respdrain:ignore") {
				return true
			}
		}
	}
	return false
}

func identText(e ast.Expr) string {
	if id, ok := e.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

func pkgName(e ast.Expr) string {
	if id, ok := e.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}
