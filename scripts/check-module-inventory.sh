#!/usr/bin/env bash
# Module-inventory freshness guard (SUPERB v17 T11): the committed
# docs/MODULE_INVENTORY.md must match origin truth. Thin wrapper over
# scripts/module-inventory.sh --check (the venue-matrix note lives there):
# origin-truth design keeps it green through the tag→push window; the
# post-push re-run lands the updated row (the deferred-leg slot, alongside
# check-gosum-canonical on integration).

set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
exec "$root/scripts/module-inventory.sh" --check
