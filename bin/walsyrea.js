#!/usr/bin/env node
"use strict";

const path = require("node:path");

if (process.argv[2] !== "--skill-path") {
  console.error("usage: walsyrea --skill-path");
  process.exit(2);
}
console.log(path.join(__dirname, "..", "walsyrea", "walsyrea", "skill"));
