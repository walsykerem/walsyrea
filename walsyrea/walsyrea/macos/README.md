# walsyrea — macOS

`lrecon 1.0.0` (one static binary) + the `walsyrea` agent skill. Nothing to
install, no Homebrew, no DLL/dylib. Copy the binary somewhere and run it.

```
macos/
  lrecon            Apple Silicon (arm64) — M1/M2/M3/M4
  lrecon-amd64      Intel Mac (x86-64)
  SHA256SUMS.txt    the files you downloaded must match this
  ../skill/SKILL.md   agent skill (12 reference files in ../skill/references/)
```

## 1. Pick the right binary

```sh
uname -m      # arm64 → ./lrecon      x86_64 → ./lrecon-amd64
chmod +x lrecon
```

Moving the file through a zip drops the execute bit; `chmod +x` restores it.

## 2. Get past Gatekeeper (unsigned binary)

```sh
xattr -d com.apple.quarantine lrecon
```

Without this macOS says "lrecon cannot be opened because the developer cannot be
verified". Alternative: `System Settings → Privacy & Security → Open Anyway`.
The binary is not notarized; the tool is for analysing your own machine's
software, not for closed-source distribution.

## 3. Prove it runs with 3 commands

```sh
./lrecon --version
# lrecon 1.0.0

./lrecon --list-ops | wc -l
# 235

./lrecon --json /bin/ls
# {"contract_version":1,"kind":"macho","format":"fat","slices":2,...}
```

If all three print what is shown above, the tool fully works on this machine.

## 4. Register the skill with your agent

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

## 5. Everything the skill needs is already on macOS

The tools `SKILL.md` points at ship with macOS: `otool`, `lipo`, `plutil`,
`codesign`, `strings`, `unzip`, `sqlite3`, `mktemp`, `python3`.
Of the 235 ops, 127 are pure Go (need no tool at all), 69 call a tool from PATH
(`objdump`, `unzip`, `jadx`…), 39 need Ghidra. A missing tool makes that op
refuse openly instead of guessing.

**Ghidra (39 G ops):** `/opt/homebrew/Cellar/ghidra/*/libexec/support/analyzeHeadless`,
`/Applications/ghidra_*/support/analyzeHeadless` and `GHIDRA_INSTALL_DIR` are
scanned automatically. If it is not found, pass it explicitly:

```sh
export LRECON_ANALYZEHEADLESS=/Applications/ghidra_12.0/support/analyzeHeadless
```

## 6. Limits

- The binaries in this package were run and verified here: `lrecon` (arm64)
  natively — `1.0.0`, 235 ops, fat Mach-O digest, `list_strings` works — and
  `lrecon-amd64` under Rosetta (`arch -x86_64`), same results. Neither was run
  on a real Intel Mac.
- Live-process and iOS-device ops are out of scope (`../skill/references/limits.md`).
- The tool **never writes** to the target file.

## 7. Integrity

```sh
shasum -a 256 lrecon
```

Compare with `SHA256SUMS.txt`.
