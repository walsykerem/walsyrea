#!/usr/bin/env node
"use strict";

const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { spawnSync } = require("node:child_process");

const root = path.join(__dirname, "..");
const bin = path.join(root, "bin");
const ext = process.platform === "win32" ? ".exe" : "";
const lreconCommit = "d477886d1d03d0537a1bbb2b0858b634f7ab8276";

function run(cmd, args, cwd) {
  const r = spawnSync(cmd, args, { cwd, stdio: "inherit" });
  if (r.error || r.status !== 0) {
    throw new Error(`${cmd} ${args.join(" ")} failed: ${r.error ? r.error.message : `exit ${r.status}`}`);
  }
}

if (spawnSync("go", ["version"], { stdio: "ignore" }).error) {
  console.error("walsyrea: Go 1.26+ is required to build from source");
  process.exit(1);
}

fs.mkdirSync(bin, { recursive: true });
run("go", ["build", "-trimpath", "-buildvcs=false", "-o", path.join(bin, `walsyrea-mcp${ext}`), "./cmd/walsyrea-mcp"], root);

const tmp = fs.mkdtempSync(path.join(os.tmpdir(), "walsyrea-lrecon-"));
try {
  run("git", ["clone", "--quiet", "https://github.com/walsykerem/lrecon.git", tmp], root);
  run("git", ["checkout", "--quiet", lreconCommit], tmp);
  run("go", ["build", "-trimpath", "-buildvcs=false", "-o", path.join(bin, `lrecon${ext}`), "./lrecon"], tmp);
} finally {
  fs.rmSync(tmp, { recursive: true, force: true });
}
