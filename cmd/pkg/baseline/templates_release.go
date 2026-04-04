// SPDX-FileCopyrightText: Copyright 2025 The OSPS Authors
// SPDX-License-Identifier: Apache-2.0

package baseline

import _ "embed"

// releaseChecklistTemplateContent is the embedded checklist template used by
// ExportChecklistRelease. The SDK has no ControlCatalog checklist renderer,
// so this stays template-based.
//
//go:embed template-checklist.md
var releaseChecklistTemplateContent string
