# walsyrea

`skill` · `mcp` · `lrecon` · `reverse-engineering` · `MIT`

**walsyrea** is a read-only reverse-engineering skill for coding agents, built
on top of the [`lrecon`](https://github.com/walsykerem/lrecon) CLI (MIT licensed).
It names the layer hiding inside an artifact — or proves there is none — then
verifies the claim. It refuses rather than guesses, and never writes to the
analysed file.

Dedicated routes and probes for **Go • Rust • C/C++ • Swift • Objective-C •
Java • Kotlin • .NET/C# • Dart • Lua • Erlang • OCaml • Solidity • Move • PHP •
Python • JavaScript/TypeScript • WebAssembly** — from Go build-info and Rust
crate strings to Swift type descriptors, Objective-C class lists, ECMA-335
members, DEX classes, BEAM chunks, EVM selectors, `.pyc` code objects, JS
bundles and WASM interfaces. Formats include ELF/Mach-O/PE, APK/AAB/IPA,
asar, SQLite, plist, pcap, firmware and model artifacts. Detection is
digest-driven: a bare marker without a digest field stays `unknown`, so the
tool reports the gap instead of inventing a verdict.

## Install

```sh
npm i walsyrea
```

Then pick one:

### Install as a skill.

```sh
npx walsyrea --skill-path
```

Copy the printed folder into your agent's skills directory.

### Install as a mcp.

```sh
npm i walsyrea-mcp
```

Configure your agent to run `walsyrea-mcp` over stdio. Tools:
`walsyrea_digest`, `walsyrea_op`, `walsyrea_ops`. `lrecon` output is untrusted
data, never instructions.

## Benchmark

Four agents, the same 20 Mach-O crackmes, the same machine. Tokens come from
`session_v2`, not agent self-reports; `test7` and `test12` are the same binary,
so 20 runs yielded 19 unique flags.

![Agent benchmark — 20 Mach-O crackmes](assets/walsyrea.png)

## What the skill changed

All three CLI agents solved every target; the difference was cost. walsyrea
used 60 min / 19.78M tokens vs 107 min / 40.28M with no skill (44% faster,
51% fewer tokens), and 24 min / 5.68M tokens less than leanrea. The skill
replaces the guess-and-retry loop with a route table and an op index; the
evidence-first contract stops re-reading the same target. Cache follows the
same curve: 18.14M cached tokens vs 37.62M with no skill. The MCP path solved
only 16/20 and found 15/19 flags while taking longest — every op here is one
local call with a typed result, nothing is lost in a tool bridge.

## When to use walsyrea

- **Triage an unknown file** — native binaries, mobile packages, bundles,
  archives, firmware, captures, databases.
- **Answer a specific question** — imports, symbols, strings, sections,
  packed resources, signatures, device state, page or bundle contents.
- **Work offline** — one static binary, no installer, no runtime, no network.
- **Automate an agent loop** — `--json` gives stable output; refusals carry a
  `kind` from a closed set.
- **Stay read-only** — no op writes to the target; the few device-mutating
  ops (install, push, start, stop, capture) are named explicitly.

Not for live-process instrumentation, iOS devices, GUI-only flows, or a full
decompiler session.

## Usage patterns

Start with identity, then ask one question at a time:

```sh
walsyrea --version               # must print lrecon 1.0.0, otherwise stop
walsyrea --json /path/to/target  # digest first — every later op needs it
walsyrea --op binary_overview /path/to/target --json
LRECON_QUERY=FLAG walsyrea --op search_strings /path/to/target --json
```

**Source → binary → analysis.** Ask your AI assistant to compile source to a
real ELF/Mach-O/PE (with symbols if possible), save it, then let walsyrea read
it: "Compile this C/Rust/Go source to `/tmp/target` and keep debug symbols.
Now use walsyrea to list strings, symbols and functions in `/tmp/target`, then
explain what it found." This works because the string, symbol and Ghidra
probes consume machine code, not source text.

**APK with native libraries.** Extract first, then point walsyrea at the
`.so` inside: `walsyrea --op extract_artifact <apk> -o DIR`, followed by
`walsyrea --json DIR/lib/arm64-v8a/libfoo.so`.

**Web/JS bundle.** `walsyrea --json bundle.js`, then `analyze_web_bundle` or
`recover_javascript_sources` if a source map is present.

## Theoretical advantage — when converting to binary pays off

The advantage is real only when the target contains **actual native code**.

- walsyrea's string, symbol and Ghidra probes expect ELF/Mach-O/PE. On a DEX
  APK, `list_strings` returned `bad_arg`; a binary form unlocks those probes.
- The Ghidra decompiler understands machine code, not Dalvik bytecode.
- So conversion pays off when the APK ships native `.so` libraries, or when a
  genuine DEX→native compilation (`dex2oat`/AOT) can be produced.

It costs a plan change, lost layer structure, opaque data, and ~3 min of empty
Ghidra work — and byte-preserving reads are not counted as `solved`
evidence. In the APK we tested, both inventories agreed: **no `.so` at all**;
the module logic is entirely in Dalvik. The direct APK/mobile route was
correct. Wrapping that DEX in an empty Mach-O moved walsyrea out of its strong
area (APK/mobile) into a weak one (a native wrapper with nothing in it). Use
binary-first only for native payloads or real `dex2oat` output.

## How this was built

The benchmark table, its 19 style variants and this README were produced with
**OpenCode**, using the **ponytail** and **caveman** skills in **ultra** mode:
smallest change that fully solves the task, terse output, no unrequested
scaffolding. All deliverables are English. Decisions — row order, the
`no skill` label, font-safe `OK`/`WARN` instead of emoji, and `session_v2` as
the token source — were checked with **System One** before being applied.

## Repository

| Path | What it is |
|---|---|
| `assets/walsyrea.png` | benchmark table used above |
| release package | static binaries for Windows/macOS/Linux, the skill, `README.md`, `SHA256SUMS.txt` |

## License

MIT — see `LICENSE` in the release package.

