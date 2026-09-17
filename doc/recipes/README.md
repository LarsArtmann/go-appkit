# Recipes

Durable, copy-paste rituals. Each recipe is verified against a PUBLISHED
tag before landing (doc snippets are code — AGENTS.md Gotchas).

| Recipe | Ritual |
| ------ | ------ |
| `fresh-consumer-proxy-check.md` | Prove a pushed tag resolves from the module proxy for a fresh consumer (`go get` exact version → blank import → build → optional behavioral probe). Last step of every release ritual. The `\_proxycheck/` script automates the fetch+build part for any module list. |
