package main

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

const header = `THIRD-PARTY NOTICES

This file gives the licence notices of the Go standard library and of the third-party Go modules in the binary of
this image. Each entry gives the location of the source code of the software, at the version in this binary. The
binary contains the software with no changes. You can get the source code at that location at no charge. The web UIs
of the image give the notices of their npm packages in /licenses/web/.
`

const separator = "================================================================================"

// licenceFile also matches LICENSE.md, LICENSE-MIT and NOTICE.txt, but not a Go file such as notice_response.go.
var licenceFile = regexp.MustCompile(`(?i)^(licen[cs]e|copying|notice|patents)([-.](txt|md|mit|apache|bsd|[0-9.]+))*$`)

// listedPackage has the fields of go list -json that the notices need.
type listedPackage struct {
	ImportPath string  `json:"ImportPath"`
	Dir        string  `json:"Dir"`
	Standard   bool    `json:"Standard"`
	Module     *module `json:"Module"`
}

type module struct {
	Path    string  `json:"Path"`
	Version string  `json:"Version"`
	Dir     string  `json:"Dir"`
	Main    bool    `json:"Main"`
	Replace *module `json:"Replace"`
}

type toolchain struct {
	Root    string
	Version string
}

type entry struct {
	name   string
	source string
	dir    string
	// files are relative to dir.
	files []string
}

func write(w io.Writer, pkgs []listedPackage, tc toolchain) error {
	entries, err := moduleEntries(pkgs)
	if err != nil {
		return err
	}

	std := entry{
		name:   "Go standard library " + tc.Version,
		source: "https://go.dev/dl/" + tc.Version + ".src.tar.gz",
		dir:    tc.Root,
		files:  []string{"LICENSE", "PATENTS"},
	}

	if _, err := io.WriteString(w, header); err != nil {
		return err
	}

	for _, e := range append([]entry{std}, entries...) {
		if err := writeEntry(w, e); err != nil {
			return err
		}
	}

	return nil
}

func moduleEntries(pkgs []listedPackage) ([]entry, error) {
	byPath := map[string]*entry{}

	for _, p := range pkgs {
		if p.Standard || p.Module == nil || p.Module.Main {
			continue
		}

		m := p.Module
		if m.Replace != nil {
			m = m.Replace
		}

		e, ok := byPath[m.Path]
		if !ok {
			var err error
			if e, err = newEntry(m); err != nil {
				return nil, err
			}

			byPath[m.Path] = e
		}

		if err := e.addPackageFiles(p.Dir); err != nil {
			return nil, err
		}
	}

	entries := make([]entry, 0, len(byPath))
	for _, e := range byPath {
		slices.SortFunc(e.files, rootFilesFirst)
		entries = append(entries, *e)
	}

	slices.SortFunc(entries, func(a, b entry) int { return cmp.Compare(a.name, b.name) })

	return entries, nil
}

func newEntry(m *module) (*entry, error) {
	files, err := licenceFiles(m.Dir, ".")
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("module %s %s has no licence file in %s", m.Path, m.Version, m.Dir)
	}

	return &entry{name: m.Path + " " + m.Version, source: proxyURL(m), dir: m.Dir, files: files}, nil
}

// addPackageFiles adds the licence files of code from a different project in the package directory, or in a directory
// between it and the module root.
func (e *entry) addPackageFiles(pkgDir string) error {
	rel, err := filepath.Rel(e.dir, pkgDir)
	if err != nil {
		return fmt.Errorf("package directory %s: %w", pkgDir, err)
	}

	if strings.HasPrefix(rel, "..") {
		return nil
	}

	for ; rel != "."; rel = filepath.Dir(rel) {
		files, err := licenceFiles(e.dir, rel)
		if err != nil {
			return err
		}

		for _, f := range files {
			if !slices.Contains(e.files, f) {
				e.files = append(e.files, f)
			}
		}
	}

	return nil
}

func licenceFiles(moduleDir, rel string) ([]string, error) {
	dirEntries, err := os.ReadDir(filepath.Join(moduleDir, rel))
	if err != nil {
		return nil, err
	}

	var files []string

	for _, d := range dirEntries {
		if d.Type().IsRegular() && licenceFile.MatchString(d.Name()) {
			files = append(files, filepath.ToSlash(filepath.Join(rel, d.Name())))
		}
	}

	return files, nil
}

func rootFilesFirst(a, b string) int {
	return cmp.Or(cmp.Compare(strings.Count(a, "/"), strings.Count(b, "/")), cmp.Compare(a, b))
}

// proxyURL gives the source archive of the module on the Go module proxy, with the escaped path of the proxy protocol.
func proxyURL(m *module) string {
	var path strings.Builder

	for _, r := range m.Path {
		if unicode.IsUpper(r) {
			path.WriteByte('!')

			r = unicode.ToLower(r)
		}

		path.WriteRune(r)
	}

	return "https://proxy.golang.org/" + path.String() + "/@v/" + m.Version + ".zip"
}

func writeEntry(w io.Writer, e entry) error {
	if _, err := fmt.Fprintf(w, "\n%s\n%s\nSource: %s\n", separator, e.name, e.source); err != nil {
		return err
	}

	for _, f := range e.files {
		text, err := os.ReadFile(filepath.Join(e.dir, filepath.FromSlash(f)))
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintf(w, "\n--- %s ---\n%s", f, text); err != nil {
			return err
		}

		if len(text) > 0 && text[len(text)-1] != '\n' {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
	}

	return nil
}
