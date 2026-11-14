// Package brands has the preset brands: the brand directories that the data holder contains. Each directory of this
// package is one preset, and its name is the name of the preset. README.md tells how to add a brand.
package brands

import (
	"embed"
	"fmt"
	"io/fs"
	"slices"
	"strings"
)

//go:embed */*
var presets embed.FS

// Names gives the name of each preset brand, in order.
func Names() []string {
	entries, err := presets.ReadDir(".")
	if err != nil {
		panic(err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names
}

// Preset gives the brand directory of the preset.
func Preset(name string) (fs.FS, error) {
	if !slices.Contains(Names(), name) {
		return nil, fmt.Errorf("%q is not a preset brand: give one of %s", name, strings.Join(Names(), ", "))
	}

	return fs.Sub(presets, name)
}
