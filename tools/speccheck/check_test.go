package main

import (
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

func TestCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		fixture      string
		extra        fstest.MapFS
		wantErrors   []string
		wantWarnings []string
	}{
		{
			fixture: "valid",
			wantWarnings: []string{
				"pki/DESIGN.md:34: requirement PKI-0003 has no test that verifies it",
			},
		},
		{
			fixture: "design-directory",
		},
		{
			fixture: "duplicate-identifier",
			wantErrors: []string{
				"pki/DESIGN.md:14: identifier PKI-0001 is not unique: pki/DESIGN.md:9 also has it",
			},
			wantWarnings: []string{
				"pki/DESIGN.md:9: requirement PKI-0001 has no test that verifies it",
				"pki/DESIGN.md:10: requirement PKI-0002 has no test that verifies it",
			},
		},
		{
			fixture: "wrong-format",
			wantErrors: []string{
				"pki/DESIGN.md:9: the anchor of identifier PKI-0001 must be pki-0001",
				"pki/DESIGN.md:10: identifier PKI-002 does not match ^[A-Z]{2,5}-[0-9]{4}$",
				"pki/DESIGN.md:11: identifier REG-0003 does not have the code PKI of component pki",
				"pki/DESIGN.md:12: a list item in a group of requirements has no identifier",
				"pki/DESIGN.md:14: a requirement table must have ID as its first column and Source as its last column",
				"pki/DESIGN.md:16: a row of a requirement table has no identifier",
				"pki/DESIGN.md:20: a judgement heading has no identifier",
				"pki/DESIGN.md:20: the text refers to PKI-0005 with no link",
				"pki/DESIGN.md:26: judgement PKI-0006 must have the items Kind, Basis and Reason in their formats",
				"pki/DESIGN.md:34: withdrawn entry PKI-0007 does not match the format",
				"services/register/DESIGN.md:5: the file has identifiers but no Withdrawn section",
				"tools/example/DESIGN.md:1: component tools/example has no code in docs/spec/conventions.md",
			},
		},
		{
			fixture: "sub-heading",
			wantErrors: []string{
				"pki/DESIGN.md:11: only a judgement can have a heading below ##, in the Judgements section",
				"pki/DESIGN.md:17: only a judgement can have a heading below ##, in the Judgements section",
				"pki/DESIGN.md:29: only a judgement can have a heading below ##, in the Judgements section",
			},
			wantWarnings: []string{
				"pki/DESIGN.md:19: requirement PKI-0001 has no test that verifies it",
			},
		},
		{
			fixture: "no-source",
			wantErrors: []string{
				"pki/DESIGN.md:9: requirement PKI-0001 has no source",
				"pki/DESIGN.md:10: requirement PKI-0002 has no source link",
				"pki/DESIGN.md:11: the reference link label accc-cp-missing has no definition",
				"pki/DESIGN.md:15: requirement PKI-0004 has no source link",
			},
			wantWarnings: []string{
				"pki/DESIGN.md:9: requirement PKI-0001 has no test that verifies it",
				"pki/DESIGN.md:10: requirement PKI-0002 has no test that verifies it",
				"pki/DESIGN.md:11: requirement PKI-0003 has no test that verifies it",
				"pki/DESIGN.md:15: requirement PKI-0004 has no test that verifies it",
			},
		},
		{
			fixture: "broken-link",
			wantErrors: []string{
				"docs/adr/0001-example.md:3: the link goes to pki/DESIGN.md#pki-0008, which is not an identifier",
				"docs/adr/0001-example.md:5: the link goes to pki/DESIGN.md#pki-0007, which is not an identifier",
				"pki/DESIGN.md:9: the link goes to pki/DESIGN.md#pki-0009, which is not an identifier",
				"pki/DESIGN.md:13: the link goes to services/register/DESIGN.md#reg-0001, which is not an identifier",
			},
			wantWarnings: []string{
				"pki/DESIGN.md:9: requirement PKI-0001 has no test that verifies it",
			},
		},
		{
			fixture: "bare-identifier",
			wantErrors: []string{
				"pki/DESIGN.md:9: the text refers to PKI-0001 with no link",
				"pki/DESIGN.md:9: the text refers to PKI-0002 with no link",
				"pki/DESIGN.md:12: the link text is only the identifier PKI-0001",
			},
			wantWarnings: []string{
				"pki/DESIGN.md:11: requirement PKI-0001 has no test that verifies it",
				"pki/DESIGN.md:12: requirement PKI-0002 has no test that verifies it",
			},
		},
		{
			fixture: "word-limit",
			extra: fstest.MapFS{
				"pki/DESIGN.md": {Data: []byte("# Mock CDR PKI\n\n" + strings.Repeat("word ", 4990) +
					"\n\n## Library specifications\n\nNone.\n\n## Withdrawn\n\nNone.\n")},
			},
			wantErrors: []string{
				"pki/DESIGN.md:1: the file has 5001 words; the limit is 5000",
			},
		},
		{
			fixture: "library-table",
			wantErrors: []string{
				"pki/DESIGN.md:3: component pki imports pkg/platform/database, but the table does not list it",
				"pki/DESIGN.md:8: the table lists pkg/cds, but component pki does not import it",
				"pki/DESIGN.md:9: a row of the library table does not start with a pkg/ path in a code span",
				"web/packages/ui/DESIGN.md:1: the file has no Library specifications section",
			},
			wantWarnings: []string{
				"web/packages/ui/DESIGN.md:5: requirement UI-0001 has no test that verifies it",
			},
		},
		{
			fixture: "verifies",
			wantErrors: []string{
				"pki/issue_test.go:6: the Verifies comment refers to PKI-0009, which does not exist",
				"pki/issue_test.go:8: the Verifies comment refers to PKI-0002, which is a judgement",
				"pki/issue_test.go:10: the Verifies comment refers to PKI-0003, which is withdrawn",
				"web/packages/ui/src/button.spec.ts:1: the Verifies comment does not match the format",
				"web/packages/ui/src/button.spec.ts:4: the Verifies comment does not match the format",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.fixture, func(t *testing.T) {
			t.Parallel()

			fsys, paths := loadFixture(t, tt.fixture, tt.extra)
			got := check(fsys, paths)

			assertFindings(t, "errors", got.errors, tt.wantErrors)
			assertFindings(t, "warnings", got.warnings, tt.wantWarnings)
		})
	}
}

func TestCheckWithNoRegistry(t *testing.T) {
	t.Parallel()

	got := check(fstest.MapFS{}, nil)

	assertFindings(t, "errors", got.errors, []string{
		"docs/spec/conventions.md:1: the file has no table of component codes",
	})
}

func loadFixture(t *testing.T, name string, extra fstest.MapFS) (fstest.MapFS, []string) {
	t.Helper()

	fsys := fstest.MapFS{}

	conventions, err := os.ReadFile(filepath.Join("testdata", "conventions.md"))
	if err != nil {
		t.Fatal(err)
	}

	fsys[conventionsPath] = &fstest.MapFile{Data: conventions}

	root := filepath.Join("testdata", name)
	if _, err := os.Stat(root); err == nil {
		fixture := os.DirFS(root)

		err = fs.WalkDir(fixture, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}

			data, err := fs.ReadFile(fixture, path)
			fsys[path] = &fstest.MapFile{Data: data}

			return err
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	maps.Copy(fsys, extra)

	paths := make([]string, 0, len(fsys))
	for path := range fsys {
		paths = append(paths, path)
	}

	slices.Sort(paths)

	return fsys, paths
}

func assertFindings(t *testing.T, kind string, got []finding, want []string) {
	t.Helper()

	lines := make([]string, len(got))
	for i, f := range got {
		lines[i] = f.String()
	}

	if !slices.Equal(lines, want) {
		t.Errorf("%s:\ngot:\n%s\nwant:\n%s", kind, strings.Join(lines, "\n"), strings.Join(want, "\n"))
	}
}

func TestSkipsTestdata(t *testing.T) {
	t.Parallel()

	got := filterPaths([]string{"pki/DESIGN.md", "tools/speccheck/testdata/valid/pki/DESIGN.md", "testdata/x_test.go"})

	if !slices.Equal(got, []string{"pki/DESIGN.md"}) {
		t.Errorf("got %v", got)
	}
}
