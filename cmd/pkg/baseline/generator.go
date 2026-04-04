// SPDX-FileCopyrightText: Copyright 2025 The OSPS Authors
// SPDX-License-Identifier: Apache-2.0

package baseline

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/gemaraproj/go-gemara/gemaraconv"

	"github.com/ossf/security-baseline/pkg/types"
)

func NewGenerator() *Generator {
	return &Generator{}
}

type Generator struct{}

// ExportMarkdownDevel renders the baseline using the embedded devel template.
func (g *Generator) ExportMarkdownDevel(b *types.Baseline, path string) error {
	return g.renderMarkdown(b, "", develTemplateContent, path)
}

// ExportChecklistDevel renders the baseline checklist using the embedded template.
func (g *Generator) ExportChecklistDevel(b *types.Baseline, path string) error {
	return g.renderMarkdown(b, "", releaseChecklistTemplateContent, path)
}

// ExportMarkdownRelease generates the versioned baseline document for a release using the
// go-gemara SDK renderer. version is the YYYY-MM-DD release identifier (e.g. "2026-04-04");
// the display form (dots) is set on the catalog copy. The output is prefixed with the Jekyll
// nav-title frontmatter required for site navigation.
func (g *Generator) ExportMarkdownRelease(b *types.Baseline, version, path string) error {
	// Shallow-copy the catalog so we can override metadata fields without mutating the caller.
	// Metadata is a value type inside ControlCatalog, so this copy is independent.
	bRelease := &types.Baseline{Lexicon: b.Lexicon}
	bRelease.Catalog = b.Catalog
	bRelease.Catalog.Metadata.Version = strings.ReplaceAll(version, "-", ".")
	bRelease.Catalog.Metadata.Draft = false

	var buf bytes.Buffer
	buf.WriteString("---\nnav-title: Current Version\n---\n\n")

	if err := g.ExportGemaraMarkdown(bRelease, &buf, gemaraconv.WithApplicabilityMatrix(true)); err != nil {
		return fmt.Errorf("generating release markdown: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), os.FileMode(0o750)); err != nil {
		return fmt.Errorf("error creating output directory %s: %w", filepath.Dir(path), err)
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}

// ExportChecklistRelease generates the versioned checklist for a release using an embedded
// template. version is the YYYY-MM-DD release identifier.
func (g *Generator) ExportChecklistRelease(b *types.Baseline, version, path string) error {
	displayVersion := strings.ReplaceAll(version, "-", ".")
	return g.renderMarkdown(b, displayVersion, releaseChecklistTemplateContent, path)
}

// renderMarkdown executes a template string against b and writes the result to path.
// version is exposed to templates via the {{ version }} function; pass "" for dev builds.
func (g *Generator) renderMarkdown(b *types.Baseline, version, templateContent, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), os.FileMode(0o750)); err != nil {
		return fmt.Errorf("error creating output directory %s: %w", filepath.Dir(path), err)
	}

	outputFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error creating output file %s: %w", path, err)
	}
	defer outputFile.Close() //nolint:errcheck

	tmpl, err := template.New("baseline").Funcs(template.FuncMap{
		"collapseNewlines": collapseNewlines,
		"addLinks": func(s string) string {
			return addLinksTemplateFunction(b.Lexicon, s)
		},
		"asLink":           asLinkTemplateFunction,
		"maxLevel":         maxLevel,
		"controlsForGroup": controlsForGroup,
		"isRetired":        isRetired,
		"applicabilityTitle": func(id string) string {
			return applicabilityTitle(b.Catalog.Metadata.ApplicabilityGroups, id)
		},
		"version": func() string { return version },
	}).Parse(templateContent)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	if err := tmpl.Execute(outputFile, b); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return nil
}
