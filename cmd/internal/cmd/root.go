// SPDX-FileCopyrightText: Copyright 2025 The OSPS Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import "github.com/spf13/cobra"

const (
	// defaultBaselinePath is the default location where the CLI will look for
	// the baseline YAML files
	defaultBaselinePath = "../baseline"
)

func Execute() error {
	rootCmd := &cobra.Command{
		Use:  "baseline-compiler",
		Long: `Baseline Compiler reads the baseline YAML and renders it to all output formats.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	addRelease(rootCmd)
	addCompile(rootCmd)
	addValidate(rootCmd)

	return rootCmd.Execute()
}
