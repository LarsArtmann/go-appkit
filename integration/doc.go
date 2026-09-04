// Package integration hosts cross-module and cross-repo end-to-end
// composition tests for go-appkit.
//
// The module exists so that E2E seams spanning multiple modules — or reaching
// into sibling repositories, such as cqrs-htmx's transport.JournalSSEStore →
// realtime replay bridge — can be tested without adding cross-module test
// dependencies to any library go.mod. It carries no production code and is
// never released; consumers never import it.
package integration
