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

func TestBundledLreconForPlatforms(t *testing.T) {
	cases := []struct {
		goos, goarch, want string
	}{
		{"darwin", "arm64", "macos/lrecon"},
		{"darwin", "amd64", "macos/lrecon-amd64"},
		{"linux", "amd64", "linux/lrecon"},
		{"linux", "arm64", "linux/lrecon-arm64"},
		{"windows", "amd64", "windows/lrecon.exe"},
		{"windows", "arm64", "windows/lrecon-arm64.exe"},
		{"freebsd", "amd64", ""},
	}
	for _, tc := range cases {
		t.Run(tc.goos+"_"+tc.goarch, func(t *testing.T) {
			want := ""
			if tc.want != "" {
				want = filepath.Join("walsyrea", "walsyrea", tc.want)
			}
			if got := bundledLreconFor(tc.goos, tc.goarch); got != want {
				t.Fatalf("bundledLreconFor(%q,%q) = %q, want %q", tc.goos, tc.goarch, got, want)
			}
		})
	}
}

func TestLreconPathPrefersLRECON(t *testing.T) {
	t.Setenv("LRECON", `C:\tools\lrecon.exe`)
	got, err := lreconPath()
	if err != nil {
		t.Fatalf("lreconPath: %v", err)
	}
	if got != `C:\tools\lrecon.exe` {
		t.Fatalf("lreconPath = %q, want env override", got)
	}
}
