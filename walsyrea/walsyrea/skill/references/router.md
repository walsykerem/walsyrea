# Router — what the digest says, where it leads

## read-digest

`lrecon --json T` sets the route; `SKILL.md` holds the rows.

| key | values | next action |
|---|---|---|
| `kind` | `#kinds` | route row |
| `format` | thin, fat (macho), the arch for elf/pe, asar, electron, pcap(-be,-ns,-ns-be), pcapng, `v<N>` (wasm), directory, or firmware: squashfs, squashfs-be, ubi, cramfs, cramfs-be, trx, uimage, dtb, jffs2, gzip, xz | fat: `lipo -info`; thin one slice |
| `slices`, `slice_info` | fat macho only | >1 arch: patch one, offsets differ |
| `dylibs` | macho, elf basenames, pe; capped at 40 | elf: `soname`; both: `imports` |
| `imports` | native, capped at 60; macho raw `_ptrace`, elf/pe bare `ptrace` | `_ptrace`/`_dlsym` → `bypass.md#layers` |
| `hardening` | pie, nx, codesign, cfg, canary, debug-info:* | codesign → re-sign after a patch |
| `capabilities` | il2cpp, managed, pinvoke, android, ios, signed-provision, flutter, react-native, electron, native-modules, wasm, sqlite, fuses, asar-unpacked, asar-unpacked-native, preload-bridge:<name>, updater, updater:Squirrel/Sparkle/electron-updater/autoUpdater/tuf | `wasm` → rowsuz-kinds; else its own row |
| `toolchain`, `mod` | `go1.26.5 darwin/arm64`, module path | Go: `go version -m` |
| `symbols` | present, stripped | decided from the in-process symbol table (`debug/macho\|elf\|pe`), no host tool; `present` means a table exists, **not** that local symbols survived — a fully stripped Mach-O still reports `present`; check `nm T \| grep -c ' [tT] '` before trusting any symbol name; fat: empty, per-arch in `slice_info` |
| `entrypoint` | `e_entry` (elf), LC_MAIN (macho), ImageBase+AddressOfEntryPoint (pe); `null` when the target declares none; the digest prints it **decimal**, the ops print hex — same value | the loader starts here: a patch at this offset runs first |
| `note:` lines (text digest only) | `deep: …`, `truncated`, `no known magic`, one `note:` each; `--json` has no `notes` key — `jq .notes` is null | act on that line |


Other 51 keys, no own action: `deps` (count),
`exports`, `minos`, `uuid`, `soname`, `build_id`, `subsystem`, `pdb`, `rpaths`,
`target`, `file_size`, `bytes_scanned`, `truncated`,
`extracted`, `sqlite_page_size`, `sqlite_version`, `sqlite_wal`, `asar_files`,
`asar_main`, `asar_product`, `asar_version`, `asar_unpacked_native`, `electron`,
`flutter_hash`, `electron_ipc`, `source_map`, `rust_crate`, `plist_key`,
`mobile_package`, `mobile_sdk`, `mobile_permissions` (text `mobile_permission`),
`firmware_parts` (count), `firmware_detail` (indented), `firmware_kind`,
`pcap_packets`, `pcap_link_type`, `pcap_span`, `pcap_protocols` (text `pcap_proto`),
`pcap_dns`, `pcap_sni`, `pcap_uris` (text `pcap_uri`),
`contract_version`, `skeleton`, `coverage`, `claims`, `conflicts`, `constants`,
`jit`, `names`, `unknown`, `skipped`.

Route split (`--list-ops --json`, 235 ops): 127 L / 69 Y / 39 G (order L, Y, G).

These 64 top-level keys are the whole digest of the pinned `lrecon 1.0.0`; a
newer build may add more. Every key is always emitted with its zero value (0,
`[]`, `false`, `""`) when nothing was found — an absent key is not "unknown", and
a zero/empty value is not "no packer/signature/stripped": the first command of the
matching `bypass.md#layers` row detects those. Needing the code →
`ghidra.md#cost-ladder`.

## kinds

detect.go names `unknown`, `macho`, `elf`, `pe`, `asar`, `dex`, `android-manifest`,
`hermes`, `flutter`, `wasm`, `sqlite`, `plist`, `nib`, `signature`, `car`, `zip`,
`pcap`, `firmware`, `text`; bundle.go adds `app` and `directory`, mobile.go
`apk`/`aab`/`ipa`, managed.go `il2cpp-metadata` and `pe-dotnet`.

Registry readers claim a kind too: magic readers (`post` false) `pyc`,
`php-opcache`, `java`, `lua`, `move`, `beam`, `v8-cache` (`v8cache.go:47`) take
`kind` = their format; `compress` (`compress.go:115`), `ocaml` (`ocaml.go:90`)
and `rlib` (`rlib.go:77`) set it from `inspect`; `phar` claims it only when
nothing else did (`claimsKind`, `php.go:46`). `firmware-ext` never lands as a
kind: a match already names the part, so `inspect` turns `kind` into `firmware`
and puts the part in `format`/`firmware_kind` (`firmware_ext.go:111`). Every
other registry reader is a post reader: the host `kind` stays and the format
lands in `format`/`capabilities`.

## match-rules

| route label | signal | rule |
|---|---|---|
| `go (toolchain)` / `rust` | `toolchain` starts `go1.` / `rustc` | prefixes, never a `kind`; crates in `rust_crate` |
| `.NET / IL2CPP` | `kind=="pe-dotnet"` (managed.go:77) / `il2cpp` (managed.go:630) / `kind=="il2cpp-metadata"` (managed.go:38) | three signals, one row |
| `electron / asar` | `format=="electron"` (bundle.go:24) / capability `electron` (asar.go:207: any path segment `/electron/`, scoped pkgs too) | format = .app owning an asar; capability = bundled electron; bare `kind=="asar"` is neither |

## rowsuz-kinds

Kinds with no route:

| kind | fallback |
|---|---|
| dex | `baksmali d T -o out` |
| android-manifest | `aapt2 dump xmltree --file AndroidManifest.xml APK`; alone `strings -a T` |
| wasm | `wasm-objdump -x T`; absent → `xxd -l 96 T` (section ids) + `strings -a T` |
| nib | `ibtool --objects T`, else `plutil -p T` |
| signature | `codesign -dvvv T`; packaged ipa/app: `lrecon --extract-all DIR T` first, then codesign the inner `.app` |
| car | `assetutil --info T` |
| v8-cache | not recoverable: continue from the source script |
| firmware | `dd` at a `firmware_detail` offset, `kind` the part |
| app | `ls T/Contents/{Resources,MacOS}`, `kind` that file |
| directory | `ls -la T`, `kind` one artifact inside |
| java | `javap -p -c -classpath . Class`; deps `jdeps`; no `javap` → metadata only |
| lua | `luac -l -p T` when the version matches; else bytecode is data, `strings -a T` |
| move | not native code: `xxd -l 128 T` + `strings -a T`; upstream `move-disassembler` if present |
| beam | `erl -noshell -eval 'beam_lib:chunks("T",[abstract_code])'`; else metadata only |
| ocaml | `ocamlobjinfo T` with a matching compiler; else metadata only |
| compress | gzip/bzip2: `gunzip -c`/`bunzip2 -c`; xz/lz4/zstd need the named tool |
| rlib | `ar t T`, `nm -g T`; the rustc metadata blob is version-pinned |

Catch-all: any kind not in this table and not in the Route table still ends at
`VERDICT: unknown` with the bytes read — never invent a route.

Source files: `.go` (and other source extensions) carry no binary magic, so
`lrecon --json T` reports `kind:"text"`, `slices:0`, `bytes_scanned:0`, plus a
whole-file `file_sha256` and `file_size` — that digest is the whole verdict. The
string/byte ops refuse text targets (`T is not ELF, Mach-O or PE`), so read the
source with your own tools. This is not source-repo analysis (see `limits.md`).

`wasm-objdump`, `aapt2` and `baksmali` are not in a stock macOS toolchain (`ibtool`,
`assetutil` are): `command -v` first, absent → `partial`, read the bytes with `xxd`/`strings`.

## Limits

- `rustDigest` `case "pe"` is dead: rust runs only for elf/macho, so a Rust PE has no `toolchain`.
- A `__DATA` word such as `0x100000b5d` is a **chained fixup** (`LC_DYLD_CHAINED_FIXUPS`), not a raw pointer: `dyld_chained_ptr_64_rebase` packs `target` (low 36 bits = unslid vmaddr), `next` (`w >> 51`) and a bind flag in the top bit. File offset = `target - __TEXT vmaddr`. `otool -l` naming that load command is the tell; read the chain with `dyld_info -fixups T`, LIEF, go-macho `pkg/fixupchains` or r2 `mach0.c` — macholib and kaitai do not, and no `adrp` pair computes it.

## Digest plan

Digest `plan` is one of these five ids; run its probes in order, stop at the first that names the layer. Host probes (`otool`, `objdump`, `unzip`) come first: a G op pays a cold Ghidra import (16 s+ on these fixtures) and is only worth it once the host probe names no layer.

| plan | probes in order |
|---|---|
| `macho` | `otool -l T` → `--op list_strings T` → `--op list_segments T` → `--op list_procedures T` |
| `elf` | `objdump -h T` → `--op inspect_elf_notes T` → `--op list_segments T` |
| `pe` | `objdump -p T` → `--op inspect_pe_imports T` → `--op inspect_pe_resources T` |
| `mobile` | `unzip -l T` (apk/ipa; dex and android-manifest: `#rowsuz-kinds`) → `--op inspect_artifact T` → `--op inspect_apk_signature T` (apk) or `--op list_jeb_units T` (only with `classes*.dex`) |
| `generic` | `--op binary_overview T` → `--op inspect_artifact T` → `references/router.md#read-digest` |

