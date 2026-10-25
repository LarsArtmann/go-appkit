// Standalone fresh-consumer proxy check (see ../fresh-consumer-proxy-check.md).
//
// Usage — run from anywhere, stdlib only, no workspace interference:
//
//	go run doc/recipes/_proxycheck/main.go \
//	    github.com/larsartmann/go-appkit@v0.5.1 \
//	    github.com/larsartmann/go-appkit/realtime@v0.1.1
//
// For each module@version argument it creates a throwaway module, fetches
// the EXACT version from the module proxy, blank-imports it, and builds.
// Exit 0 means every target passed as a fresh consumer would.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	targets := os.Args[1:]
	if len(targets) == 0 {
		fmt.Fprintln(os.Stderr, "usage: go run doc/recipes/_proxycheck/main.go <module>@<version> [...]")
		os.Exit(2)
	}

	failed := false
	for _, target := range targets {
		if err := check(target); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", target, err)
			failed = true

			continue
		}
		fmt.Fprintf(os.Stdout, "PASS %s\n", target)
	}

	if failed {
		os.Exit(1)
	}
}

func check(target string) error {
	module, version, found := strings.Cut(target, "@")
	if !found || module == "" || version == "" {
		return fmt.Errorf("target must be module@version, got %q", target)
	}

	dir, err := os.MkdirTemp("", "proxycheck-")
	if err != nil {
		return fmt.Errorf("scratch dir: %w", err)
	}
	defer os.RemoveAll(dir)

	goMod := "module proxycheck\n\ngo 1.26\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644); err != nil {
		return fmt.Errorf("go.mod: %w", err)
	}

	consumer := fmt.Sprintf("package main\n\nimport _ %q\n\nfunc main() {}\n", module)
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(consumer), 0o644); err != nil {
		return fmt.Errorf("main.go: %w", err)
	}

	env := append(os.Environ(), "GOWORK=off")
	for _, argv := range [][]string{
		{"go", "get", target},
		{"go", "build", "./..."},
	} {
		cmd := exec.Command(argv[0], argv[1:]...)
		cmd.Dir = dir
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", strings.Join(argv, " "), err)
		}
	}

	return nil
}
