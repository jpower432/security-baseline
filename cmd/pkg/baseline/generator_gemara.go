// SPDX-License-Identifier: Apache-2.0

package baseline

import (
	"fmt"
	"io"

	"github.com/gemaraproj/go-gemara/gemaraconv"
	"github.com/goccy/go-yaml"

	"github.com/ossf/security-baseline/pkg/types"
)

// ExportGemara writes the assembled ControlCatalog as Gemara-compliant YAML.
func (g *Generator) ExportGemara(b *types.Baseline, w io.Writer) error {
	out, err := yaml.Marshal(&b.Catalog)
	if err != nil {
		return fmt.Errorf("marshaling catalog: %w", err)
	}
	if _, err := w.Write(out); err != nil {
		return fmt.Errorf("writing YAML: %w", err)
	}
	return nil
}

// ExportGemaraMarkdown renders the ControlCatalog as Markdown using go-gemara's embedded templates.
// When the baseline includes lexicon.yaml entries, they are passed for term autolinking and a trailing
// ## Lexicon section (unless WithLexiconAutolink is used with metadata.lexicon, which uses a remote Gemara Lexicon instead).
func (g *Generator) ExportGemaraMarkdown(b *types.Baseline, w io.Writer, opts ...gemaraconv.MarkdownOption) error {
	mdOpts := make([]gemaraconv.MarkdownOption, 0, len(opts)+1)
	if len(b.Lexicon) > 0 {
		inline := make([]gemaraconv.InlineLexiconTerm, len(b.Lexicon))
		for i, e := range b.Lexicon {
			inline[i] = gemaraconv.InlineLexiconTerm{
				Term:       e.Term,
				Definition: e.Definition,
				Synonyms:   e.Synonyms,
				References: e.References,
			}
		}
		mdOpts = append(mdOpts, gemaraconv.WithInlineLexicon(inline))
	}
	mdOpts = append(mdOpts, opts...)

	out, err := gemaraconv.CatalogToMarkdown(&b.Catalog, mdOpts...)
	if err != nil {
		return fmt.Errorf("catalog to markdown: %w", err)
	}
	if _, err := w.Write(out); err != nil {
		return fmt.Errorf("writing markdown: %w", err)
	}
	return nil
}
