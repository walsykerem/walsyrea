# walsyrea

A self-contained reverse-engineering toolkit: **`lrecon 1.0.0`**, one static
binary per platform with **235 ops** (Mach-O / ELF / PE, Go, Rust, .NET, DEX,
APK, IPA, pyc, plist, SQLite, firmware, …), plus the **`walsyrea` agent
skill** that teaches a coding agent when and how to call it.

No installer, no runtime, no DLLs, no Python/Java/.NET. Copy the binary and run.

## What is in this archive

```
README.md            this file
LICENSE
walsyrea/
  README.md          full package notes (install, per-OS details, hashes)
  LICENSE
  windows/           lrecon.exe (x86-64), lrecon-arm64.exe, README.md, SHA256SUMS.txt
  macos/             lrecon (Apple silicon), lrecon-amd64 (Intel), README.md, SHA256SUMS.txt
  linux/             lrecon (x86-64), lrecon-arm64, README.md, SHA256SUMS.txt
  skill/             SKILL.md + 12 references + scripts/find.sh
```

Pick one platform folder. The `skill/` folder is shared by all three.

## Quick start

Windows (Intel/AMD) — PowerShell:

```powershell
.\walsyrea\windows\lrecon.exe --version                       # lrecon 1.0.0
(.\walsyrea\windows\lrecon.exe --list-ops | Measure-Object -Line).Lines   # 235
.\walsyrea\windows\lrecon.exe --json C:\Windows\System32\notepad.exe      # pe digest
```

On an ARM PC use `lrecon-arm64.exe` (rename it to `lrecon.exe`).
First run trips SmartScreen because the exe is unsigned: **More info → Run
anyway**, or `Unblock-File .\lrecon.exe` once.

macOS:

```sh
./walsyrea/macos/lrecon --version                 # lrecon 1.0.0
./walsyrea/macos/lrecon --list-ops | wc -l        # 235
./walsyrea/macos/lrecon --json /bin/ls            # macho digest
```

Linux:

```sh
./walsyrea/linux/lrecon --version                 # lrecon 1.0.0
./walsyrea/linux/lrecon --list-ops | wc -l        # 235
./walsyrea/linux/lrecon --json /bin/ls            # elf digest
```

If those three print what is shown, the tool works: the `--json` digest is the
core of every op and the Mach-O/ELF/PE parsing is pure Go.

## Install the agent skill

Copy `walsyrea/skill/` into your agent's skill directory:

```
%USERPROFILE%\.config\opencode\skills\walsyrea\      (OpenCode, Windows)
%USERPROFILE%\.claude\skills\walsyrea\               (Claude Code, Windows)
~/.config/opencode/skills/walsyrea/                  (OpenCode, macOS/Linux)
```

Then point the skill at the binary (defaults to `./lrecon/lrecon`, `LRECON`
overrides it):

```
set LRECON=C:\tools\walsyrea\lrecon.exe
```

The skill is what makes an agent use `lrecon` instead of guessing: route table,
op index, contract schema, evidence rules, and hard limits.

## Ops at a glance

| Route | Count | Needs |
|---|---|---|
| `G` | 39 | Ghidra `analyzeHeadless` |
| `L` | 127 | nothing (pure Go) |
| `Y` | 69 | an external tool from `PATH` (`objdump`, `jadx`, `frida`, …) |
| ext | 80 | — (`--list-ext-ops`) |

A missing external tool makes that op refuse openly
(`skipped:<reason>`), never guess. `lrecon` never writes to the analysed file.

## Integrity

```powershell
certutil -hashfile lrecon.exe SHA256      # compare with windows\SHA256SUMS.txt
```

```sh
cd walsyrea/macos && shasum -a 256 -c SHA256SUMS.txt
```

## Limits

- Windows binaries are cross-compiled and verified as PE (x86-64/arm64, only
  `kernel32.dll`, ASLR+NX, static Go), but were **not run on a Windows kernel**.
  The three quick-start commands are the first real run.
- Live-process instrumentation, iOS devices, and GUI-only workflows are out of
  scope; see `walsyrea/skill/references/limits.md`.
- Ghidra-backed ops need `LRECON_ANALYZEHEADLESS` set to `analyzeHeadless.bat`
  (Windows) or `analyzeHeadless` (macOS/Linux).

## License

See `LICENSE`.
