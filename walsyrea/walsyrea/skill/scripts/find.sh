#!/usr/bin/env bash
# Shim over the two catalog queries the binary lacks: compact list, find, budget.
set -euo pipefail
BIN=${LRECON:-lrecon}

case "${1:-}" in
--list-ops)
	if [[ "${2:-}" == --compact ]]; then
		"$BIN" --list-ops --json | python3 -c 'import json,sys
for r in json.load(sys.stdin):
    n=r.get("needs") or []
    print("%s\t%s\t%s" % (r["name"], r["route"], ",".join(n) if isinstance(n,list) else n))'
	else
		"$BIN" --list-ops
	fi
	;;
--find)
	"$BIN" --list-ops | grep -i -- "${2:?--find needs a pattern}"
	;;
--budget | --max-bytes)
	n=${2:?$1 needs a number}
	shift 2
	[[ "${1:-}" == -- ]] && shift
	exec "$BIN" --max-bytes "$n" "$@"
	;;
*)
	echo "usage: find.sh --list-ops [--compact] | --find PAT | --budget N -- ARGS" >&2
	exit 2
	;;
esac
