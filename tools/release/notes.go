package main

import (
	_ "embed"
	"fmt"
	"io"
	"strings"
	"text/template"
)

//go:embed notes.md.tmpl
var notesText string

var notesTemplate = template.Must(template.New("notes").Parse(notesText))

type notesChart struct {
	Name   string
	Digest string
}

// writeNotes writes the release notes. A run that publishes the charts must give the digest of each chart.
func writeNotes(w io.Writer, p releasePlan, chartDigests map[string]string) error {
	charts := make([]notesChart, 0, len(p.Charts))

	for _, name := range p.Charts {
		digest := chartDigests[name]
		if digest == "" && p.Mode != modeCheck {
			return fmt.Errorf("no digest for the chart %s", name)
		}

		charts = append(charts, notesChart{Name: name, Digest: digest})
	}

	data := struct {
		releasePlan
		Charts []notesChart
	}{p, charts}

	data.Changelog = strings.TrimSpace(data.Changelog) + "\n"

	return notesTemplate.Execute(w, data)
}
