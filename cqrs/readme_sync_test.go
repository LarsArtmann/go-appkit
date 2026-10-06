package cqrs

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"regexp"
	"strings"
	"testing"
)

var readmeQualifiedID = regexp.MustCompile(`\b([a-z][a-z0-9]*)\.([A-Z][A-Za-z0-9_]*)`)

var readmeExampleRef = regexp.MustCompile(`\bExample[A-Z]\w*`)

var readmeExampleFunc = regexp.MustCompile(`(?m)^func (Example[A-Z]\w*)\(`)

var readmeExercisedPrefixes = map[string]bool{
	"cl":             true,
	"clprojections":  true,
	"command":        true,
	"event":          true,
	"id":             true,
	"metaengine":     true,
	"projectionhost": true,
	"system":         true,
}

func TestReadmeRecipes_StayInSyncWithModuleCode(t *testing.T) {
	t.Parallel()

	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}

	api := exportedAPINames(t)
	moduleCode := allGoSources(t)

	for _, match := range readmeQualifiedID.FindAllStringSubmatch(string(readme), -1) {
		prefix, name := match[1], match[2]

		switch {
		case prefix == "cqrs":
			if !api[name] {
				t.Errorf("README references cqrs.%s, which the module does not export; "+
					"the recipe drifted from the API or invents one", name)
			}
		case readmeExercisedPrefixes[prefix]:
			if !strings.Contains(moduleCode, prefix+"."+name) {
				t.Errorf("README references %s.%s, which appears nowhere in module code; "+
					"the recipe drifted from the real API or no Example exercises it", prefix, name)
			}
		}
	}
}

func TestReadmeRecipeExamples_AreLinkedBothWays(t *testing.T) {
	t.Parallel()

	readme, err := os.ReadFile("README.md")
	if err != nil {
		t.Fatalf("read README.md: %v", err)
	}

	examples, err := os.ReadFile("example_test.go")
	if err != nil {
		t.Fatalf("read example_test.go: %v", err)
	}

	exampleNames := map[string]bool{}
	for _, match := range readmeExampleFunc.FindAllStringSubmatch(string(examples), -1) {
		exampleNames[match[1]] = true
	}

	mentioned := map[string]bool{}
	for _, name := range readmeExampleRef.FindAllString(string(readme), -1) {
		mentioned[name] = true

		if !exampleNames[name] {
			t.Errorf("README references %s, which example_test.go does not define", name)
		}
	}

	for name := range exampleNames {
		if !mentioned[name] {
			t.Errorf("example %s is referenced nowhere in README.md; "+
				"document it next to its recipe so both sides stay in sync", name)
		}
	}
}

func exportedAPINames(t *testing.T) map[string]bool {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read module directory: %v", err)
	}

	names := map[string]bool{}
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, fileName, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", fileName, err)
		}

		for _, decl := range file.Decls {
			collectExportedNames(decl, names)
		}
	}

	return names
}

func collectExportedNames(decl ast.Decl, names map[string]bool) {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		if d.Recv == nil && d.Name.IsExported() {
			names[d.Name.Name] = true
		}
	case *ast.GenDecl:
		for _, spec := range d.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				if s.Name.IsExported() {
					names[s.Name.Name] = true
				}
			case *ast.ValueSpec:
				for _, ident := range s.Names {
					if ident.IsExported() {
						names[ident.Name] = true
					}
				}
			}
		}
	}
}

func allGoSources(t *testing.T) string {
	t.Helper()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read module directory: %v", err)
	}

	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") {
			continue
		}

		data, err := os.ReadFile(fileName)
		if err != nil {
			t.Fatalf("read %s: %v", fileName, err)
		}

		parts = append(parts, string(data))
	}

	return strings.Join(parts, "\n")
}
