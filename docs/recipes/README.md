# Recipes

Durable, copy-paste rituals. Each recipe is verified against a PUBLISHED
tag before landing (doc snippets are code — AGENTS.md Gotchas).

| Recipe                          | Ritual                                                                                                                                                                                                                                                                |
| ------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `fresh-consumer-proxy-check.md` | Prove a pushed tag resolves from the module proxy for a fresh consumer (`go get` exact version → blank import → build → optional behavioral probe). Last step of every release ritual. The `\_proxycheck/` script automates the fetch+build part for any module list. |
| `battery-swap-recipe.template.md` | The per-app adoption-wave checklist (SUPERB v12 A-train): inventory → battery map → risk-ranked swap order → LOC ledger (net MUST be negative) → suite green + counter delta. Copy per consumer repo; appkit ships the recipe, the consumer lands the swap. First filled-in instances: dnsblockd, ChastityAPI (A2/A3). |
