# lrecon --op — one item per call

`--op NAME [TARGET]` prints `{"op","target","result"}`, exit 0. `--list-ops`
name/line; `--list-ops --json` is an array of `{name,route,needs,desc}`,
235 rows (127 L / 69 Y / 39 G), source of `op-index.md`; `-l` is undefined (rc=2).

Catalog helper: `LRECON=/abs/path/lrecon bash scripts/find.sh --list-ops --compact`.
It needs `LRECON` or `lrecon` on `PATH`; with neither it dies `lrecon: command not found`.

## Modes

Shapers, `--op` or `--list-ops`: `--compact` changes nothing a caller can measure —
byte-identical on `--op`, `--list-ops`, `--list-ops --json` and `--list-ext-ops`
(44342 ↔ 44342); the default output is already that compact form, `--compact=false`
is the long one (`--list-ops --json --compact=false` 54105 B vs 44342 default). Pass the flag only if some other tool wants one line. `--count-only` on
`--list-ops` answers the catalog question (5 tok). All four shapers
(`--count-only`, `--fields a,b`, `--top N`, `--by field`) need a record array
(`rows`): `list_procedures`/`list_segments`/`list_strings` refuse them with `usage`
(`this answer carries no record array`) — use `--coverage` there. Wrong or missing
argument: the op prints its own usage line.

## Failure

`needs` names what the call wants besides the target. `procedure_*` and most G ops
want `LRECON_AT` or `LRECON_NAME`; a bare `--op procedure_pseudo_code T` therefore
answers `needs_arg` with that hint — the target is fine, the address is missing.
`inspect_native_exports` is PE/ELF only; on a Mach-O it answers `bad_arg` — Mach-O
exports are in `inspect_macho`.

Every failure exits 2, writes `lrecon: …` to stderr and, with `--json` or a named
`--op`, one envelope on **stdout**: `{"contract_version":1,"op","error","kind"}`. `kind` is closed:
`needs_arg`, `bad_arg`, `out_of_scope`, `usage`, `no_engine`, `timeout`, `io`,
`state_conflict`, `internal`. Filesystem failures are `io`; a value wrong for this call is `bad_arg`. `no_engine` = no engine for this kind (wrong-kind file, or the tool is missing);
`out_of_scope` = never done by contract (live process/URL, frida, rizin).
Table-only classes map to live kinds (`TestOpClassLiveKindMap`).

## Routes

L = tables the binary carries (stripped → G); G = `analyzeHeadless` +
`NativeOp.java`, 50 s; Y = adb/lldb, 60 s.

## Env

`LRECON_AT` wins over `LRECON_NAME`; `LRECON_FROM`+`LRECON_TO` both required.
Read order, defaults, full list: `contract.md#env`.

## Byte symptoms

Read the bytes before the op: one constant names the primitive, so you never reverse the wrong model. A symptom that fits none of these is a hint to keep reading, not to guess.

| tell | primitive | first move |
|---|---|---|
| 256 dwords, `table[0]==0`, `table[1]` = `0x77073096` | crc-32 | poly = `table[1]`; don't diff outputs |
| xor then multiply by `0x01000193` / `0x100000001b3` | fnv-1a | fnv-1 multiplies first — order is the whole difference |
| shift right 1 plus conditional xor of `0xb400` / `0x80200003` | lfsr | the xor constant is the tap set |
| `x ^= x<<13; x ^= x>>17; x ^= x<<5` | xorshift | shifts come from the code, output inverts |
| `63 7c 77 7b f2 6b 6f c5`, or `0x1b` in a column mix | aes | key schedule sits in the same frame |
| `00 01 02 … ff` array with a 256-step swap loop | rc4 | KSA loop; the key is its input |
| `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdef…+/=` | base64 | `search_strings`/`list_strings` first, decode before disassembly |
| `78 9c`/`78 01`/`78 da`, or `1f 8b` | zlib / gzip | inflate the segment; the payload is not code |
| `adrp` + `add`/`ldr` with a 12-bit immediate | PIC address | `adrp` page plus that immediate; not a data pointer |
| small `__DATA` value like `0x100000b5d` | chained fixup | `router.md#limits` — `dyld` chain, not a pointer |
| `ld1`/`st1`, `q` regs, `tbl`, `cmeq` on 16 lanes | neon/simd | compare lanes, not bytes |
| `LC_DYLD_CHAINED_FIXUPS` in `otool -l` | dyld rebase/bind | `dyld_info -fixups T` |

## Names

Every op name lives in `op-index.md` (L 127 / G 39 / Y 69, current with 1.0.0);
this file carries the modes, the failure shape and the ext-ops list.

## External-tool items (`--ext-op`)

`lrecon --ext-op NAME [ARG...]`, `--list-ext-ops`; contract and machine answer in
`lrecon/testdata/ext_ops.json`, regenerated with
`go test -run ExtTableContract -update-ext`. `not_applicable` = no target pass.
`jadx_plugins` (507), `apktool_clean_frameworks` (510) take no operand: a target
is `bad_arg`. All 80 names `--list-ext-ops` prints:

adb_attach, adb_connect, adb_detach, adb_disable_verity, adb_disconnect, adb_emu,
adb_enable_verity, adb_forward_add, adb_forward_list, adb_forward_remove, adb_get_serialno,
adb_get_state, adb_jdwp, adb_keygen, adb_reboot, adb_remount, adb_reverse_add, adb_reverse_list,
adb_reverse_remove, adb_root, adb_server, adb_tcpip, adb_unroot, adb_usb, adb_wait_for_device,
apktool_build, apktool_clean_frameworks, apktool_frame_path, apktool_install_framework,
apktool_lib, apktool_list_frameworks, apktool_no_res, apktool_no_src, apktool_only_manifest,
apktool_publicize_resources, ghidra_analysis_timeout, ghidra_commit, ghidra_connect, ghidra_cspec,
ghidra_keystore, ghidra_library_paths, ghidra_loader, ghidra_mirror, ghidra_noanalysis,
ghidra_overwrite, ghidra_prescript, ghidra_processor, ghidra_readonly, jadx_call_graph, jadx_cfg,
jadx_deobf, jadx_export_gradle, jadx_fallback, jadx_mappings, jadx_no_res, jadx_no_src,
jadx_output_json, jadx_plugins, jadx_raw_cfg, jadx_single_class_full, lldb_breakpoint_delete,
lldb_breakpoint_list, lldb_breakpoint_set, lldb_diagnostics, lldb_disassemble, lldb_expression,
lldb_frame_variable, lldb_image_list_full, lldb_image_lookup, lldb_memory_read, lldb_memory_write,
lldb_process_attach, lldb_process_launch, lldb_register_read, lldb_register_write,
lldb_source_list, lldb_target_list_full, lldb_thread_backtrace, lldb_thread_list_full,
lldb_watchpoint

## Op invocation

All 235 ops share one call form — no op has a private CLI:

```
[VAR=… …] $LRECON --op NAME [TARGET] --json
```

- `NAME` = any name from `--list-ops`; route (L/Y/G) and family never change the form. `--json` → envelope `{"op","target","result"}`.
- `TARGET` = (a) a bare artifact path, or (b) a **job-JSON file path**: a JSON object with keys `file`, `state`, `old`, `new`, `view`, `address`, `text`, `note`, `name`, `count`, `labels`, `claims`, `string`, `bytes`, `hypothesis`, `resolution`, `resolved_as`, `apply`, `reason`, `symbols`, `unresolved`, `crash`, `plist`, `include`, `resolved`, `rows`, `expect`, `deep`, `maxBytes`, `minString`; `file` names the artifact, the rest are that op's extra inputs. A plain executable, a saved digest or a directory is not a job — it falls back to the path itself. Multi-input families use (b): `verify_*`, `update_*`, `set_address_name`, `set_addresses_names`, `set_comment`, `set_inline_comment`, `trace_application_feature`.
- Extra args ride env slots, never positional: `LRECON_AT` (address; beats `LRECON_NAME`), `LRECON_NAME` (symbol, only when AT empty), `LRECON_FROM`+`LRECON_TO` (both required, pair ops), `LRECON_ARGS` (comma-separated names/addresses), `LRECON_QUERY`, `LRECON_PATTERN`, `LRECON_STATE`/`LRECON_SESSION` (stateful ops: `open_binary`, bookmarks, `record_unknown`, obligation ledger, every debug/session op — adb, frida, gdb, rizin, electron/web). `LRECON_JOB`/`LRECON_FUZZ_OP` are harness knobs: leave unset.
- Per-op slot: `needs` in `lrecon --list-ops --json`. Exit 2 has no `needs` key — read `error`: `needs_arg` = set the named slot, then one retry; `bad_arg` = bad flag/value or wrong kind (`inspect_native_exports`/`verify_managed_native_boundaries` need PE/ELF; `inspect_native_dispatch_metadata` needs Obj-C/libdispatch), `out_of_scope` = never (live process/URL), `no_engine` = device/tool absent. Host probes in Route (`otool`, `objdump`, …) are cheap previews, never a substitute for a `--op` call. `capture_*`, `inspect_elf_notes`, `inspect_pe_*`, `inspect_asset_catalog`, `inspect_dart_aot` are ordinary `--op` calls under this same form.
