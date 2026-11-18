# walsyrea — Linux

`lrecon 1.0.0` (one static binary) + the `walsyrea` agent skill. No
dependencies: CGO disabled, no glibc dependency — the same file runs on Ubuntu,
Debian, Fedora, Arch and Alpine.

```
linux/
  lrecon            x86-64 (server/PC)
  lrecon-arm64      ARM64 (Raspberry Pi 4/5, ARM server, Asahi)
  SHA256SUMS.txt    the files you downloaded must match this
  ../skill/SKILL.md   agent skill (12 reference files in ../skill/references/)
```

## 1. Pick the right binary

```sh
uname -m      # x86_64 → ./lrecon      aarch64 → ./lrecon-arm64
chmod +x lrecon
```

## 2. Prove it runs with 3 commands

```sh
./lrecon --version
# lrecon 1.0.0

./lrecon --list-ops | wc -l
# 235

./lrecon --json /bin/ls
# {"contract_version":1,"kind":"elf","format":"x86-64",...}
```

If all three print what is shown above, the tool fully works on this machine.

## 3. Register the skill with your agent

Copy the `../skill/` folder (at the package root, shared by all three platforms)
into your agent's skill directory, for example Claude Code or OpenCode:

```sh
cp -R ../skill ~/.claude/skills/walsyrea           # Claude Code
cp -R ../skill ~/.config/opencode/skills/walsyrea  # OpenCode
```

Then point the agent at the binary:

```sh
export LRECON="$HOME/tools/walsyrea/lrecon"
```

`SKILL.md` defaults to `./lrecon/lrecon`; `LRECON` overrides that.

## 4. Command equivalents on Linux

127 of the 235 ops are pure Go: they need no external tool. 69 call a tool from
PATH, 39 need Ghidra — a missing tool makes that op refuse openly instead of
guessing.

| skill says (macOS) | Linux equivalent |
|---|---|
| `otool -l T`, `lipo -info T` | `lrecon --op inspect_macho --json T` (on a Mach-O target) |
| `objdump -h/-p T` | same (`binutils`; Debian/Ubuntu: `apt install binutils`) |
| `go version -m T` | same (if Go is installed); otherwise `lrecon --json T` → `toolchain` |
| `strings -a -n 8 T` | same (`binutils`), or `lrecon --op list_strings --json T` |
| `plutil -p T` | `lrecon --op inspect_plist --json T` |
| `unzip -l T` | same, or `lrecon --json T` → `deps` |
| `sqlite3 T .tables` | same (`sqlite3` package) |
| `mktemp -d` | same |
| `codesign -dv T`, `ldid` | none (macOS-signed Mach-O only); `lrecon --op inspect_macho --json T` → `hardening` |
| `jadx`, `apktool`, `adb`, `frida` | same commands, Linux builds exist |

**Ghidra (39 G ops):** `/opt/ghidra*/support/analyzeHeadless`, `GHIDRA_INSTALL_DIR`
and `~/tools/ghidra*` are scanned automatically. If it is not found:

```sh
export LRECON_ANALYZEHEADLESS=/opt/ghidra_12.0/support/analyzeHeadless
```

## 5. Limits

- Both binaries were **executed** on a real Linux kernel (Docker, debian
  `stable-slim`): `lrecon-arm64` natively in an aarch64 container, `lrecon` in
  an x86-64 container under emulation (Docker Desktop Rosetta/qemu). Both gave
  `--version` → `1.0.0`, `--list-ops | wc -l` → `235`, `--json /bin/ls` →
  `kind: elf`. Build: `statically linked`, no dynamic library, NX.
- Live-process and iOS-device ops are out of scope (`../skill/references/limits.md`).
- The tool **never writes** to the target file.

## 6. Integrity

```sh
sha256sum lrecon
```

Compare with `SHA256SUMS.txt`.
