# Ghidra script — handoff, output, reading order

Call shape and cost ladder: `ghidra.md`.

## Digest handoff

`-postScript DigestApply.java <digest.json>` on the same `-import` prints
`lang-ok`/`LANG MISMATCH`, rebases a PIE `-loader BinaryLoader` import to
0x100000, and prints `digest_applied=<field>=<value>` or `no-digest-field <name>`
for minos, uuid, build_id, subsystem, toolchain, pdb.

Containers have no loader: digest, import a native member, apply its digest.

| container | native member |
|---|---|
| asar | `--extract-all D` -> `native/*.node` |
| flutter | APK `lib/*/libapp.so` (ELF) |
| squashfs | `unsquashfs -d D` -> inner ELF |
| cramfs, cramfs-be | `cramfsck -x D T` -> inner ELF |
| ubi, ubifs | `ubireader_extract_images T`, then `ubireader_extract_files` -> inner ELF |
| jffs2 | `jefferson -d D` -> inner ELF |
| trx, uimage | `binwalk -e T` (or `dd` past the header) -> inner ELF |
| gzip, xz | `gunzip -c T` / `xz -dc T` -> inner member, then `kind` it |
| dex, hermes | bytecode: metadata route |
| wasm, sqlite, pcap, dtb | none: digest is the endpoint |

## What the script writes

- `extends GhidraScript`, `run()`, `currentProgram`.
- Output path from argv, not from an env var: `getScriptArgs()` gives
  `<op> <outFile> [args...]` (`lrecon/ghidra/scripts/NativeOp.java:97-103`); a
  relative path lands in the process CWD, not the project dir. The env-var form
  (`System.getenv("GHIDRA_OUTPUT_DIR")`, `.` fallback) belongs to the sim probe
  `env/sim300/ghidra/Probe.java:38-39`, not to the script lrecon runs.
- Loop hygiene: skip `func.isExternal()` and `func.isThunk()` bodies
  (`NativeOp.java:800`); only the sim probe honours `monitor.isCancelled()`
  (`Probe.java:61`).
- Emit JSON or CSV per artifact kind, one file each: summary, functions,
  decompiled bodies, strings, call edges, symbols. Parse after the JVM exits —
  do not scrape the console, which carries progress noise; `-log` holds the real
  log and the exit code lies often enough to check the log too.

Decompile-all only while the image is small: above roughly 5k functions or
50 MB, pivot — imports, strings, xrefs first, then decompile the handful of
functions whose xrefs actually reach the interesting string or table.

## Reading order

Recover types, then constants, then vtables, then signatures, and only then read
a decompiled body. Fingerprint known third-party libraries first and diff against
a pristine copy at the same version — the unique code is the only part worth
decompiling.

Decompiler output is not source-equivalent. Every claim about a branch, a length
check or a constant gets cross-checked against the disassembly and against the
bytes at the address the digest reported.

## Limits

- Static only. Debugger windows, tracing and live memory are tool-bound and
  cannot run through the headless entry point; use them from a GUI session.
- Managed bytecode (.NET, dex, Dart AOT) is not this tool's job — the digest's
  managed path names the metadata to read instead.
- Fat Mach-O: `lipo -info <fat>` lists the slices; thin with the arch it prints,
  `lipo -thin arm64e -output /tmp/thin <fat>`, then feed the thin file.
  `-thin arm64` fails on Apple system binaries, which carry only `arm64e`
  ("does not contain the specified architecture"). The digest says `arm64` where
  lipo says `arm64e` (and `x86-64` vs `x86_64`): map `slice_info` arch to the
  `lipo -info` name before thinning. Only thin arm64 was measured here.
