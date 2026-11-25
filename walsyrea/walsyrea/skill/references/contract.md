# Contract — goal, setup, stage order, runtime rules

`lrecon 1.0.0`. SKILL.md carries the gate; this file carries the rest.
Every rule here is a refusal or an order, not a preference.
Not authorized here: fabricated evidence, log-only proof, probe reruns as checks, permission-parser routes.

## Goal

One artifact in, one verdict out; a probe transcript is not the deliverable.
A repo, a live process, an iOS app without a device:
`references/limits.md#not-supported`, `VERDICT: unknown`, stop.

## Step 0: Setup

Step 0 is identification only (`lrecon --version` then one `lrecon --json T`); no setup or toolchain ceremony precedes it — setup only if the digest demands it; its first failure ends the run.

1. `lrecon --version` → `lrecon 1.0.0`. Anything else: `limits.md` and the Route
   table describe another build. Stop.
2. `lrecon --json T` → one digest. `kind` picks the Route row; no digest, no probe.
3. `W=$(mktemp -d)` per target. `--extract-all` does **not** guard its
   destination (non-empty dir: writes anyway, rc=`0`) — so never extract into one.
4. A failed digest is exit 2 with a JSON envelope on **stdout**
   (`{"op":null,"error":"stat /nope: no such file or directory","kind":"io"}`);
   quote it, stop. No second probe on an unexplained failure.

Defaults are explicit, never inferred: `contract_version: 1` in every digest,
`--max-bytes` 4096 B (>0; `0` = `bad_arg` rc=2, omit for the default), Ghidra
50 s (`LRECON_GHIDRA_TIMEOUT`).

## Env

Ref read order (`lrecon/op_native_util.go:nativeRef`, `:nativePair`):

| # | variable | rule |
|---|---|---|
| 1 | `LRECON_AT` | beats `LRECON_NAME` |
| 2 | `LRECON_NAME` | only when AT empty |
| 1 | `LRECON_FROM`+`LRECON_TO` | both required; pair ops only |
| — | `LRECON_ARGS`, `LRECON_QUERY`, `LRECON_PATTERN` | per-op slots |
| — | `LRECON_STATE`, `LRECON_SESSION` | every `metaStateRoot` consumer (`open_binary`, debug ops, bookmarks, `record_unknown`, obligation ledger); stateless ops ignore them |
| — | `LRECON_OP_DEADLINE` | js/web/extract budget; Ghidra fallback only if `LRECON_GHIDRA_TIMEOUT` empty |
| — | `LRECON_GHIDRA_LOCK_WAIT`, `LRECON_ANALYZEHEADLESS`, `LRECON_GHIDRA` | lock wait, headless path, Ghidra root. Nothing skips the G ops: `LRECON_SKIP_GHIDRA` is read only by `op_table_contract_test.go` |
| — | `LRECON_DECODERS`, `LRECON_JADX_THREADS`, `LRECON_APK_TIMEOUT` | apk/dex decoders, threads, budget |
| — | `LRECON_JOB`, `LRECON_FUZZ_OP`, `LRECON_BUNDLE_CAP_BYTES` | harness knobs; leave unset |

Budgets keep their default on a bad value; `LRECON_LEN` is `bad_arg`
(`limits.md`). Measured 2026-11-03.

## null vs ""

In the digest a string never emits `null`: `""` = no value for this kind. A
slice/object is `null` when its layer did not run, `[]` when it ran empty. A
count is never `null`; `0` covers zero and not-applicable. Measured 2026-11-03.

| field | empty | null | [] (ran, none) |
|---|---|---|---|
| `format`, `minos`, `uuid`, `soname`, `toolchain`, `mod` | `""` | — | — |
| `deps` | — | not scanned | ELF/PE `[]` |
| `hardening` | — | text `null` | — |
| `skipped`, `unknown` | — | fat `null` | ELF arrays |
| `claims`, `source_map` | — | zip `null` | thin Mach-O |
| `coverage` | — | text `null` | — |
| `slices` | `0` thin | — | — |

`slices` counts fat architectures in the digest (`0` = thin); `inspect_macho`
reports the one slice it read, so the same thin file is `slices:0` in the digest
and `1` in that op. Read a field with its source, never cross-compare.

`--op find_xrefs_to_name` uses `"function":null` outside a function. `jit` uses
state words (`not_detected`, `not_scanned`), not `""`. `jq '.x // "y"'` swallows
`false`/`""`; test with `has("x")`.

## Stages

This is the skill's own artifact-chain contract: fixed chain, in this order:
**Identify → Unpack → Recover → Prove**, one file per stage under `$W`:

```
T  ->  $W/<name>.digest.json      lrecon --json T
   ->  $W/<name>.op-<op>.json     lrecon --op <op> T      (one per probe)
   ->  $W/<name>.proof.txt        the single run that proves the claim
```

- `<name>` is the target basename with `[^A-Za-z0-9._-]` replaced by `-`
  (`sed 's|.*/||; s/[^A-Za-z0-9._-]/-/g'`). No directory, no space, no clock.
- Skip a stage only when its **input** is absent: no digest, no op; no op, no
  proof. "Not needed yet" is not absent.
- Recover never overwrites the digest: a patched copy lives in `$W/copy/` and
  gets its own digest.
- Naming is the contract: a later run tells what was measured and against which
  bytes, no log.

## Scope

- **Keep scope exact.** The ask names the artifact and the goal; anything else is
  a borrowed workflow, followed to the letter of what it states and no further.
  The file constrains the probe, never the target. Mechanically: the declared set IS the scope; anything outside it is residue that fails, not gets cleaned. BRIEF.md/specs.json/PLAN.md/src/outputs FOREIGN (not this skill's chain); native `$W` chain authoritative.
- Archive members stay in that artifact's `deps`, never their own reports.

## Creator editing boundary
Text the operator wrote by hand — a Route row, a `limits.md` value, a failure line — is theirs.
A change must name that line first and change only that line; rewriting a hand-edited section needs an explicit ask. Deleting is not cleaning up.

## Fan-out

Decide the report count before the first probe, and say it out loud:

1. Inventory cheaply — `lrecon --json` per candidate, `file -b` when it is not a
   known format. That inventory is the count.
2. **n artifacts → n digests → at most n reports**; only the artifacts the goal
   names go past the digest. `k probed` is stated with `n`, so a wrong count is
   visible before the work, not after.
3. Fan-out is only for independent artifacts; one artifact's layers stay
   sequential (`references/limits.md#parallel`).
4. Nothing is spawned for a single artifact. One artifact, one run, one verdict.

## Contract

- Exit gate is four fields: `KIND`, `LAYERS`, `VERDICT`, `EVIDENCE`.
- Digest keys are fixed at `contract_version: 1`. An absent key is not an empty
  value; `skipped:<reason>` is a refusal, not a zero.
- Packaged payload: `kind` stays the host, inventory in `deps`/`capabilities`.
- A claim carries `0x<offset>: <bytes written>`, `<in>`→`<out>` for a replay, or `unknown`.
  A clock, PID or random name in `EVIDENCE` is not evidence.
- **Refusal kinds: live vs table.** Live `kind` is the closed 9 of
  `lrecon/errors.go`: `needs_arg`, `bad_arg`, `out_of_scope`, `no_engine`,
  `timeout`, `usage`, `io`, `state_conflict`, `internal`. The table
  (`lrecon/testdata/ops.json`) adds classes the CLI never emits: `wrong_kind`
  (live is `bad_arg`), `skipped_ghidra`/`skipped_tool` (live `no_engine` when
  absent, `timeout`+stage on a deadline, `ok` when present), `not_applicable`
  (no target pass exists). Two live shapes: a Ghidra op on a wrong-kind file
  answers `no decompiler kind here` + `kind:no_engine` though Ghidra is
  installed; `--deep SYM` with an unknown symbol answers `kind:internal` —
  symbol-missing, not a tool fault. Map + pinning tests:
  `lrecon/op_table_test_helper_test.go` (`tableClassLiveKind`),
  `lrecon/op_table_contract_test.go` (`TestOpClassLiveKindMap`,
  `TestTableContractMapClassUse`). Never invent a runtime kind.
- Cheap before expensive: digest → symbol/string op → Ghidra
  (`references/ghidra.md#cost-ladder`). A cheap layer either answers or names the
  layer that needs the expensive one.
- `solved` needs the artifact running with exit 0 and one of two proofs: a diff
  exactly equal to the bytes written, or — a decode-only layer (constant/string
  table, KDF) or `LAYERS: none` (all layer rows probed empty) — one replayed
  input→output matching your prediction, decoy rejected. Else `partial` + blocker,
  or `unknown` + gap.

## Runtime rules

- **One failure, one sentence.** Read stderr, act on what it says, stop. Do not
  stack a second probe on a failure you cannot explain.
- **One probe, one result.** Re-running a probe is not a check. After the
  proving run, stop (re-lint after an edit is the exception).
- **Trust results, not logs.** `rc=0` plus the field you needed, or the stderr
  line. A probe that printed nothing usable did not succeed.
- **Say the cost before paying it.** The Ghidra pass is the only multi-minute
  step; announce it, run it once, never two at a time
  (`references/limits.md#parallel`).
- **Deterministic probes only**: no network, no wall clock, no backgrounded
  probe on one target; order comes from the stage chain.
- **Init guard**: never write into a non-empty directory, never into the target.
