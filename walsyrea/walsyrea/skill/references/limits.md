# Limits and refusals

`lrecon 1.0.0`. Every line here is a cap or a refusal, not a preference.

## Not supported

- source repos, live processes, pcap payload/carving, iOS without a device: name
  it, `VERDICT: unknown`, stop. This excludes repo/framework analysis only: a
  single Go source file still gets a digest (`kind:"text"`, `slices:0`,
  `bytes_scanned:0`, whole-file `sha256` and `file_size`) and nothing else — the
  string/byte ops refuse it with `is not ELF, Mach-O or PE` (see `router.md`).
- `out_of_scope` ops, refused by name: the frida session/instrument rows
  (`start_frida_session`, `close_frida_session`, `frida_session_status`,
  `resume_frida_session`, `load_frida_script`, `unload_frida_script`,
  `instrument_with_frida`, `trace_native_values`), `observe_native_calls`, the
  rizin console/debug rows (`execute_rizin_command`, `cutter_command`,
  `rizin_debug_command`, `rizin_debug_session_status`,
  `start_rizin_debug_session`, `close_rizin_debug_session`), `start_gdb_session`,
  and `observe_javascript_runtime` with a frida target.
- anything needing a disassembler or demangler: mangling families are labelled,
  never decoded.
- a signal no digest field names (`xxtea`, `jailbreak`): read the bytes
  yourself; `ptrace` in `imports` is the only antidebug hint.
- a language from markers alone: a hit is one mechanism, so a marker for zig,
  delphi, lua, evm or erlang (no mangling family) stays in `unknown`.
- Launching the target to defeat a layer is allowed (`references/bypass.md`
  runs it under `lldb`, `frida-trace -f`); attaching to a running process is not.

## Hard limits

| limit | value | enforced at |
|---|---|---|
| ghidra headless pass | 50 s + 12 s/MiB, 30 m cap (`LRECON_GHIDRA_TIMEOUT`, duration or seconds, wins); over budget → root `kind:"timeout"`, `stage:"import"`, prose `error`, rc=2; `LRECON_OP_DEADLINE` is the generic switch, used only when the former is empty; a cache hit never runs it | `lrecon/op_native_util.go:ghidraBudget`, `ghidraTimeoutForPath` |
| native string/pattern scan | first 64 MiB per target | `lrecon/op_native_util.go:48`, `:49` |
| native rows returned | 100; strings min length 4 | `lrecon/op_native_util.go:49`, `lrecon/op_list_strings.go:36` |
| one `--extract` member | 512 MiB (counted from bytes written) | `lrecon/extract.go:20`, `:255`, `:280` |
| `--extract-all` whole tree | 2 GiB (counted from bytes written) | `lrecon/extract.go:24`, `:126`, `:388` |
| whole-file static op read | 256 MiB, refused not cut | `lrecon/op_static_util.go:22`, `:37` |
| `--max-bytes` text budget | 4096 B default; `0` or negative = `bad_arg` rc=2, omit the flag for the default | `lrecon/main.go:72`, `:129-131` |
| `--max-bytes` with `--json`/`--op`/`--ext-op` | machine output is never cut: budget forced to 0 unless the flag is explicit, then 1 stderr note | `lrecon/main.go:311-320`, `:505-512` |
| in-line JSON answer | 10 MiB default; `LRECON_ANSWER_CAP_BYTES` only lowers it (0/negative/≥default ignored); past cap → file + `written_to` | `lrecon/op_native_util.go:386`, `:390-397` |
| headless runs per target | exactly 1 (file lock, 30 s wait then error) | `lrecon/op_native_cache.go:198-239`, `:26` |
| import cache key | target bytes (sha256) + Ghidra version, 16 hex; every G op on one binary shares it | `lrecon/op_native_cache.go:ghidraEntryKey` |
| import cache size | 20 entries, 7 days | `lrecon/op_native_cache.go:23-24` |
| headless runs per host | 4 (operational ceiling; only the one-per-target lock is in code) | `lrecon/op_native_cache.go:198-239` |
| language marker scan | first 16 MiB; one witness per language, one hit per rule | `lrecon/sigpack.go:27`, `:30`, `:190` |
| language marker pack | 35 fixed literals, schema 1; a rule with a weak atom is dropped at load | `lrecon/sigpack.go:24`, `:70-90` |
| language claim | 2 mechanisms (marker + mangling family) on disjoint ranges; 1 family short, no claim | `lrecon/sigpack.go:286-308`, `lrecon/claims.go:175` |
| asar header JSON | 64 MiB; 4096 member paths kept | `lrecon/asar.go:22`, `:23` |
| wasm whole-file read | 256 MiB, refused not cut; 64 names per digest | `lrecon/wasm.go:21`, `:18` |
| Go build-info scan | first 32 MiB; a blob past that window reads as absent | `lrecon/native.go:1012` |
| Erlang BEAM chunk read | 1 MiB per chunk; 256 atom names | `lrecon/beam.go:20`, `:19` |
| pcap packets | 2,000,000 per capture | `lrecon/pcap.go:14` |
| JS bundle strings | 256 reported per bundle | `lrecon/op_analyze_javascript_application.go:83` |
| GGUF/model metadata nesting | 4 array levels | `lrecon/op_inspect_model_artifact.go:265` |
| SQLite schema | 0 measured caps: 64-page b-tree depth guard, no row cap | `lrecon/op_inspect_sqlite_database.go:159` |
| Flutter pool uri/token | 0 counters: not measured, the Dart reader reads the 20-byte header only | `lrecon/dart.go:17-20` |

## Forbidden

- a wall-clock or random value inside `EVIDENCE` — offsets and bytes only.
- a clock- or random-derived artifact name: `$W/<name>.digest.json` is fixed
  (`skill/references/contract.md#stages`), so two runs produce the same names.
- a backgrounded probe racing another probe on the same target: order comes from
  the stage chain, not from a sleep.
- network calls of any kind.
- writing into the target you analyse.
- `of=/dev/` writes.
- unquoted `rm -rf $VAR`; quote it, or delete the literal `mktemp -d` path.
- printing secrets, keys, tokens or credentials.
- claiming `solved` without one run artifact.

## Failure | Fix

| failure | fix |
|---|---|
| `unknown op "foo" (235 registered; see --list-ops)` | `lrecon --list-ops` |
| `op "foo" needs a target path: --target <path>` (stderr) | `lrecon --op foo T`; when `needs` lists `LRECON_AT`/`LRECON_NAME`, that is the missing argument, not the target |
| `LRECON_GHIDRA=/bad is not an executable file` | `env -u LRECON_GHIDRA lrecon --op foo T` |
| `ghidra analyzeHeadless not installed: run from a Ghidra 12.x install` | `lrecon --json T` (digest needs no Ghidra) |
| `flag provided but not defined: -bogus` + `ipucu: lrecon --help` (rc=2, no envelope) | fix the flag name; `--help` for the list |
| `ghidra script "Foo.java" wrote unparsable JSON: unexpected end of JSON input` | `lrecon --op foo T` once, then report `partial` |
| `ghidra cache <dir>: lock still held after 30s` | `lrecon --op foo T` once more |
| `--extract: "m" is 734003200 bytes, over 512 MB extraction cap` | `unzip T m` |
| `--extract-all: over 2048 MB total cap` | `unzip T -d D 'dir/*'` |
| `lrecon: --max-bytes 40 does not apply to --json (753 bytes); drop the flag for machine output` | `lrecon --json T` |
| `lrecon: /path is a directory` | `lrecon /path/name` |
| a command exits non-zero and the reason is not obvious | print that stderr line, fix that one thing, stop — no second probe stacked on it |
| `rc=0` but the field you needed is missing | treat it as a failure, read stderr, do not re-run the same probe |
| the same probe was already run against the same bytes | it is the same evidence; use the first output and stop |

## Parallel

- One import per target: the cache lock (`lrecon/op_native_cache.go:198-239`)
  serialises import and `-process`. Never start a second import while it is
  held; retry after the 30 s wait.
- 4 headless runs per host is the ceiling: one per target is enforced, 4 total
  fits the 2 GiB default heap (`skill/references/ghidra.md:55-56`).
- Per call, `mktemp -d` the extract dir and delete that exact dir when done.
- One writer per project: a manual run needs a unique `projName`
  (`skill/references/ghidra.md:77-78`).
- Append-only ledger: `bash env/bench/check.sh metrics RUN_ID SESSION_ID WORKDIR`
  with a fresh `run_id`; a duplicate is refused (`env/bench/check.sh:80-81`).
- Read-only probes (digest, `--op list_strings`, `file`) may run concurrently;
  only import and analysis serialise.

## No reruns

One probe, one result; a re-run is not a check.

## Cost visibility

Say the cost before paying it; Ghidra is the only multi-minute step; run once.

## Done when

Contract audit (not per-run): `bash env/bench/check.sh --axes` checks 15 axes
fail-closed; needs `W` + `AXES_TARGET=<basename>`, rc!=0 = fail. `adhd_free` checks
`tools/adhd_free.sh` (no W): tracked + untracked tree, fail-closed if adhd shingles
appear even in an "adhd"-named file. Lint: `bash env/bench/check.sh`.

- **Identify**: `lrecon --json T` printed, `kind` named, one Route row run (or
  `#rowsuz-kinds`), then stop — no probe stack.
- **Unpack**: every layer is named in `LAYERS:` or `partial` names the unknown.
- **Recover**: a byte patch shows `EVIDENCE: 0x<offset>` plus before/after bytes;
  a literal `dd` patch needs no `-import`.
- **Prove**: one verification run, then stop; `solved` only with that artifact.
- A `solved`/`partial`/`unknown` verdict exits with no further probes; before a multi-minute op (Ghidra), state its cost and get an explicit go.
