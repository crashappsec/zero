#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Package Health Scanner - Claude Enhanced Version
# Wrapper that runs package-health.sh with --claude flag enabled by default
#
# Usage: ./package-health--claude.sh [options] <target>
#############################################################################

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Pass all arguments to main script with --claude prepended
exec "$SCRIPT_DIR/package-health.sh" --claude "$@"
