# Op smoke table — 235 ops, route × bucket

**Generated from `op_smoke.tsv` by `tools/op_smoke.py --write-table`; do
not edit by hand.** Change a number by rerunning the script, not by editing this
document. The single source of truth is `op_smoke.tsv` (repo root) — that holds for
the **lrecon smoke** table only; REA probe tables are a separate measurement (below).

## Generation and verification

```sh
python3 tools/op_smoke.py                     # write op_smoke.tsv (incl. sha pin)
python3 tools/op_smoke.py --write-table       # generate this document from the tsv
python3 tools/op_smoke.py --markdown          # print the summary table to stdout
python3 tools/op_smoke.py --check             # verify sha pin + 235 coverage + evidence
python3 tools/op_smoke.py --live              # rerun the targeted smoke (235 × 20 s)
python3 tools/op_smoke.py --bucket-source rea # bucket source: probe20/final229.tsv
```

`--check` verifies three things: the `# lrecon-sha256` pin at the top of the table
matches the current `lrecon/lrecon`, all 235 ops appear exactly once, and the cells
total 235. On a pin mismatch it exits rc=1 with **stale table**; it never
silently carries old numbers forward.

## Two datasets — never mixed

| Dataset | What it measures | Role in this document |
|---|---|---|
| `op_smoke.tsv` | lrecon CLI, **targeted** 235 ops (target given, op argument not) | Single source; the route × bucket matrix comes from it |
| `probe20/final229.tsv` | REA MCP probe, retries **with arguments** (235 ops) | Coverage cross-check only; not a bucket source |
| `probe20/wave3.tsv` | REA probe intermediate wave (72 lines) | Cross-check only |

`final229.tsv` buckets (48 WORKING / 45 HONEST_REFUSAL / 97 ARG_VALIDATION / 38 NO_ENGINE / 1 TIMEOUT) do **not** measure what the lrecon smoke
buckets measure: one is the targeted CLI contract with no op argument, the other a
REA session replaying calls that were given an op argument and rejected. Mixing
them breaks the NEEDS_ARG / ARG_VALIDATION split.

## Sources

| Item | Source |
|---|---|
| op list (235) | `lrecon --list-ops` |
| routes 127 L / 69 Y / 39 G | `ops_probe.tsv`; otherwise `final.md` bullet lines |
| route (remaining ops) | already implemented in lrecon → **L**; no exception (`inspect_with_objdump` is pure Go) |
| bucket (default) | per-op smoke TSV: `--evidence` > `OP_SMOKE_EVIDENCE` > `smoke/op_smoke.tsv` > `probe20/op_smoke.tsv` |
| binary pin | `# lrecon-bin`, `# lrecon-sha256`, `# lrecon-bytes` (tsv header) |

## Buckets

| Bucket | Meaning |
|---|---|
| `WORKING` | rc=0, returned `{"result": …}` (data or empty result) |
| `NEEDS_ARG` | rc=2, op-specific argument required (`LRECON_AT is not set`, `pass <…>`, job field) |
| `ARG_VALIDATION` | rc=2, supplied argument invalid/incomplete (wrong shape, empty value) |
| `HONEST_REFUSAL` | rc=2, honest refusal (wrong shape, no cursor, unsupported target) |
| `NO_ENGINE` | rc=2, engine/tool not installed (gdb, frida, adb, jadx…) |
| `TIMEOUT` | timed out — **not folded into HONEST_REFUSAL**, its own bucket |

## Summary matrix

| Route | Op | WORKING | NEEDS_ARG | ARG_VALIDATION | HONEST_REFUSAL | NO_ENGINE | TIMEOUT |
|---|---|---|---|---|---|---|---|
| L (pure Go) | 127 | 27 | 6 | 8 | 86 | 0 | 0 |
| Y (external tool) | 69 | 9 | 0 | 1 | 59 | 0 | 0 |
| G (Ghidra) | 39 | 7 | 23 | 0 | 9 | 0 | 0 |
| **Total** | 235 | 43 | 29 | 9 | 154 | 0 | 0 |

Reading: of 235 ops, 127 are on the L route, 69 on Y,
39 on G. Only 43
(%18.3) return data on a targeted call.
The rest: NEEDS_ARG 29, HONEST_REFUSAL 154,
NO_ENGINE 0, TIMEOUT 0,
ARG_VALIDATION 9.

## Binary pin

`lrecon/lrecon` · sha256 `238a7be62d6d9c291a4d9b83cccd6a9b206610f49e270ea8b7de5e79e985c3d0` · 14953730 bytes. This pin sits at the top of `op_smoke.tsv`;
`--check` compares it against the current binary.

## `op_smoke.tsv`

Columns: `route`, `bucket`, `count`, `ops` (comma-separated op names). One line is one
cell; `count=0` cells are written too (`ops` = `-`) so the matrix stays complete and
diffable. Per-cell op lists and the verbatim tsv text live in a separate file:
`op-table-list.md` (written by the same generator, split out to keep this document
under the skill gate's 8192 B limit).
