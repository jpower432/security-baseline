// SPDX-FileCopyrightText: Copyright 2025 The OSPS Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ossf/security-baseline/pkg/baseline"
)

func addCompile(parentCmd *cobra.Command) {
	var baselinePath string

	compileCmd := &cobra.Command{
		Use:   "compile",
		Short: "Compile the baseline YAML to all output formats",
		Long: `compile loads, validates, and renders the baseline to all output formats:

  docs/versions/devel.md
  docs/versions/devel-checklist.md
  build/baseline.gemara.yaml
  build/baseline.oscal.json`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(_ *cobra.Command, _ []string) error {
			absBaseline, err := filepath.Abs(baselinePath)
			if err != nil {
				return fmt.Errorf("resolving baseline path: %w", err)
			}
			repoRoot := filepath.Dir(absBaseline)

			loader := baseline.NewLoader()
			loader.DataPath = baselinePath

			bline, err := loader.Load()
			if err != nil {
				return err
			}

			validator := baseline.NewValidator()
			if err := validator.Check(bline); err != nil {
				fmt.Fprint(os.Stderr, "❌ Baseline validation failed:\n")
				return err
			}

			gen := baseline.NewGenerator()

			develMD := filepath.Join(repoRoot, "docs", "versions", "devel.md")
			if err := gen.ExportMarkdownDevel(bline, develMD); err != nil {
				return fmt.Errorf("generating devel.md: %w", err)
			}
			fmt.Printf("✅ %s\n", develMD)

			develChecklist := filepath.Join(repoRoot, "docs", "versions", "devel-checklist.md")
			if err := gen.ExportChecklistDevel(bline, develChecklist); err != nil {
				return fmt.Errorf("generating devel-checklist.md: %w", err)
			}
			fmt.Printf("✅ %s\n", develChecklist)

			buildDir := filepath.Join(repoRoot, "build")
			if err := os.MkdirAll(buildDir, 0o750); err != nil {
				return fmt.Errorf("creating build dir: %w", err)
			}

			gemaraOut := filepath.Join(buildDir, "baseline.gemara.yaml")
			gf, err := os.Create(gemaraOut)
			if err != nil {
				return fmt.Errorf("creating %s: %w", gemaraOut, err)
			}
			defer gf.Close() //nolint:errcheck
			if err := gen.ExportGemara(bline, gf); err != nil {
				return fmt.Errorf("generating gemara YAML: %w", err)
			}
			fmt.Printf("✅ %s\n", gemaraOut)

			oscalOut := filepath.Join(buildDir, "baseline.oscal.json")
			of, err := os.Create(oscalOut)
			if err != nil {
				return fmt.Errorf("creating %s: %w", oscalOut, err)
			}
			defer of.Close() //nolint:errcheck
			if err := gen.ExportOSCAL(bline, of); err != nil {
				return fmt.Errorf("generating OSCAL JSON: %w", err)
			}
			fmt.Printf("✅ %s\n", oscalOut)

			return nil
		},
	}

	compileCmd.Flags().StringVarP(
		&baselinePath, "source-dir", "b", defaultBaselinePath,
		"path to directory containing the baseline YAML data",
	)
	parentCmd.AddCommand(compileCmd)
}
