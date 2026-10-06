module github.com/larsartmann/go-appkit/systemd

go 1.27.1

require (
	github.com/coreos/go-systemd/v22 v22.7.0
	github.com/larsartmann/go-appkit v0.7.0
	github.com/larsartmann/go-error-family v0.11.0
)

require (
	github.com/aymanbagabas/go-osc52/v2 v2.0.1 // indirect
	github.com/charmbracelet/colorprofile v0.4.3 // indirect
	github.com/charmbracelet/lipgloss v1.1.0 // indirect
	github.com/charmbracelet/log v1.0.0 // indirect
	github.com/charmbracelet/x/ansi v0.11.8 // indirect
	github.com/charmbracelet/x/cellbuf v0.0.15 // indirect
	github.com/charmbracelet/x/term v0.2.2 // indirect
	github.com/clipperhouse/displaywidth v0.11.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/go-logfmt/logfmt v0.6.1 // indirect
	github.com/justinas/nosurf v1.2.0 // indirect
	github.com/larsartmann/go-etag/entitytag v0.6.1 // indirect
	github.com/larsartmann/go-etag/server v0.6.1 // indirect
	github.com/larsartmann/httputil v1.4.1 // indirect
	github.com/lucasb-eyer/go-colorful v1.4.1 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/mattn/go-runewidth v0.0.30 // indirect
	github.com/muesli/termenv v0.16.0 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	github.com/xo/terminfo v1.2.0 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/exp v0.0.0-20261005173118-76772065c9b0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/time v0.16.0 // indirect
)

// DEV-ONLY replace: StartHooks does not exist in a published core tag yet.
// REMOVE this replace and bump the require above to the first core tag
// carrying StartHooks (v0.8.0) BEFORE tagging this module — pre-tag-checks.sh
// fails any tag whose go.mod still carries a filesystem replace.
replace github.com/larsartmann/go-appkit => ../
