package main

import (
	"path/filepath"
	"strings"
	"testing"
)

const fixtureMod = "testdata/mod/example.com/"

func alpha() *module {
	return &module{Path: "example.com/alpha", Version: "v1.2.0", Dir: fixtureMod + "alpha@v1.2.0"}
}

func TestWriteGivesTheNoticesOfEachLinkedModule(t *testing.T) {
	t.Parallel()

	upper := &module{Path: "example.com/Upper", Version: "v0.1.0", Dir: fixtureMod + "!upper@v0.1.0"}
	pkgs := []listedPackage{
		{ImportPath: "fmt", Standard: true},
		{ImportPath: "example.com/Upper", Dir: upper.Dir, Module: upper},
		{ImportPath: "example.com/alpha", Dir: alpha().Dir, Module: alpha()},
		{ImportPath: "example.com/alpha/internal/copied", Dir: alpha().Dir + "/internal/copied", Module: alpha()},
		{ImportPath: "example.com/main", Dir: "/src", Module: &module{Path: "example.com/main", Main: true}},
	}

	var out strings.Builder
	if err := write(&out, pkgs, toolchain{Root: "testdata/goroot", Version: "go1.27.4"}); err != nil {
		t.Fatal(err)
	}

	want := header + `
================================================================================
Go standard library go1.27.4
Source: https://go.dev/dl/go1.27.4.src.tar.gz

--- LICENSE ---
Go licence

--- PATENTS ---
Go patents

================================================================================
example.com/Upper v0.1.0
Source: https://proxy.golang.org/example.com/!upper/@v/v0.1.0.zip

--- LICENSE.md ---
Upper licence

================================================================================
example.com/alpha v1.2.0
Source: https://proxy.golang.org/example.com/alpha/@v/v1.2.0.zip

--- LICENSE ---
Alpha licence

--- NOTICE.txt ---
Alpha notice

--- internal/copied/LICENSE ---
Copied licence
`
	if got := out.String(); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestWriteUsesTheReplacementOfAModule(t *testing.T) {
	t.Parallel()

	m := &module{Path: "example.com/original", Version: "v9.0.0", Replace: alpha()}
	pkgs := []listedPackage{{ImportPath: "example.com/original", Dir: alpha().Dir, Module: m}}

	var out strings.Builder
	if err := write(&out, pkgs, toolchain{Root: "testdata/goroot", Version: "go1.27.4"}); err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(
		out.String(),
		"\nexample.com/alpha v1.2.0\nSource: https://proxy.golang.org/example.com/alpha/@v/v1.2.0.zip\n",
	) {
		t.Errorf("the notices do not give the replacement module:\n%s", out.String())
	}
}

func TestWriteFailsWhenAModuleHasNoLicence(t *testing.T) {
	t.Parallel()

	bare := &module{Path: "example.com/bare", Version: "v1.0.0", Dir: filepath.FromSlash(fixtureMod + "bare@v1.0.0")}
	pkgs := []listedPackage{{ImportPath: "example.com/bare", Dir: bare.Dir, Module: bare}}

	err := write(&strings.Builder{}, pkgs, toolchain{Root: "testdata/goroot", Version: "go1.27.4"})
	if err == nil || !strings.Contains(err.Error(), "example.com/bare v1.0.0 has no licence file") {
		t.Errorf("got error %v, want an error that names the module with no licence file", err)
	}
}
