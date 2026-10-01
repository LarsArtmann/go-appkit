module github.com/larsartmann/go-appkit/systemd

go 1.27.1

require (
	github.com/coreos/go-systemd/v22 v22.7.0
	github.com/larsartmann/go-appkit v0.7.0
)

// DEV-ONLY replace: StartHooks does not exist in a published core tag yet.
// REMOVE this replace and bump the require above to the first core tag
// carrying StartHooks (v0.8.0) BEFORE tagging this module — pre-tag-checks.sh
// fails any tag whose go.mod still carries a filesystem replace.
replace github.com/larsartmann/go-appkit => ../
