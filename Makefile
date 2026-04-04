# SPDX-FileCopyrightText: Copyright 2025 The OSPS Authors
# SPDX-License-Identifier: Apache-2.0

.PHONY: release compile validate

# Usage: make release VERSION=2026-04-04
release:
ifndef VERSION
	$(error VERSION is required, e.g. make release VERSION=2026-04-04)
endif
	cd cmd && go run . release --source-dir ../baseline $(VERSION)

compile:
	cd cmd && go run . compile

validate:
	cd cmd && go run . validate
