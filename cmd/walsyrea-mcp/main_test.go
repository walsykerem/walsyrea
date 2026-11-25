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

func TestPlatformLreconName(t *testing.T) {
	cases := []struct {
		goos, goarch, want string
	}{
		{"darwin", "arm64", "lrecon-arm64"},
		{"darwin", "amd64", "lrecon"},
		{"linux", "amd64", "lrecon"},
		{"linux", "arm64", "lrecon-arm64"},
		{"windows", "amd64", "lrecon.exe"},
		{"windows", "arm64", "lrecon-arm64.exe"},
		{"freebsd", "amd64", ""},
	}
	for _, tc := range cases {
		if got := platformLreconName(tc.goos, tc.goarch); got != tc.want {
			t.Errorf("platformLreconName(%q,%q) = %q, want %q", tc.goos, tc.goarch, got, tc.want)
		}
	}
}

func TestLreconPathResolutionOrder(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(cwd); err != nil {
			t.Error(err)
		}
	})

	envDir := filepath.Join(root, "env")
	pathDir := filepath.Join(root, "path")
	platformDir := filepath.Join(root, "platform")
	emptyDir := filepath.Join(root, "empty")
	for _, dir := range []string{envDir, pathDir, platformDir, emptyDir} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	writeStub := func(dir, name string) string {
		t.Helper()
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, nil, 0o700); err != nil {
			t.Fatal(err)
		}
		return path
	}

	exeName := "lrecon"
	if runtime.GOOS == "windows" {
		exeName = "lrecon.exe"
	}
	envPath := writeStub(envDir, exeName)
	pathPath := writeStub(pathDir, exeName)
	bundledPath := bundledLrecon()

	platformPath := ""
	if name := platformLreconName(runtime.GOOS, runtime.GOARCH); name != "" {
		platformPath = writeStub(platformDir, name)
	}

	cases := []struct {
		name, env, dirPath, platformDir, want string
		bundled, wantErr                      bool
	}{
		{"env wins", envPath, pathDir, platformDir, envPath, true, false},
		{"PATH wins", "", pathDir, platformDir, pathPath, true, false},
		{"platform before bundled", "", "", platformDir, platformPath, true, false},
		{"bundled last", "", "", emptyDir, bundledPath, true, false},
		{"nothing found", "", "", emptyDir, "", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantErr == false && tc.want == "" {
				t.Skip("unsupported platform")
			}
			_ = os.Remove(bundledPath)
			if tc.bundled {
				writeStub(filepath.Dir(bundledPath), filepath.Base(bundledPath))
			}
			t.Setenv("LRECON", tc.env)
			t.Setenv("PATH", tc.dirPath)
			t.Setenv("WALSYREA_PLATFORM_DIR", tc.platformDir)
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
