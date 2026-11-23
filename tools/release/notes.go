package main

import (
	_ "embed"
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

// writeNotes writes the release notes. A chart with no digest is not published yet.
func writeNotes(w io.Writer, p releasePlan, chartDigests map[string]string) error {
	charts := make([]notesChart, 0, len(p.Charts))

	for _, name := range p.Charts {
		charts = append(charts, notesChart{Name: name, Digest: chartDigests[name]})
	}

	data := struct {
		releasePlan
		Charts []notesChart
	}{p, charts}

	data.Changelog = strings.TrimSpace(data.Changelog) + "\n"

	return notesTemplate.Execute(w, data)
}
