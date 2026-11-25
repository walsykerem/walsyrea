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
	if !ok {
		t.Fatalf("unexpected result: %#v", resp.Result)
	}
	info, ok := result["serverInfo"].(map[string]any)
	if !ok || info["name"] != serverName || info["version"] != "1.0.0" {
		t.Fatalf("unexpected serverInfo: %#v", result["serverInfo"])
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

func TestLreconPathResolutionOrder(t *testing.T) {
	root := t.TempDir()
	mkDir := func(name string) string {
		t.Helper()
		dir := filepath.Join(root, name)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	writeStub := func(dir, name string) string {
		t.Helper()
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, nil, 0o700); err != nil {
			t.Fatal(err)
		}
		return path
	}
	envDir, pathDir, exeDir, emptyDir := mkDir("env"), mkDir("path"), mkDir("exe"), mkDir("empty")
	name := "lrecon"
	if runtime.GOOS == "windows" {
		name = "lrecon.exe"
	}
	envPath := writeStub(envDir, name)
	pathPath := writeStub(pathDir, name)
	besidePath := writeStub(exeDir, name)

	realExecutable := executable
	t.Cleanup(func() { executable = realExecutable })

	cases := []struct {
		name, env, pathEnv, exeDir, want string
		wantErr                          bool
	}{
		{"env wins", envPath, pathDir, emptyDir, envPath, false},
		{"PATH wins", "", pathDir, emptyDir, pathPath, false},
		{"beside executable", "", "", exeDir, besidePath, false},
		{"nothing found", "", "", emptyDir, "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			executable = func() (string, error) { return filepath.Join(tc.exeDir, "walsyrea-mcp"), nil }
			t.Setenv("LRECON", tc.env)
			t.Setenv("PATH", tc.pathEnv)
			got, err := lreconPath()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("lreconPath() = %q, want error", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("lreconPath: %v", err)
			}
			if got != tc.want {
				t.Fatalf("lreconPath() = %q, want %q", got, tc.want)
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

func TestInitCacheClearsOnStart(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("WALSYREA_CACHE_DIR", dir)
	if err := os.WriteFile(filepath.Join(dir, "old.log"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, "stale"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitkeep"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	logFile, err := initCache()
	if err != nil {
		t.Fatal(err)
	}
	defer logFile.Close()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{".gitkeep": true, cacheLogName: true}
	for _, entry := range entries {
		if !want[entry.Name()] {
			t.Fatalf("unexpected cache entry %q", entry.Name())
		}
		delete(want, entry.Name())
	}
	if len(want) != 0 {
		t.Fatalf("missing cache entries: %v", want)
	}
}
