// SPDX-License-Identifier: Apache-2.0

package baseline

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExportGemaraMarkdown_includesBaselineLexicon(t *testing.T) {
	baselineDir := filepath.Join("..", "..", "..", "baseline")
	l := Loader{DataPath: baselineDir}
	b, err := l.Load()
	require.NoError(t, err)
	require.NotEmpty(t, b.Lexicon, "baseline/lexicon.yaml should load")

	var buf bytes.Buffer
	g := NewGenerator()
	err = g.ExportGemaraMarkdown(b, &buf)
	require.NoError(t, err)
	out := buf.String()
	assert.Contains(t, out, "## Lexicon")
	assert.Contains(t, strings.Split(out, "## Lexicon")[1], "### Administrator")
}
