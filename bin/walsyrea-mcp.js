#!/usr/bin/env node
"use strict";

const { spawn } = require("node:child_process");
const path = require("node:path");

const arch = process.platform === "win32" || process.arch !== "x64" ? process.arch : "amd64";
const key = `${process.platform}-${arch}`;
const supported = new Set([
  "darwin-arm64", "darwin-amd64", "linux-amd64", "linux-arm64", "win32-x64", "win32-arm64",
]);
if (!supported.has(key)) {
  console.error(`walsyrea-mcp: unsupported platform ${key}`);
  process.exit(1);
}

const pkg = `@walsyrea/platform-${key}`;
let bin;
try {
  bin = path.join(path.dirname(require.resolve(`${pkg}/package.json`)), "bin");
} catch {
  console.error(`walsyrea-mcp: ${pkg} not installed; reinstall walsyrea for ${key}`);
  process.exit(1);
}
const exe = path.join(bin, process.platform === "win32" ? "walsyrea-mcp.exe" : "walsyrea-mcp");
const lrecon = path.join(bin, process.platform === "win32" ? "lrecon.exe" : "lrecon");
const env = { ...process.env };
if (!env.LRECON) env.LRECON = lrecon;

const child = spawn(exe, [], { stdio: "inherit", env });
child.on("error", (err) => {
  console.error("walsyrea-mcp:", err.message);
  process.exit(1);
});
child.on("exit", (code, signal) => {
  if (signal) process.kill(process.pid, signal);
  else process.exit(code ?? 0);
});
