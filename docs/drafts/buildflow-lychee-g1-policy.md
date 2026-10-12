## TL;DR

The lychee preflight print-admits the authenticate-vs-exclude policy decision (survey g1) is still open, then warns on every run in every project that links private repos. Ask: decide g1 and implement it BuildFlow-side.

## Problem

`discovery/doctor/checks/lychee_fp_risk.go:38` (@ 8a349283a): "…the fleet authenticate-vs-exclude policy decision (survey g1) is still open." The check short-circuits when `GITHUB_TOKEN` is set (line 54) and otherwise emits the warning (line 70) telling the user to pick one of two fixes themselves.

Net effect: a permanent per-run warning for every private-repo project, plus a worse side effect — the warning noise buries real findings. In go-appkit's 2026-10-12 run, lychee reported 5 findings: 2 were genuinely dead links I only caught because I triaged the list manually; the other 3 were the known 404-on-private / unpushed-file / 418-teapot classes the warning predicts.

## Ask

Decide g1 and implement the winner in BuildFlow so the choice is not re-delegated to every repo:

- either inject `GITHUB_TOKEN` into the lychee step when the environment has one and exclude the LarsArtmann namespace when it does not,
- or default the namespace exclude and document the token path.

Either way the preflight warning goes away for repos that did nothing wrong.

💘 Generated with Crush
