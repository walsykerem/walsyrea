#!/usr/bin/env node
"use strict";

const { spawn } = require("node:child_process");
const path = require("node:path");

const ext = process.platform === "win32" ? ".exe" : "";
const exe = path.join(__dirname, `walsyrea-mcp${ext}`);
const lrecon = path.join(__dirname, `lrecon${ext}`);
const env = { ...process.env };
if (!env.LRECON) env.LRECON = lrecon;

const child = spawn(exe, [], { stdio: "inherit", env });
child.on("error", (err) => {
  console.error(`walsyrea-mcp: ${err.message}; reinstall walsyrea so postinstall builds the binaries`);
  process.exit(1);
});
child.on("exit", (code, signal) => {
  if (signal) process.kill(process.pid, signal);
  else process.exit(code ?? 0);
});
