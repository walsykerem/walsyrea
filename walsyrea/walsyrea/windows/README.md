# walsyrea — Windows

`lrecon 1.0.0` (one static exe) + the `walsyrea` agent skill. No installer, no
DLLs, no .NET/Java/Python. Copy the exe somewhere and run it.

```
windows\
  lrecon.exe          Intel/AMD (x86-64)
  lrecon-arm64.exe    ARM (Snapdragon X, Surface ARM)
  SHA256SUMS.txt      the files you downloaded must match this
  ..\skill\SKILL.md  agent skill (12 reference files in ..\skill\references\)
```

## 1. Pick the right exe

Normal PC → `lrecon.exe`. On an ARM PC, rename `lrecon-arm64.exe` to
`lrecon.exe` (the x64 build also runs on ARM, just slower).

On first run Windows says "Windows protected your PC" — the exe is unsigned.
**More info → Run anyway.** Or do it once up front:

```powershell
Unblock-File .\lrecon.exe
```

If antivirus removes it: this is a reverse-engineering tool with no behavioural
signature; add the folder as an exception.

## 2. Prove it runs with 3 commands

```powershell
.\lrecon.exe --version
# lrecon 1.0.0

(.\lrecon.exe --list-ops | Measure-Object -Line).Lines
# 235

.\lrecon.exe --json C:\Windows\System32\notepad.exe
# {"contract_version":1,"kind":"pe","format":"x86-64",...}
```

If all three print what is shown above, the tool fully works on this machine:
the `--json` digest is the core of every op, and the pe/elf/macho parsing is
pure Go.

## 3. Register the skill with your agent

Copy the `..\skill\` folder (at the package root, shared by all three platforms)
into your agent's skill directory, for example Claude Code or OpenCode:

```
%USERPROFILE%\.claude\skills\walsyrea\SKILL.md
%USERPROFILE%\.claude\skills\walsyrea\references\*.md

%USERPROFILE%\.config\opencode\skills\walsyrea\SKILL.md     (OpenCode)
```

Then point the agent at the exe (`set`, not `export`):

```cmd
set LRECON=C:\tools\walsyrea\lrecon.exe
```

`SKILL.md` defaults to `./lrecon/lrecon`; `LRECON` overrides that. To make it
permanent: `setx LRECON "C:\tools\walsyrea\lrecon.exe"`.

## 4. Command equivalents on Windows

`SKILL.md` points at macOS tools. Here is the same job on Windows.
127 of the 235 ops are pure Go: they need no external tool. 69 call a tool from
PATH (`objdump`, `unzip`, `jadx`…), 39 need Ghidra — a missing tool makes that
op refuse openly instead of guessing.

| skill says (macOS) | Windows equivalent |
|---|---|
| `otool -l T`, `lipo -info T` | `lrecon --op inspect_macho --json T`. With LLVM installed: `llvm-objdump --macho --private-headers T` |
| `objdump -h/-p T` | `llvm-objdump -h T` / `-p T` (`winget install LLVM.LLVM`), or `lrecon --json T` |
| `go version -m T` | same (if Go is installed); otherwise `lrecon --json T` → `toolchain` field |
| `strings -a -n 8 T` | `lrecon --op list_strings --json T` |
| `grep -m3 -ao X T` | `lrecon --op search_strings`, or PowerShell `Select-String -Pattern X T` |
| `plutil -p T` | `lrecon --op inspect_plist --json T` |
| `unzip -l T` | `tar -tf T` (ships with Win10+) |
| `sqlite3 T .tables` | `winget install SQLite.SQLite` |
| `mktemp -d` | `$env:TEMP` |
| `codesign -dv T`, `ldid` | none (macOS-signed Mach-O only); `lrecon --op inspect_macho --json T` → `hardening` |
| `jadx`, `apktool`, `adb`, `frida` | same commands, Windows builds exist |

**Ghidra (39 G ops):** Windows Ghidra ships `analyzeHeadless.bat`; it is not
found automatically. Pass it explicitly:

```cmd
set LRECON_ANALYZEHEADLESS=C:\ghidra_12.0\support\analyzeHeadless.bat
```

Without Ghidra these 39 ops refuse openly
(`skipped:ghidra analyzeHeadless not installed`) instead of answering wrongly.

`..\skill\references\ops.md` mentions a helper: `bash ../skill/scripts/find.sh`.
That needs bash + `python3` (Git Bash/WSL). If you do not want that, the
one-line equivalent is:

```powershell
.\lrecon.exe --list-ops | Select-String -Pattern string
```

## 5. Limits

- Windows is **not tested**: the exe was cross-compiled here and verified as PE
  (x86-64/arm64, only `kernel32.dll`, ASLR+NX, static Go), but it was not run on
  a Windows kernel. The 3 commands above are the first real run.
- Live-process and iOS-device ops are out of scope (`..\skill\references\limits.md`).
- The tool **never writes** to the target file; the ops that change a device are
  listed in the README.

## 6. Integrity

```powershell
certutil -hashfile lrecon.exe SHA256
```

Compare with `SHA256SUMS.txt`.
