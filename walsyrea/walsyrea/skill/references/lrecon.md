# lrecon 1.0.0 — modes, digest reading, failure

`lrecon <target>` = text digest (default 4096-byte budget), `--json` = the same
keys machine-readable, never truncated. One target per call.

| exit | meaning | measured cases |
|---|---|---|
| 0 | result printed | digest, `no differences`, `--version`, `--explain` |
| 1 | `--diff` found differences — the only 1 | two builds of one product; stdout is the field digest, not an error |
| 2 | everything else that did not answer | bad path; `--deep` bad symbol / stripped target / container; `--diff` unreadable OLD; `--extract`/`--extract-all` bad member or non-container; usage; declined target |

A `--json` refusal prints its envelope on stdout too: `ops.md#failure`.

## Modes

| when | flag | returns |
|---|---|---|
| routing | `--json T` | digest keys, never truncated |
| digest as a file | `-o FILE T` | the same JSON written to FILE; `-o` does not apply to `--op` (rc=2 `bad_arg`, result goes to stdout) |
| one field's source | `--explain FIELD T` | `field: <name>` + `attribution: <source @offset>` or `none`; unknown field = `lrecon: unknown field …` on stderr, rc=2, no envelope (with `--json`: `bad_arg`). `--help --all` names fields `--explain` rejects (`--explain claim` = rc=2). `attribution: none` means the digest emitted no claim/skip for that field, not that the field is unknown — `entrypoint` prints a value and no attribution |
| layer named or `notes` says `deep:` (Mach-O/ELF/PE) | `--deep SYM T` | symbol table read in-process (`debug/macho\|elf\|pe`). Go: `go tool objdump`, 200 lines + `truncated to N of M`. Other: symbol line + `deep:` line naming the Ghidra op (`LRECON_NAME=SYM lrecon --op procedure_assembly T`); `SYM` is Ghidra's function name, not the symbol-table name (`_main` is `entry`), so the printed handoff dies on `no function named main` — resolve from `--op list_procedures T` or pass `LRECON_AT=0x…` (beats `LRECON_NAME`); `ghidra.md#shape`. Container (ipa/apk/zip/asar) = rc=2 `not Mach-O, ELF or PE`: extract, then `--deep` the member. Stripped/unknown symbol = stderr + rc=2; no symbol-table entry → pclntab fallback hex dump + handoff; data symbol prints only its line |
| two builds of one product | `--diff OLD T` | one line per changed field; `no differences` = nothing moved, rc=0; any change = rc=1 |
| payload nested (asar/zip/apk/ipa) | `--extract-all DIR T` | writes members as it walks, relative paths, 2 GB total; non-container = rc=2 `neither a zip/apk/ipa nor an asar`; ipa Mach-O at `Payload/<app>.app/<app>` |
| one member | `--extract M T -o DIR` | one member into DIR; `-o` mandatory (rc=2 `needs -o DIR`), never cwd; 512 MB per member, same failure contract |
| piping | `--json` | plain text, no color flag exists |
| what next | `--next T` | three `lrecon --op … T` lines from a fixed per-kind list; on a plain Mach-O they are generic (Go ops included) — hints, not the route |

The text digest's `evidence:` line is a hash over the evidence set, not the file
hash (`shasum -a 256 T` and `--explain file_sha256` both disagree with it) — quote
`target.file_sha256` from `--json` when the artifact hash is what you need.

| flag | state | behaviour |
|---|---|---|
| `--version` | implemented | prints `lrecon 1.0.0`, exit 0, target ignored |
| `--json`, `-o`, `--explain`, `--max-bytes`, `--deep`, `--diff`, `--extract`, `--extract-all` | implemented | see above; `--json` refuses to combine with exactly five: `--diff`, `--extract`, `--extract-all`, `--deep`, `--next` (rc=2 `usage`); `--json` accepts `--explain` (rc=0), `-o` (rc=0, writes the file), `--max-bytes` (rc=0, stderr note) |
| `--pid N`, `--process`, `--runtime`, `--live`, `--attach` | reserved | parsed, then declined: `lrecon: live runtime targets are out of scope: lrecon reads static files only`, exit 2, no attach happens |

## Reading digest

Fields that change the next move:

- `note:` lines (text digest; `--json` carries none) — `deep: …` names the
  next read; a gap is `truncated`, `capped` or
  `native: EOF` (file ended inside a table/body: a cut ELF is `kind: elf` +
  `note: native: EOF`). A gap is not a result.
- `slices`, `slice_info` — >1 arch: thin one slice before disassembling, or offsets are wrong.
- `hardening` — `codesign` (the signature check, `sigcheck`) means a patch must be re-signed; `relro-partial` = no `DT_BIND_NOW`, so lazy binding survives; absent ≠ disabled.
- `constants`, `format`, `toolchain`, `capabilities` — `crc32_poly`/`crc32_table` = crc32, `chacha_sigma` = chacha, `aes_sbox` = AES, not `aes-gcm`; `compress` = gzip/bzip2/xz/lzma/lz4/zstd, a bare zlib inflate has no header; `toolchain`/`mod` = Go `go buildinfo`; `mangling:<family>` (rust mangling `rust-legacy`/`rust-v0`) = a label, not a name; `inspect_signature` = the `Entitlements` slots.
- `imports`, `dylibs` — `ptrace`, `dlsym`, `mmap`, `mprotect`: `bypass.md#layers`.
- `capabilities`, `asar_files` — Electron: header count only, unpack before code.
- `coverage` (`parsed/skipped/unknown`) over `bytes_scanned` of `file_size`. A
  `skipped` entry is `reason: detail` from a closed set. Low `parsed` on a big
  file is an unfinished read; `truncated` cuts the digest, not the file. One
  unit = one record a reader attributed (symbol, chunk, partition, language
  hit); `parsed+skipped+unknown` = that count, not `nm` totals. `unknown[]` is a separate axis (slots no claim settled), never added
  to the sum.
- `claims` — `value (source @offset)`: an attribution with its evidence. A language
  or producer claim needs two independent sources; one witness is a hint.

Size is no reason to skip: `--json` is never cut.

## New kinds

`pyc` and `php-opcache` are magic readers that set `kind`: a pyc's magic is its
only version field (unmapped → `pyc-version-unverified`), an opcache file is a
header plus pools (`opcache:system_id|mem_size|str_size` in `capabilities`,
refusals `opcache-truncated`, `opcache-header-unverified`). `phar` claims `kind`
only when nothing else did, so a phar in a zip stays `zip`; a tar/zip-based phar
is a refusal (`phar-tar-based`, `phar-zip-based`). `swift`, `pkg`, `bun`,
`deno-compile`, `node-sea`, `klib`, `.NET single-file` are post readers too:
`kind` stays the host binary (macho/elf/pe), the payload lands in
`deps`/`capabilities` (`.NET single-file` → `dotnet-singlefile`, `klib` stays
`zip`), and `skipped:<reason>` is a refusal, not empty. `mangling` labels a
symbol family only — there is no demangler.

## Failure

- `inspect_go_binary` result is flat: `{kind, toolchain, module, deps, names}`;
  `deps` ≤60 (`lrecon/native.go`), and `names` is only the counter
  `pclntab:<count>` — no symbol list. For names use `--deep` or `list_strings`.
- rc=2, no digest: path wrong or unreadable. `ls -l T`; nothing to parse.
- rc=0, `kind: unknown`: <4 B always; ≥4 B when the first 512 B hold a NUL or
  ≥10% control bytes (`looksText`, `lrecon/detect.go:356`). Printable garbage is
  `kind: text`, so a garbage `text` also means nothing — `xxd -l 64 T`, `file -b T`.
- rc=2, message, partial tree: `--extract-all` writes as it walks. On the same
  248 MB asar: 17 170 files / 189 MB written, then
  `lrecon: --extract-all: "node_modules/resolve/test/pathfilter/deep_ref/main.js" body unreadable`,
  rc=2 (2 GB cap: `lrecon/extract.go:24 var maxExtractAll = 2 << 30`). An
  existing DIR is merged: files already there are overwritten silently, and the
  `extracted N` note counts members written, not new paths — clean the DIR yourself.
- rc=2, message, no tree: `--extract-all` on a non-container —
  `lrecon: --extract-all: real.elf is neither a zip/apk/ipa nor an asar: zip: not a valid zip file`.
  It is not a copy: `--json` or `--deep` that file.
- `--max-bytes N` is a text budget only: `--json`, `--op` and `--ext-op` are never
  cut (one whole JSON on stdout is the contract), and a limit below the answer
  size prints a stderr note instead of truncating it.
- rc=2, message: `--deep` spelling miss or a stripped target. The digest is
  fine; only disassembly failed. A symbol in a non-Go binary is not a failure:
  the line says what symbol it is and which Ghidra op disassembles it.
