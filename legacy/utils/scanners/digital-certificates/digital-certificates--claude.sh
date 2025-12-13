#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Digital Certificates Scanner - Claude Enhanced Version
# Wrapper that runs digital-certificates.sh with --claude flag enabled by default
#
# Usage: ./digital-certificates--claude.sh [options] <target>
#############################################################################

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Pass all arguments to main script with --claude prepended
exec "$SCRIPT_DIR/digital-certificates.sh" --claude "$@"
