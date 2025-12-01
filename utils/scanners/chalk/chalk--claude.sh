#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Chalk Build Analyser - Claude Enhanced Version
# Wrapper that runs chalk.sh with --claude flag enabled by default
#
# Usage: ./chalk--claude.sh [options] <chalk-report.json>
#############################################################################

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Pass all arguments to main script with --claude prepended
exec "$SCRIPT_DIR/chalk.sh" --claude "$@"
