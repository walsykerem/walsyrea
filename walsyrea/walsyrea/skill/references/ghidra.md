# Headless decompiler handoff

`lrecon` stops at the digest. When a `deep:` note says "read the code", this is
the one page needed to hand a binary to a static decompiler without a GUI.

`unsupported java version` means no installed JDK is in range (Ghidra 12.x asks
for 21+; 12.1.4 ran every G op on JDK 21.0.12.1 here). Install a newer JDK
only if the launcher refuses; `LaunchSupport` picks the JDK (`support/launch.sh:148`).

```sh
set -eu
D=$(dirname "$(readlink -f "$(command -v ghidraRun)")")  # cask: .../bin, distro: root
H=; for d in "$D/support" "$D/libexec/support" "$(dirname "$D")/libexec/support"; do
  if [ -x "$d/analyzeHeadless" ]; then H=$d/analyzeHeadless; break; fi
done
if [ -z "$H" ]; then echo "no analyzeHeadless under $D (layout, not java)"; exit 1; fi
"$H" 2>&1 | grep -m1 'Headless Analyzer Usage'
```

Python scripts are PyGhidra (3): `pyghidraRun` exists, `jythonRun` does not — write **Java**.

## Shape

```sh
GH=$H                          # preflight above
mkdir -p <projDir>             # a missing dir aborts the run
"$GH" <projDir> <projName> -import <bin> [-overwrite] [-processor <id>] \
  -scriptPath <dir> -postScript Foo.java [arg ...] \
  [-postScript Bar.java [arg ...]] -analysisTimeoutPerFile 300 [-deleteProject]
```

This block is Ghidra's generic call shape; `-max-cpu` and `-log` are **not flags
lrecon passes**. lrecon passes only `-import`, `-analysisTimeoutPerFile` (its own
budget, 50 s default), `-process <bin> -noanalysis` when needed, `-scriptPath`,
`-postScript` and `-deleteProject` (`lrecon/op_native_util.go:ghidraScriptRefusal`,
`:writeIfChanged`; meta path `lrecon/op_meta_util.go:579-591`).

- `projDir` and `projName` are the first two positional args, never flags: one
  positional arg alone aborts with `Bad argument: <path>`. The directory must
  already exist (`mkdir -p`) or the run aborts with "Directory not found". Keep
  them outside the sample directory: with `-import`/`-process`, a `projDir` path
  element starting with `.` aborts with "Path element starting with '.' is not
  permitted", and a `projName` starting with `.` with "Project name starting
  with '.' is not permitted". A dot in the `-import` path is fine; a bare run
  says "Nothing to do ... must specify -import, -process, or prescript and/or
  postscript".
- `-scriptPath` is mandatory: the loader only searches
  `$USER_HOME/ghidra_scripts` (when it exists) and the distro's own script dirs
  otherwise. A relative path works.
- Script args are positional after the script name, read with `getScriptArgs()`.
  Quote every arg; `&&`/`;` inside an unquoted arg splits the command silently.
- A `--deep` handoff prints `LRECON_NAME=<symbol-table name>`; Ghidra looks up
  its own function name, so `LRECON_NAME=main` dies on `no function named main`.
  Resolve it from `--op list_procedures`, or pass `LRECON_AT=0x…` (`lrecon.md#modes`).
- `GHIDRA_HEADLESS_MAXMEM=4G` in the environment raises the heap: the 2 GiB
  default heap is what you get without it, and OOM is the usual "randomly
  hangs" cause on a large image. `-analysisTimeoutPerFile` is a separate
  per-file wall-clock cap, not memory. `GHIDRA_MAXMEM` sets the heap for every
  Ghidra instance; the headless variable wins when both are set. `MAXMEM` is a
  trap: `launch.sh` reads it only as a positional argument, so exporting it
  changes nothing.

## Cost ladder

Cheapest first. Reach for the top rung only when there is no analyzed project.

1. `-process <name> -noanalysis -postScript` — rerun scripts on a project that is
   already analyzed. Seconds instead of minutes.
2. `-import -noanalysis` — import only: loader output, no analysis, so the
   function list stays short (5 of 7 functions on a small arm64 binary).
3. `-import` — full analysis, once.
4. One `-import` per query — never; the analysis is repaid every time.

One JVM, one import, N `-postScript` flags is the correct ordering: import →
analyze → export. Keep the project open; `-deleteProject` only for throwaway runs.

Parallel jobs need a **unique `projName`** (PID plus a suffix); a shared name
dies on the project lock.

Script, digest handoff, reading order: `ghidra-script.md`. Probe both layouts,
never hardcode a path.
