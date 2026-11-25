# walsyrea contract (min)

`lrecon 1.0.0`. Three commands, in order (`--version` is identification, not setup):

```
lrecon --version
lrecon --json T
lrecon --op list_strings --json T
```

The first must print `lrecon 1.0.0`; the second gives `kind` and `plan`; the third is the only payload you need before choosing a route. Stop on exit 2 and quote the `{"op":…,"error":…,"kind":…}` envelope — there is no `needs` key in it: what the op wants is in the `error` text, and `needs:[…]` shows up only in `--list-ops --json` rows. `needs_arg` = the op needs `LRECON_AT`/`LRECON_NAME` (or another slot) besides a target; `bad_arg` = bad flag or value; `timeout` = budget spent.

## Scope

Shipped artifact, not a source repo or a live process. iOS without a device and `out_of_scope` ops are dropped (`limits.md#not-supported`).

## Digest

`kind`, `format`, `slices`, `deps`, `capabilities`, `hardening`, `skipped`, `plan`. `skipped:<reason>` is a refusal, not a finding. `assumed_target` + `confidence` appear only when the target was inferred (one file in cwd: 0.6; `LRECON_TARGET`: 0.9).

`plan` is one of `macho`, `elf`, `pe`, `mobile`, `generic`; run its probes in order (`references/router.md#digest-plan`).

## Stages

Four artifacts under `$W=$(mktemp -d)`, `<name>` = sanitized basename:
`T` → `$W/<name>.digest.json` → `$W/<name>.op-<op>.json` → `$W/<name>.proof.txt`.
Never extract into a non-empty dir.

## Env

`LRECON_TARGET` inferred target, `LRECON_AT` / `LRECON_NAME` address or symbol, `LRECON_GHIDRA_CACHE` (default `~/Library/Caches/lrecon-ghidra`), `LRECON_GHIDRA_TIMEOUT`, `LRECON_ANSWER_CAP_BYTES` (default 10485760 = 10 MiB).

## Exit gate

`KIND`, `LAYERS`, `VERDICT` (`solved` only with the proof defined in `contract.md#contract`: rc=0 + one of its two proofs), `EVIDENCE` (`0x<offset>: <bytes written>`, or `<in>`→`<out>` for a replay when `LAYERS: none`) filled; anything unnamed is `unknown`. One artifact per stage, Ghidra only after the cheap probes, no evidence committed.
