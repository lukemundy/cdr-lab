package main

import (
	"cmp"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"maps"
	"path"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	conventionsPath = "docs/spec/conventions.md"
	modulePath      = "github.com/lukemundy/cdr-lab"
	wordLimit       = 5000
	idPattern       = `^[A-Z]{2,5}-[0-9]{4}$`
)

// The formats of docs/spec/conventions.md. The prefix expressions find the identifier first, so that the check can
// tell a requirement with a wrong identifier from a requirement with no source.
var (
	idRe              = regexp.MustCompile(idPattern)
	listItemPrefixRe  = regexp.MustCompile(`^- <a id="([^"]*)"></a>\*\*([^*]*)\*\*(.*)$`)
	ruleAndSourcesRe  = regexp.MustCompile(`^ (.+\.) \((.+)\)$`)
	tableIDCellRe     = regexp.MustCompile(`^<a id="([^"]*)"></a>\*\*([^*]*)\*\*$`)
	judgementPrefixRe = regexp.MustCompile(`^### <a id="([^"]*)"></a>(\S+) (.+)$`)
	withdrawnRestRe   = regexp.MustCompile(`^ (Moved to \[.+\]\(.+#[a-z]{2,5}-[0-9]{4}\)|Removed)\.$`)
	kindRe            = regexp.MustCompile(`^- Kind: (choice|conflict|gap)$`)
	basisRe           = regexp.MustCompile(`^- Basis: Step (1|2|3\.1|3\.2|3\.3)\. .+$`)
	reasonRe          = regexp.MustCompile(`^- Reason: .+$`)
	libraryRowRe      = regexp.MustCompile("^`(pkg/[^`]+)`$")
	registryRowRe     = regexp.MustCompile("^\\| `([^`]+)` \\| `([A-Z]{2,5})` \\|$")
	inlineLinkRe      = regexp.MustCompile(`\[([^\]]*)\]\(([^)\s]*)\)`)
	referenceLinkRe   = regexp.MustCompile(`\[([^\]]*)\]\[([^\]]*)\]`)
	linkDefinitionRe  = regexp.MustCompile(`^\[([^\]]+)\]:\s+(\S+)`)
	anyLinkRe         = regexp.MustCompile(`\[[^\]]+\](\([^)]+\)|\[[^\]]+\])`)
	idFragmentRe      = regexp.MustCompile(`^[a-z]{2,5}-[0-9]{4}$`)
	codeSpanRe        = regexp.MustCompile("`[^`]*`")
	definitionSiteRe  = regexp.MustCompile(`<a id="[^"]*"></a>(\*\*[^*]+\*\*|[A-Z]{2,5}-[0-9]{4})`)
	verifiesRe        = regexp.MustCompile(`^\s*// Verifies: [A-Z]{2,5}-[0-9]{4}(, [A-Z]{2,5}-[0-9]{4})*$`)
	verifiesIDRe      = regexp.MustCompile(`[A-Z]{2,5}-[0-9]{4}`)
	subHeadingRe      = regexp.MustCompile(`^#{3,6}(\s|$)`)
)

var fixedHeadings = map[string]bool{
	"Context":                true,
	"Areas":                  true,
	"Library specifications": true,
	"Goals":                  true,
	"Non-goals":              true,
	"Judgements":             true,
	"Withdrawn":              true,
}

type finding struct {
	path string
	line int
	msg  string
}

func (f finding) String() string { return fmt.Sprintf("%s:%d: %s", f.path, f.line, f.msg) }

type result struct {
	errors   []finding
	warnings []finding
}

type kind int

const (
	requirement kind = iota
	judgement
	withdrawn
)

type definition struct {
	id   string
	kind kind
	path string
	line int
}

type libraryRow struct {
	library string
	line    int
}

// spec is one specification file.
type spec struct {
	path      string
	component string
	code      string
	lines     []string
	// libraryLine is 0 when the file has no Library specifications section.
	libraryLine int
	libraries   []libraryRow
}

type checker struct {
	fsys       fs.FS
	result     result
	codes      map[string]string // component directory to code
	specs      []*spec
	ids        map[string]definition // identifier to its definition
	anchors    map[string]bool       // "path#anchor" of each definition
	titles     map[string]string     // "path#anchor" of each judgement: its title
	verified   map[string]bool
	bareIDRe   *regexp.Regexp
	components map[string][]*spec
}

func (c *checker) errorf(file string, line int, format string, args ...any) {
	c.result.errors = append(c.result.errors, finding{file, line, fmt.Sprintf(format, args...)})
}

func (c *checker) warnf(file string, line int, format string, args ...any) {
	c.result.warnings = append(c.result.warnings, finding{file, line, fmt.Sprintf(format, args...)})
}

// check reads paths relative to the repository root.
func check(fsys fs.FS, paths []string) result {
	c := &checker{
		fsys:       fsys,
		ids:        map[string]definition{},
		anchors:    map[string]bool{},
		titles:     map[string]string{},
		verified:   map[string]bool{},
		components: map[string][]*spec{},
	}

	if !c.readRegistry() {
		return c.result
	}

	for _, p := range paths {
		if component, ok := specComponent(p); ok {
			c.readSpec(p, component)
		}
	}

	for _, s := range c.specs {
		c.checkLines(s)
	}

	for _, p := range paths {
		switch {
		case strings.HasSuffix(p, ".md") && p != conventionsPath:
			c.checkLinkTargets(p)
		case isTestFile(p):
			c.checkVerifies(p)
		}
	}

	c.checkLibraries(paths)
	c.warnUnverified()

	sortFindings(c.result.errors)
	sortFindings(c.result.warnings)

	return c.result
}

func sortFindings(findings []finding) {
	slices.SortStableFunc(findings, func(a, b finding) int {
		return cmp.Or(cmp.Compare(a.path, b.path), cmp.Compare(a.line, b.line))
	})
}

// filterPaths removes the test fixtures, so that a fixture with errors does not fail the check of the repository.
func filterPaths(paths []string) []string {
	return slices.DeleteFunc(slices.Clone(paths), func(p string) bool {
		return slices.Contains(strings.Split(p, "/"), "testdata")
	})
}

// specComponent gives the component of a specification file: <component>/DESIGN.md or <component>/design/*.md.
func specComponent(p string) (string, bool) {
	dir, base := path.Split(p)
	dir = strings.TrimSuffix(dir, "/")

	switch {
	case base == "DESIGN.md" && dir != "":
		return dir, true
	case path.Base(dir) == "design" && strings.HasSuffix(base, ".md") && path.Dir(dir) != ".":
		return path.Dir(dir), true
	default:
		return "", false
	}
}

func isTestFile(p string) bool {
	for _, suffix := range []string{"_test.go", ".test.ts", ".test.tsx", ".spec.ts", ".spec.tsx"} {
		if strings.HasSuffix(p, suffix) {
			return true
		}
	}

	return false
}

func (c *checker) readLines(p string) ([]string, bool) {
	data, err := fs.ReadFile(c.fsys, p)
	if err != nil {
		c.errorf(p, 1, "cannot read the file: %v", err)

		return nil, false
	}

	return strings.Split(string(data), "\n"), true
}

func (c *checker) readRegistry() bool {
	data, err := fs.ReadFile(c.fsys, conventionsPath)
	if err != nil {
		c.errorf(conventionsPath, 1, "the file has no table of component codes")

		return false
	}

	c.codes = map[string]string{}
	inSection := false

	for line := range strings.Lines(string(data)) {
		line = strings.TrimSuffix(line, "\n")
		if strings.HasPrefix(line, "#") {
			inSection = line == "### Component codes"

			continue
		}

		if m := registryRowRe.FindStringSubmatch(line); inSection && m != nil {
			c.codes[m[1]] = m[2]
		}
	}

	if len(c.codes) == 0 {
		c.errorf(conventionsPath, 1, "the file has no table of component codes")

		return false
	}

	codes := make([]string, 0, len(c.codes))
	for _, code := range c.codes {
		codes = append(codes, code)
	}

	slices.Sort(codes)
	c.bareIDRe = regexp.MustCompile(`\b(` + strings.Join(codes, "|") + `)-[0-9]{4}\b`)

	return true
}

// outsideFences calls fn for each line that is not in a fenced code block, with its line number.
func outsideFences(lines []string, fn func(n int, line string)) {
	fence := ""

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if fence != "" {
			if strings.HasPrefix(trimmed, fence) {
				fence = ""
			}

			continue
		}

		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fence = trimmed[:3]

			continue
		}

		fn(i+1, line)
	}
}

// readSpec records the definitions of a specification file and checks its structure.
func (c *checker) readSpec(p, component string) {
	lines, ok := c.readLines(p)
	if !ok {
		return
	}

	code, ok := c.codes[component]
	if !ok {
		c.errorf(p, 1, "component %s has no code in %s", component, conventionsPath)

		return
	}

	s := &spec{path: p, component: component, code: code, lines: lines}
	c.specs = append(c.specs, s)
	c.components[component] = append(c.components[component], s)

	if words := len(strings.Fields(strings.Join(lines, "\n"))); words > wordLimit {
		c.errorf(p, 1, "the file has %d words; the limit is %d", words, wordLimit)
	}

	c.readStructure(s)

	isOverview := path.Base(p) == "DESIGN.md" || (path.Base(path.Dir(p)) == "design" && path.Base(p) == "README.md")
	if !isOverview {
		return
	}

	if s.libraryLine == 0 {
		c.errorf(p, 1, "the file has no Library specifications section")
	}

	hasWithdrawn := false

	outsideFences(lines, func(_ int, line string) {
		hasWithdrawn = hasWithdrawn || line == "## Withdrawn"
	})

	if !hasWithdrawn {
		c.errorf(p, 1, "the file has no Withdrawn section")
	}
}

// readStructure reads the sections of a file: the requirements, the judgements, the withdrawn list and the
// library table.
func (c *checker) readStructure(s *spec) {
	section := ""
	judgementID := ""
	judgementItems := []string{}
	judgementLine := 0
	inTableBody := false

	endJudgement := func() {
		if judgementID != "" && !validJudgementItems(judgementItems) {
			c.errorf(s.path, judgementLine, "judgement %s must have the items Kind, Basis and Reason in their formats",
				judgementID)
		}

		judgementID, judgementItems = "", nil
	}

	outsideFences(s.lines, func(n int, line string) {
		if !strings.HasPrefix(line, "|") {
			inTableBody = false
		}

		switch {
		case strings.HasPrefix(line, "## "):
			endJudgement()

			section = strings.TrimPrefix(line, "## ")
			if section == "Library specifications" {
				s.libraryLine = n
			}

			return
		case strings.HasPrefix(line, "### ") && section == "Judgements":
			endJudgement()

			m := judgementPrefixRe.FindStringSubmatch(line)
			if m == nil {
				c.errorf(s.path, n, "a judgement heading has no identifier")

				return
			}

			if c.define(s, n, m[1], m[2], judgement) {
				judgementID, judgementLine = m[2], n
				c.titles[s.path+"#"+m[1]] = m[3]
			}

			return
		case subHeadingRe.MatchString(line):
			c.errorf(s.path, n, "only a judgement can have a heading below ##, in the Judgements section")

			return
		}

		switch {
		case section == "Judgements":
			if judgementID != "" && strings.HasPrefix(line, "- ") {
				judgementItems = append(judgementItems, line)
			}
		case section == "Withdrawn":
			c.readWithdrawn(s, n, line)
		case section == "Library specifications":
			c.readLibraryRow(s, n, line)
		case !fixedHeadings[section] && section != "":
			if strings.HasPrefix(line, "|") {
				c.readTableRow(s, n, line, &inTableBody)
			} else if strings.HasPrefix(line, "- ") {
				c.readListItem(s, n, line)
			}
		}
	})

	endJudgement()
}

func validJudgementItems(items []string) bool {
	return len(items) == 3 && kindRe.MatchString(items[0]) && basisRe.MatchString(items[1]) &&
		reasonRe.MatchString(items[2])
}

// define records an identifier. It returns false, with an error, for a wrong format or a duplicate.
func (c *checker) define(s *spec, n int, anchor, id string, k kind) bool {
	switch {
	case !idRe.MatchString(id):
		c.errorf(s.path, n, "identifier %s does not match %s", id, idPattern)

		return false
	case anchor != strings.ToLower(id):
		c.errorf(s.path, n, "the anchor of identifier %s must be %s", id, strings.ToLower(id))

		return false
	case !strings.HasPrefix(id, s.code+"-"):
		c.errorf(s.path, n, "identifier %s does not have the code %s of component %s", id, s.code, s.component)

		return false
	}

	if first, ok := c.ids[id]; ok {
		c.errorf(s.path, n, "identifier %s is not unique: %s:%d also has it", id, first.path, first.line)

		return false
	}

	c.ids[id] = definition{id: id, kind: k, path: s.path, line: n}
	c.anchors[s.path+"#"+anchor] = true

	return true
}

func (c *checker) readListItem(s *spec, n int, line string) {
	m := listItemPrefixRe.FindStringSubmatch(line)
	if m == nil {
		c.errorf(s.path, n, "a list item in a group of requirements has no identifier")

		return
	}

	if !c.define(s, n, m[1], m[2], requirement) {
		return
	}

	parts := ruleAndSourcesRe.FindStringSubmatch(m[3])
	switch {
	case parts == nil:
		c.errorf(s.path, n, "requirement %s has no source", m[2])
	case !anyLinkRe.MatchString(parts[2]):
		c.errorf(s.path, n, "requirement %s has no source link", m[2])
	}
}

func tableCells(line string) []string {
	cells := strings.Split(strings.Trim(strings.TrimSpace(line), "|"), "|")
	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}

	return cells
}

func isSeparatorRow(cells []string) bool {
	return !slices.ContainsFunc(cells, func(cell string) bool { return strings.Trim(cell, "-: ") != "" })
}

// readTableRow reads one row of a requirement table. inBody is false until the header and the separator row of the
// table have been read.
func (c *checker) readTableRow(s *spec, n int, line string, inBody *bool) {
	cells := tableCells(line)

	switch {
	case isSeparatorRow(cells):
		*inBody = true

		return
	case !*inBody:
		if cells[0] != "ID" || cells[len(cells)-1] != "Source" {
			c.errorf(s.path, n, "a requirement table must have ID as its first column and Source as its last column")
		}

		return
	}

	m := tableIDCellRe.FindStringSubmatch(cells[0])
	if m == nil {
		c.errorf(s.path, n, "a row of a requirement table has no identifier")

		return
	}

	if c.define(s, n, m[1], m[2], requirement) && !anyLinkRe.MatchString(cells[len(cells)-1]) {
		c.errorf(s.path, n, "requirement %s has no source link", m[2])
	}
}

func (c *checker) readWithdrawn(s *spec, n int, line string) {
	if !strings.HasPrefix(line, "- ") {
		return
	}

	m := listItemPrefixRe.FindStringSubmatch(line)
	if m == nil {
		c.errorf(s.path, n, "a withdrawn entry has no identifier")

		return
	}

	if c.define(s, n, m[1], m[2], withdrawn) && !withdrawnRestRe.MatchString(m[3]) {
		c.errorf(s.path, n, "withdrawn entry %s does not match the format", m[2])
	}
}

func (c *checker) readLibraryRow(s *spec, n int, line string) {
	if !strings.HasPrefix(line, "|") {
		return
	}

	cells := tableCells(line)
	if isSeparatorRow(cells) || cells[0] == "Library" {
		return
	}

	m := libraryRowRe.FindStringSubmatch(cells[0])
	if m == nil {
		c.errorf(s.path, n, "a row of the library table does not start with a pkg/ path in a code span")

		return
	}

	s.libraries = append(s.libraries, libraryRow{library: m[1], line: n})
}

// checkLines finds bare identifiers and undefined reference link labels in a specification file.
func (c *checker) checkLines(s *spec) {
	labels := map[string]bool{}

	outsideFences(s.lines, func(_ int, line string) {
		if m := linkDefinitionRe.FindStringSubmatch(line); m != nil {
			labels[strings.ToLower(m[1])] = true
		}
	})

	outsideFences(s.lines, func(n int, line string) {
		if linkDefinitionRe.MatchString(line) {
			return
		}

		text := codeSpanRe.ReplaceAllString(line, "")
		text = definitionSiteRe.ReplaceAllString(text, "")

		for _, m := range referenceLinkRe.FindAllStringSubmatch(text, -1) {
			label := cmp.Or(m[2], m[1])
			if !labels[strings.ToLower(label)] {
				c.errorf(s.path, n, "the reference link label %s has no definition", label)
			}
		}

		text = inlineLinkRe.ReplaceAllStringFunc(text, func(link string) string {
			linkText := inlineLinkRe.FindStringSubmatch(link)[1]
			if c.bareIDRe.FindString(linkText) == strings.TrimSpace(linkText) && linkText != "" {
				c.errorf(s.path, n, "the link text is only the identifier %s", strings.TrimSpace(linkText))

				return ""
			}

			return linkText
		})

		for _, id := range c.bareIDRe.FindAllString(text, -1) {
			c.errorf(s.path, n, "the text refers to %s with no link", id)
		}
	})
}

// checkLinkTargets makes sure that each link to an identifier, from any Markdown file, goes to an identifier that
// exists.
func (c *checker) checkLinkTargets(p string) {
	lines, ok := c.readLines(p)
	if !ok {
		return
	}

	outsideFences(lines, func(n int, line string) {
		text := withoutLinksInCodeSpans(line)

		type link struct{ text, target string }

		links := []link{}
		for _, m := range inlineLinkRe.FindAllStringSubmatch(text, -1) {
			links = append(links, link{m[1], m[2]})
		}

		if m := linkDefinitionRe.FindStringSubmatch(text); m != nil {
			links = append(links, link{"", m[2]})
		}

		for _, l := range links {
			file, fragment, found := strings.Cut(l.target, "#")
			if !found || !idFragmentRe.MatchString(fragment) || strings.Contains(file, ":") {
				continue
			}

			if file == "" {
				file = p
			} else {
				file = path.Join(path.Dir(p), file)
			}

			if !c.anchors[file+"#"+fragment] {
				c.errorf(p, n, "the link goes to %s#%s, which is not an identifier", file, fragment)

				continue
			}

			if linkTitle, ok := strings.CutPrefix(l.text, "Judgement: "); ok {
				c.checkJudgementLink(p, n, file+"#"+fragment, linkTitle)
			}
		}
	})
}

// checkJudgementLink makes sure that a Judgement: link goes to a judgement and has its title as text. The case of
// the first letter can differ, because the title starts a heading and the link text is in a sentence.
func (c *checker) checkJudgementLink(p string, n int, target, linkTitle string) {
	id := strings.ToUpper(target[strings.LastIndex(target, "#")+1:])

	title, ok := c.titles[target]
	if !ok {
		c.errorf(p, n, "the Judgement: link goes to %s, which is not a judgement", id)

		return
	}

	if lowerFirst(linkTitle) != lowerFirst(withoutLinksInCodeSpans(title)) {
		c.errorf(p, n, "the link text %q is not the title of judgement %s: %q", linkTitle, id, title)
	}
}

// withoutLinksInCodeSpans removes the brackets and parentheses from each code span, so that a code span is never a
// link but stays in the link text that contains it.
func withoutLinksInCodeSpans(s string) string {
	return codeSpanRe.ReplaceAllStringFunc(s, func(span string) string {
		return strings.Map(func(r rune) rune {
			if strings.ContainsRune("[]()", r) {
				return -1
			}

			return r
		}, span)
	})
}

func lowerFirst(s string) string {
	r, size := utf8.DecodeRuneInString(s)

	return string(unicode.ToLower(r)) + s[size:]
}

func (c *checker) checkVerifies(p string) {
	lines, ok := c.readLines(p)
	if !ok {
		return
	}

	for i, line := range lines {
		if !strings.HasPrefix(strings.TrimSpace(line), "// Verifies") {
			continue
		}

		if !verifiesRe.MatchString(line) {
			c.errorf(p, i+1, "the Verifies comment does not match the format")

			continue
		}

		for _, id := range verifiesIDRe.FindAllString(line, -1) {
			def, ok := c.ids[id]

			switch {
			case !ok:
				c.errorf(p, i+1, "the Verifies comment refers to %s, which does not exist", id)
			case def.kind == judgement:
				c.errorf(p, i+1, "the Verifies comment refers to %s, which is a judgement", id)
			case def.kind == withdrawn:
				c.errorf(p, i+1, "the Verifies comment refers to %s, which is withdrawn", id)
			default:
				c.verified[id] = true
			}
		}
	}
}

// checkLibraries compares the library table of each overview with the direct pkg/ imports of the non-test Go files
// of the component.
func (c *checker) checkLibraries(paths []string) {
	for component, specs := range c.components {
		imports := c.componentImports(component, paths)

		for _, s := range specs {
			if s.libraryLine == 0 {
				continue
			}

			listed := map[string]bool{}

			for _, row := range s.libraries {
				listed[row.library] = true

				if !imports[row.library] {
					c.errorf(s.path, row.line, "the table lists %s, but component %s does not import it", row.library,
						component)
				}
			}

			for _, library := range slices.Sorted(maps.Keys(imports)) {
				if !listed[library] {
					c.errorf(
						s.path,
						s.libraryLine,
						"component %s imports %s, but the table does not list it",
						component,
						library,
					)
				}
			}
		}
	}
}

func (c *checker) componentImports(component string, paths []string) map[string]bool {
	imports := map[string]bool{}

	for _, p := range paths {
		if !strings.HasPrefix(p, component+"/") || !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			continue
		}

		data, err := fs.ReadFile(c.fsys, p)
		if err != nil {
			c.errorf(p, 1, "cannot read the file: %v", err)

			continue
		}

		file, err := parser.ParseFile(token.NewFileSet(), p, data, parser.ImportsOnly)
		if err != nil {
			c.errorf(p, 1, "cannot read the imports: %v", err)

			continue
		}

		for _, imp := range file.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}

			if library, ok := libraryOf(importPath); ok && library != component {
				imports[library] = true
			}
		}
	}

	return imports
}

// libraryOf maps an import path to its library: pkg/<name>, or pkg/platform/<name>.
func libraryOf(importPath string) (string, bool) {
	rest, ok := strings.CutPrefix(importPath, modulePath+"/pkg/")
	if !ok {
		return "", false
	}

	parts := strings.Split(rest, "/")
	if parts[0] == "platform" && len(parts) > 1 {
		return "pkg/platform/" + parts[1], true
	}

	return "pkg/" + parts[0], true
}

func (c *checker) warnUnverified() {
	for _, def := range c.ids {
		if def.kind == requirement && !c.verified[def.id] {
			c.warnf(def.path, def.line, "requirement %s has no test that verifies it", def.id)
		}
	}
}
