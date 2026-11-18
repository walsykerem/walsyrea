# walsyrea

`lrecon 1.0.0` — one static binary that answers questions about a packaged
binary/app through 235 `--op` + 80 `--ext-op`, and openly refuses what it
cannot answer. It **never writes** to the target file. The `walsyrea` agent
skill ships alongside it.

```
walsyrea/
  skill/            agent skill — shared by all three platforms (SKILL.md + 12 references + scripts/find.sh)
  windows/          lrecon.exe (x86-64), lrecon-arm64.exe, README.md, SHA256SUMS.txt
  macos/            lrecon (arm64), lrecon-amd64 (Intel), README.md, SHA256SUMS.txt
  linux/            lrecon (x86-64), lrecon-arm64, README.md, SHA256SUMS.txt
  LICENSE           MIT
```

## Which folder?

| Machine | Folder | Binary |
|---|---|---|
| Windows Intel/AMD | `windows/` | `lrecon.exe` |
| Windows ARM (Snapdragon X, Surface ARM) | `windows/` | `lrecon-arm64.exe` (rename it to `lrecon.exe`) |
| Apple Silicon (M1–M4) | `macos/` | `lrecon` |
| Intel Mac | `macos/` | `lrecon-amd64` |
| Linux x86-64 | `linux/` | `lrecon` |
| Linux ARM64 (Pi 4/5, ARM server) | `linux/` | `lrecon-arm64` |

Enter your folder; its `README.md` covers the platform-specific setup
(Windows SmartScreen, macOS Gatekeeper, Linux chmod). These 3 commands are the
same everywhere:

```sh
./lrecon --version            # lrecon 1.0.0
./lrecon --list-ops | wc -l   # 235
./lrecon --json <any-binary>
```

On Windows use `.\lrecon.exe` instead of `./lrecon`, and
`(.\lrecon.exe --list-ops | Measure-Object -Line).Lines` instead of `wc -l`.

## Skill

```sh
cp -R ../skill ~/.claude/skills/walsyrea             # Claude Code
cp -R ../skill ~/.config/opencode/skills/walsyrea    # OpenCode
export LRECON="$PWD/lrecon"
```

Windows: `xcopy /E /I ..\skill %USERPROFILE%\.claude\skills\walsyrea`,
`set LRECON=C:\path\lrecon.exe`. The OpenCode skill directory on Windows is
`%USERPROFILE%\.config\opencode\skills\walsyrea`.

Inside `SKILL.md`, `lrecon` means `$LRECON`. The skill ships with
`disable-model-invocation: false` — the agent loads it on its own. Set it to
`true` if you want to load it yourself.

## What it does, what it does not

- **127 ops are pure Go**: no external tool needed (Mach-O/ELF/PE/Go/Rust/.NET/
  DEX/pyc/asar/apk/ipa/SQLite/plist/pcap parsing, strings, symbols, hardening).
- **69 ops call a tool from PATH** (`objdump`, `unzip`, `jadx`, `frida`…); if
  the tool is missing the op refuses instead of guessing.
- **39 ops need Ghidra** (`analyzeHeadless`); without it they return
  `skipped:`.
- No disassembler/decompiler — that is Ghidra's job.

## Integrity

```sh
cd macos && shasum -a 256 -c SHA256SUMS.txt     # linux: sha256sum -c
cd windows && shasum -a 256 -c SHA256SUMS.txt   # powershell: certutil -hashfile lrecon.exe SHA256
```

## Verification status

- `macos/lrecon` (arm64) was **executed** here: `1.0.0`, 235 ops, `/bin/ls` →
  `macho fat 2 slice`, `list_strings` works. `macos/lrecon-amd64` returned the
  same results under Rosetta (`arch -x86_64`).
- `linux/lrecon-arm64` and `linux/lrecon` were **executed** on a real Linux
  kernel (Docker, debian `stable-slim`; arm64 native, amd64 emulated): `1.0.0`,
  235 ops, `/bin/ls` → `kind: elf`. Build: static, no dynamic library, NX.
- `windows/*.exe`: cross-compiled from the current source and verified as PE
  (x86-64/arm64, only `kernel32.dll`, ASLR+NX). Not run on a Windows kernel —
  the first real run is yours.
