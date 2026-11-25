package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestInitializeRoundTrip(t *testing.T) {
	var out bytes.Buffer
	line := append([]byte(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25"}}`), '\n')
	handleLine(context.Background(), json.NewEncoder(&out), line)

	var resp rpcResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	result, ok := resp.Result.(map[string]any)
	if !ok || result["serverInfo"] == nil {
		t.Fatalf("unexpected result: %#v", resp.Result)
	}
}

func TestRunLreconUsesLRECON(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell stub")
	}
	dir := t.TempDir()
	stub := filepath.Join(dir, "lrecon")
	script := "#!/bin/sh\nprintf '%s\\n' '{\"kind\":\"text\"}'\n"
	if err := os.WriteFile(stub, []byte(script), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LRECON", stub)

	out, err := runLrecon(context.Background(), nil, "--json", "target")
	if err != nil {
		t.Fatalf("run lrecon: %v", err)
	}
	if out != "{\"kind\":\"text\"}\n" {
		t.Fatalf("unexpected output: %q", out)
	}
}
