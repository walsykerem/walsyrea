# Bypass — unpack, patch, prove

One row per layer: detect once, defeat once, prove with bytes. A layer you cannot
name is not a layer you can defeat — say `partial` and quote what is unknown.

## Layers

| layer | detect | defeat |
|---|---|---|
| UPX | `strings -a -n 4 T \| grep -m1 'UPX!'` | `upx -d T -o T.unpacked` |
| custom packer (mmap + XOR) | `nm -u T \| grep -E '_(mmap\|mprotect)$'` | find the loop in `objdump -d T`, dump decrypted pages: `lldb -b -o 'b *0xOFF' -o run -o 'memory read -f x -c 128 $x0' T` |
| KDF | `strings -a -n 8 T \| grep -iE 'pbkdf2\|scrypt\|argon\|rounds'` | replay the rounds in one `python3` script; read the round constant from disassembly, never guess it |
| CFF / state machine | `objdump -d T \| grep -c 'cmp.*#0x'`; on Mach-O use `otool -tvV T` (x86_64 prints `$0x`, so `#0x` silently misses, and a fat file resolves to a different slice per tool); arm64 often shows `cbz`/`cbnz` + a jump table instead, so a low `cmp` count proves nothing; one state slot reused per block | dump the state→block table once (`xxd -s TBL -l 256 T`), write the real order; do not single-step every state |
| anti-debug ptrace | `nm -u T \| grep -E '_ptrace$'`; `otool -tV T \| grep -n -B2 -A2 ptrace` | nop each call site (`00 01 3F D6` → `1F 20 03 D5` on arm64), then re-sign; every site must be nop'd, not just the first found |
| ObjC anti-debug | `strings -a T \| grep -m2 -E 'NSProcessInfo\|PT_DENY_ATTACH'` | patch the selector call, or `frida-trace -f BUNDLE -m '-[NSProcessInfo …]'` |
| obfuscated constant/string table | `list_strings` returns only short junk while the binary prints readable text at runtime (`scanned_whole:true` does not mean "all readable") | replay the loop `procedure_pseudo_code` shows (key, stride, XOR/add) in one `python3` over `read_bytes` at the table address; check the result against the known output |
| encryption / integrity | `otool -l T \| grep -A4 LC_ENCRYPTION_INFO` (cryptid 1 = the layer); a signature blob (`LC_CODE_SIGNATURE`) alone is **not** a layer — it only forces a re-sign after a patch; an ipa/apk/zip must be extracted first — `otool -l T` prints `is not an object file` on **stdout** with exit 0, so the grep goes quietly empty | decrypt `__TEXT` first, then `codesign --remove-signature T; codesign -s - T`; iOS: `ldid -S T` |
| VM / bytecode | `otool -tvV T \| grep -cE 'cmp.*#0x'` (arm64: `cbz`/`cbnz` + jump table) — a handler **loop** is the tell: one block comparing opcode bytes, dispatching back to itself. `strings` finds nothing (tokens are `vm_`/`opcode`-free, and a decoy `FLAG{…}` is the usual false positive) | read the loop (`procedure_pseudo_code`) to get the table address, stride and key, then decode the table with `read_bytes` (packed 2-byte opcodes / XOR'd or rotated entries are decoded before use — `xxd` on it prints junk); the table sits in a data segment, never in the string table. A handler loop **with no literal opcode table** is still a VM: read each compare constant as an opcode. Only no loop **and** no table → `partial` |
| .NET managed | capability `managed`, or `strings -a T \| grep -m3 mscorlib` | `ilspycmd T -o out`, read the CIL; absent → `partial` |
| no layer | all rows above probed empty — for the obfuscation row that means no constant/string table you can name by address (runtime text you cannot locate is still a layer) | `LAYERS: none`, skip this page: read code, go to Recover; prove behaviour (next section) |

A runtime decoy is not a `check_deception_signals` class: that op reads three
static classes (`canary-credential`, `honey-sourcemap`, `shadow-state-machine`) and
answers `deception_risk:"none"` on a target that only decoys at runtime. The exit
code is the tell: a wrong or missing argv prints the same `FLAG{…}` and exits
non-zero (`rc=1` bad input, `rc=2` no argv), and that decoy string sits plaintext in
the binary — `strings` finds it, then ignore it; only `rc=0` is a solve. Run under
`lldb`/`frida` only for the ptrace/anti-debug rows: a debugger changes what the
target does, so the run you prove with is a plain `./T <args>`.

`upx`, `frida-trace`, `ldid`, `apktool` and `ilspycmd` are not in a stock macOS toolchain
(`lldb` is: `/usr/bin/lldb`). Install before the layer, or record it as
`partial` — never as defeated.

## Patch ritual

In `$TMPDIR/patchlab` with `rechal` (hypothetical example: `rechal` is not in
this repo, the addresses and `status = 45` are representative — substitute your
own Mach-O target):

```sh
set -eu
cp -p rechal rechal.orig
shasum -a 256 rechal rechal.orig
T=rechal T0=rechal.orig OFF=$((0x100000970 - 0x100000000 + 0))   # 0x970
```

1. Locate: file offset = address − **segment** vmaddr + **segment** fileoff.
   Here: `__TEXT` vmaddr `0x100000000` fileoff `0`; `__text` sect addr
   `0x100000620` offset `0x620`; `0x100000970` → `0x970`. The *section* offset
   gives `0xf90` (wrong); `addr − 0x100000000` works only when fileoff is 0.
2. Verify: `xxd -s "$OFF" -l 4 "$T0"` prints `00000970: 0001 3fd6` (`blr x8`).
   No match → wrong offset, stop.
3. Write 4 bytes, one instruction: `printf '\x1f\x20\x03\xd5' | dd of="$T" bs=1
   seek="$OFF" conv=notrunc` (`nop`); 2 bytes leave half a word.
4. Diff **before** signing: `cmp -l "$T0" "$T" | wc -l` = the positions that actually
   changed, not the bytes you wrote — a write that repeats an original byte counts
   nothing (`nop` over 4 bytes where one matched prints `3`). Compare the offsets
   against your `seek` range; anything outside it is a wrong offset, stop. `codesign` rewrites
   the signature blob: after it, `466` + `EOF on rechal.orig`.
5. Re-sign, Mach-O only: `codesign --remove-signature "$T"; codesign -s - "$T"`.
6. Prove with the stop status, not `$?`:
   `lldb -b -o run -o quit -- ./rechal.orig x` → `exited with status = 45`;
   `./rechal` → `status = 1`, not 45, so deny-attach is dead, layer `partial`.
   Pass argv (`x`); with none the argc check exits `2` first.

## Proof

`solved` needs the artifact running and one of two proofs:
- **bytes** — the observable differs from `T.orig` and `cmp` shows exactly the patch;
- **replay** — a layer you defeat in the open (constant/string table, KDF) or
  `LAYERS: none` has no bytes to write: prove it with one replayed run (a known
  input, the observable your reading predicts, decoy rejected) and put that
  input→output in `EVIDENCE`. A decrypted plaintext counts only if a second
  sample decrypts with the same key.

The **exit code** separates decoy from real, never stdout: the artifact prints the
same `FLAG{…}` shape for a wrong input but exits non-zero, so rc=0 is the accepted
input. A stdout match with rc!=0 is a decoy — report it as such.
Everything else is `partial`, and `partial` quotes the exact blocker in one line.
