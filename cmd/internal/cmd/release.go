// SPDX-FileCopyrightText: Copyright 2025 The OSPS Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ossf/security-baseline/pkg/baseline"
)

var versionFormat = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func addRelease(parentCmd *cobra.Command) {
	var baselinePath string

	releaseCmd := &cobra.Command{
		Use:   "release VERSION",
		Short: "Generate all release artifacts and update site documentation",
		Long: `release generates the versioned baseline and checklist documents for a new
release and updates the surrounding site files to make it the current version.

VERSION must be in YYYY-MM-DD format (e.g. 2026-04-04).

Files written:
  docs/versions/<VERSION>.md
  docs/versions/<VERSION>-checklist.md

Files updated:
  docs/versions/<PREVIOUS>.md  (nav-title: Current Version removed)
  docs/_config.yml              (nav_pages updated)
  docs/index.md                 (current/previous version links updated)

After running, review the diff, update docs/release_notes.md, then open a PR.`,
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			version := args[0]
			if !versionFormat.MatchString(version) {
				return fmt.Errorf("version must be YYYY-MM-DD (e.g. 2026-04-04), got %q", version)
			}

			absBaseline, err := filepath.Abs(baselinePath)
			if err != nil {
				return fmt.Errorf("resolving baseline path: %w", err)
			}
			repoRoot := filepath.Dir(absBaseline)
			docsVersionsDir := filepath.Join(repoRoot, "docs", "versions")

			if _, err := os.Stat(docsVersionsDir); err != nil {
				return fmt.Errorf("docs/versions not found at %s — is --source-dir pointing to the right place?", docsVersionsDir)
			}

			versionFile := filepath.Join(docsVersionsDir, version+".md")
			checklistFile := filepath.Join(docsVersionsDir, version+"-checklist.md")
			for _, f := range []string{versionFile, checklistFile} {
				if _, err := os.Stat(f); err == nil {
					return fmt.Errorf("output file already exists: %s", f)
				}
			}

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

			// Demote the previous current version before writing new files so the
			// scanner does not find and immediately strip the nav-title we are about to add.
			prevVersion, err := demotePreviousCurrent(docsVersionsDir)
			if err != nil {
				return fmt.Errorf("updating previous version: %w", err)
			}
			if prevVersion != "" {
				fmt.Printf("✅ Removed nav-title from %s.md\n", prevVersion)
			}

			gen := baseline.NewGenerator()

			if err := gen.ExportMarkdownRelease(bline, version, versionFile); err != nil {
				return fmt.Errorf("generating version document: %w", err)
			}
			fmt.Printf("✅ %s\n", versionFile)

			if err := gen.ExportChecklistRelease(bline, version, checklistFile); err != nil {
				return fmt.Errorf("generating checklist: %w", err)
			}
			fmt.Printf("✅ %s\n", checklistFile)

			configPath := filepath.Join(repoRoot, "docs", "_config.yml")
			if err := updateConfigNavPages(configPath, version); err != nil {
				return fmt.Errorf("updating _config.yml: %w", err)
			}
			fmt.Printf("✅ %s\n", configPath)

			indexPath := filepath.Join(repoRoot, "docs", "index.md")
			if err := updateIndexMd(indexPath, version); err != nil {
				return fmt.Errorf("updating index.md: %w", err)
			}
			fmt.Printf("✅ %s\n", indexPath)

			fmt.Printf("\nRelease %s ready. Review the diff, update docs/release_notes.md, then open a PR.\n", version)
			return nil
		},
	}

	releaseCmd.Flags().StringVarP(
		&baselinePath, "source-dir", "b", defaultBaselinePath,
		"path to directory containing the baseline YAML data",
	)
	parentCmd.AddCommand(releaseCmd)
}

// demotePreviousCurrent finds the docs/versions file marked as the current version
// (by its nav-title frontmatter), strips that frontmatter, and returns the version slug.
func demotePreviousCurrent(versionsDir string) (string, error) {
	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return "", err
	}

	const navTitleBlock = "---\nnav-title: Current Version\n---\n\n"

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".md") || strings.Contains(name, "-checklist") {
			continue
		}
		path := filepath.Join(versionsDir, name)
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		if bytes.HasPrefix(content, []byte(navTitleBlock)) {
			if err := os.WriteFile(path, content[len(navTitleBlock):], 0o644); err != nil {
				return "", err
			}
			return strings.TrimSuffix(name, ".md"), nil
		}
	}
	return "", nil
}

// updateConfigNavPages replaces the versioned entry in the Jekyll nav_pages list.
func updateConfigNavPages(configPath, newVersion string) error {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`([ \t]+-[ \t]+)versions/\d{4}-\d{2}-\d{2}\.md`)
	if !re.MatchString(string(content)) {
		return fmt.Errorf("no versioned entry found in nav_pages in %s", configPath)
	}
	updated := re.ReplaceAllString(string(content), "${1}versions/"+newVersion+".md")
	return os.WriteFile(configPath, []byte(updated), 0o644)
}

// updateIndexMd promotes newVersion to current and demotes the previous current to the
// previous-versions list.
func updateIndexMd(indexPath, version string) error {
	content, err := os.ReadFile(indexPath)
	if err != nil {
		return err
	}

	displayVer := strings.ReplaceAll(version, "-", ".")
	newCurrentLine := fmt.Sprintf(
		"* Current version: [v%s](versions/%s) (<a href=\"versions/%s-checklist.md\">checklist</a>) ([release notes](release_notes.md#%s))",
		displayVer, version, version, version,
	)

	lines := strings.Split(string(content), "\n")

	currentIdx := -1
	prevHeaderIdx := -1
	var oldCurrentRest string

	for i, line := range lines {
		if strings.HasPrefix(line, "* Current version: ") {
			currentIdx = i
			oldCurrentRest = strings.TrimPrefix(line, "* Current version: ")
		}
		if strings.HasPrefix(line, "* Previous versions:") {
			prevHeaderIdx = i
		}
	}
	if currentIdx == -1 {
		return fmt.Errorf("'* Current version:' not found in %s", indexPath)
	}
	if prevHeaderIdx == -1 {
		return fmt.Errorf("'* Previous versions:' not found in %s", indexPath)
	}

	lines[currentIdx] = newCurrentLine

	// Insert demoted line as first entry under "Previous versions:"
	result := make([]string, 0, len(lines)+1)
	for i, line := range lines {
		result = append(result, line)
		if i == prevHeaderIdx {
			result = append(result, "    * "+oldCurrentRest)
		}
	}

	return os.WriteFile(indexPath, []byte(strings.Join(result, "\n")), 0o644)
}
