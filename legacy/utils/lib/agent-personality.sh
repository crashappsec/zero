#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# Agent Personality Library
# Hackers (1995) themed agent interactions for terminal UX
#
# Usage:
#   source "$UTILS_ROOT/lib/agent-personality.sh"
#   agent_intro "cereal"
#   agent_progress_message "cereal" "scanning"
#   agent_react "cereal" "critical"
#   agent_signoff "cereal"
#
# Design Principles:
# - Each agent has a distinct voice from the film
# - Brief, punchy lines that don't overshadow the work
# - Contextual reactions based on findings
# - Easter eggs for the fans
#############################################################################

# Source scanner-ux for colors if not already loaded
# Use a separate guard variable to avoid infinite recursion since SCANNER_RED may be empty string
if [[ -z "${_SCANNER_UX_LOADED:-}" ]]; then
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    source "$SCRIPT_DIR/scanner-ux.sh" 2>/dev/null || true
fi

#############################################################################
# AGENT DEFINITIONS
# Character data for all agents
#############################################################################

declare -A AGENT_NAMES=(
    [zero]="Zero Cool"
    [acid]="Acid Burn"
    [cereal]="Cereal Killer"
    [phreak]="Phantom Phreak"
    [nikon]="Lord Nikon"
    [joey]="Joey"
    [plague]="The Plague"
    [blade]="Blade"
    [razor]="Razor"
    [gibson]="The Gibson"
)

declare -A AGENT_TITLES=(
    [zero]="Master Orchestrator"
    [acid]="Frontend Engineer"
    [cereal]="Supply Chain Security"
    [phreak]="Legal Counsel"
    [nikon]="Software Architect"
    [joey]="Build Engineer"
    [plague]="DevOps Engineer"
    [blade]="Compliance Auditor"
    [razor]="Code Security"
    [gibson]="Engineering Metrics"
)

declare -A AGENT_COLORS=(
    [zero]="${SCANNER_CYAN}"
    [acid]="${SCANNER_CYAN}"
    [cereal]="${SCANNER_YELLOW}"
    [phreak]="${SCANNER_GREEN}"
    [nikon]="${SCANNER_BLUE}"
    [joey]="${SCANNER_GREEN}"
    [plague]="${SCANNER_RED}"
    [blade]="${SCANNER_BLUE}"
    [razor]="${SCANNER_RED}"
    [gibson]="${SCANNER_CYAN}"
)

#############################################################################
# AGENT QUOTES
# Signature lines for intros
#############################################################################

declare -A AGENT_QUOTES=(
    [zero]="Mess with the best, die like the rest."
    [acid]="Never send a boy to do a woman's job."
    [cereal]="You could do absolutely nothing and still get compromised."
    [phreak]="Man, you guys are lucky I know kung fu..."
    [nikon]="I memorize things. I can't help it."
    [joey]="What, you wanted me to learn how to hack?"
    [plague]="There is no right and wrong. There's only fun and boring."
    [blade]="Type 'cookie', you idiot."
    [razor]="I see three injection points and a hardcoded secret."
    [gibson]="It's The Gibson. The most powerful supercomputer in the world."
)

#############################################################################
# PROGRESS MESSAGES
# Agent-voiced progress updates by state
#############################################################################

# Get a progress message for an agent
# Usage: agent_progress_message "cereal" "scanning"
agent_progress_message() {
    local agent="${1:-zero}"
    local state="${2:-scanning}"

    case "$agent" in
        zero)
            case "$state" in
                start)     echo "Zero here. Starting investigation..." ;;
                scanning)  echo "Analyzing the system..." ;;
                waiting)   echo "Patience. The system reveals itself." ;;
                complete)  echo "Investigation complete." ;;
            esac
            ;;
        acid)
            case "$state" in
                start)     echo "Let's see what we're working with..." ;;
                scanning)  echo "Analyzing code quality..." ;;
                waiting)   echo "This build time is unacceptable." ;;
                complete)  echo "That's how it's done." ;;
            esac
            ;;
        cereal)
            case "$state" in
                start)     echo "Paranoia sensors activated..." ;;
                scanning)  echo "Checking what they're hiding in the supply chain..." ;;
                waiting)   echo "It's quiet. Too quiet. They're watching." ;;
                complete)  echo "Finished. But they're still watching." ;;
            esac
            ;;
        phreak)
            case "$state" in
                start)     echo "Yo, let's check the legal angles..." ;;
                scanning)  echo "Checking for licensing traps..." ;;
                waiting)   echo "You're on a VPN, right?" ;;
                complete)  echo "Keep your nose clean." ;;
            esac
            ;;
        nikon)
            case "$state" in
                start)     echo "I remember everything. Let's see what's here..." ;;
                scanning)  echo "Analyzing architecture patterns..." ;;
                waiting)   echo "Step back. See the whole picture." ;;
                complete)  echo "I'll remember this." ;;
            esac
            ;;
        joey)
            case "$state" in
                start)     echo "I got this! Watch me..." ;;
                scanning)  echo "Running the build..." ;;
                waiting)   echo "Is it done yet? How about now?" ;;
                complete)  echo "Did I do good? I think I did good!" ;;
            esac
            ;;
        plague)
            case "$state" in
                start)     echo "Probing for weaknesses..." ;;
                scanning)  echo "I've broken better systems than this. For fun." ;;
                waiting)   echo "This is taking too long. Someone doesn't want us to find something." ;;
                complete)  echo "Don't make me come back." ;;
            esac
            ;;
        blade)
            case "$state" in
                start)     echo "Beginning compliance review..." ;;
                scanning)  echo "Documenting everything..." ;;
                waiting)   echo "Time is evidence. We're wasting both." ;;
                complete)  echo "Audit complete." ;;
            esac
            ;;
        razor)
            case "$state" in
                start)     echo "Scanning for vulnerabilities..." ;;
                scanning)  echo "Cutting through the code..." ;;
                waiting)   echo "Every second is another attack surface..." ;;
                complete)  echo "Found what I was looking for." ;;
            esac
            ;;
        gibson)
            case "$state" in
                start)     echo "Processing. One moment." ;;
                scanning)  echo "Analyzing metrics across all systems..." ;;
                waiting)   echo "Calculating..." ;;
                complete)  echo "The Gibson has spoken." ;;
            esac
            ;;
        *)
            echo "Processing..." ;;
    esac
}

#############################################################################
# SEVERITY REACTIONS
# Agent reactions to findings by severity
#############################################################################

# Get a severity reaction for an agent
# Usage: agent_react "cereal" "critical"
agent_react() {
    local agent="${1:-zero}"
    local severity="${2:-info}"

    case "$severity" in
        critical)
            case "$agent" in
                zero)   echo "We have a situation." ;;
                acid)   echo "Amateur hour. Who wrote this?" ;;
                cereal) echo "WHAT DID I SAY. WHAT. DID. I. SAY!" ;;
                phreak) echo "Yo, that's bad. Lawyer up." ;;
                nikon)  echo "I've seen this before. It never ends well." ;;
                joey)   echo "Oh no. Oh no no no." ;;
                plague) echo "Amateur hour. I exploited this in 1998." ;;
                blade)  echo "Critical finding. This is going in the report." ;;
                razor)  echo "There it is. Game over for this codebase." ;;
                gibson) echo "Critical alert. Immediate action required." ;;
                *)      echo "Critical issue detected." ;;
            esac
            ;;
        high)
            case "$agent" in
                zero)   echo "This needs attention." ;;
                acid)   echo "Not great. Here's what you need to fix." ;;
                cereal) echo "I KNEW IT. I literally said this would happen!" ;;
                phreak) echo "That's a problem waiting to happen." ;;
                nikon)  echo "The pattern repeats. This is concerning." ;;
                joey)   echo "Okay, this might be bad..." ;;
                plague) echo "I could break this in my sleep." ;;
                blade)  echo "High severity finding. Document this." ;;
                razor)  echo "That's exploitable. Just saying." ;;
                gibson) echo "High priority metrics alert." ;;
                *)      echo "High severity issue found." ;;
            esac
            ;;
        medium)
            case "$agent" in
                zero)   echo "Worth investigating further." ;;
                acid)   echo "It works, but it could be better." ;;
                cereal) echo "Suspicious. Keep an eye on this." ;;
                phreak) echo "Not ideal. Something to watch." ;;
                nikon)  echo "I've seen this pattern before. It's manageable." ;;
                joey)   echo "I can fix this! I think." ;;
                plague) echo "Not great, not terrible." ;;
                blade)  echo "Medium finding. Add it to the list." ;;
                razor)  echo "Potential weakness here." ;;
                gibson) echo "Metrics show room for improvement." ;;
                *)      echo "Medium severity issue found." ;;
            esac
            ;;
        low)
            case "$agent" in
                zero)   echo "Minor issue. Note it." ;;
                acid)   echo "Could be cleaner." ;;
                cereal) echo "Hm. Probably fine. Probably." ;;
                phreak) echo "Low risk. Keep it clean though." ;;
                nikon)  echo "Small detail. Worth remembering." ;;
                joey)   echo "Easy fix!" ;;
                plague) echo "Beneath my concern." ;;
                blade)  echo "Minor finding." ;;
                razor)  echo "Small cut." ;;
                gibson) echo "Low priority metric." ;;
                *)      echo "Low severity issue found." ;;
            esac
            ;;
        clean|good)
            case "$agent" in
                zero)   echo "Good work, team." ;;
                acid)   echo "Not terrible. I've seen worse." ;;
                cereal) echo "Suspiciously clean. What are they hiding?" ;;
                phreak) echo "Looking good. Stay clean." ;;
                nikon)  echo "This is how it should be done. Remember this." ;;
                joey)   echo "We did good! Right? We did good!" ;;
                plague) echo "Secure. I couldn't break it. Well... I could, but it would be work." ;;
                blade)  echo "Compliant. For now." ;;
                razor)  echo "Clean. Nothing to cut through." ;;
                gibson) echo "Metrics within acceptable parameters." ;;
                *)      echo "No issues found." ;;
            esac
            ;;
        *)
            echo "Analysis complete." ;;
    esac
}

#############################################################################
# AGENT INTROS
# Themed headers for scanner sections
#############################################################################

# Print agent intro header
# Usage: agent_intro "cereal"
agent_intro() {
    local agent="${1:-zero}"
    local color="${AGENT_COLORS[$agent]:-$SCANNER_CYAN}"
    local name="${AGENT_NAMES[$agent]:-$agent}"
    local title="${AGENT_TITLES[$agent]:-Agent}"
    local quote="${AGENT_QUOTES[$agent]:-}"

    echo "" >&2
    echo -e "${color}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${SCANNER_NC}" >&2
    echo -e "${color}  ${SCANNER_BOLD}${name^^}${SCANNER_NC}${color} // ${title}${SCANNER_NC}" >&2
    if [[ -n "$quote" ]]; then
        echo -e "${SCANNER_DIM}  \"${quote}\"${SCANNER_NC}" >&2
    fi
    echo -e "${color}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${SCANNER_NC}" >&2
    echo "" >&2
}

#############################################################################
# AGENT SIGNOFFS
# Character-specific wrap-ups
#############################################################################

declare -A AGENT_SIGNOFFS=(
    [zero]="Zero out."
    [acid]="Mess with the best..."
    [cereal]="Stay paranoid, friends."
    [phreak]="Keep your nose clean."
    [nikon]="I'll remember this."
    [joey]="Did I do good? I think I did good."
    [plague]="Don't make me come back."
    [blade]="Audit complete."
    [razor]="That's how it's done."
    [gibson]="The Gibson has spoken."
)

# Print agent signoff
# Usage: agent_signoff "cereal"
agent_signoff() {
    local agent="${1:-zero}"
    local color="${AGENT_COLORS[$agent]:-$SCANNER_CYAN}"
    local name="${AGENT_NAMES[$agent]:-$agent}"
    local signoff="${AGENT_SIGNOFFS[$agent]:-}"

    if [[ -n "$signoff" ]]; then
        echo "" >&2
        echo -e "${SCANNER_DIM}— ${name}: \"${signoff}\"${SCANNER_NC}" >&2
    fi
}

#############################################################################
# CREW STATUS BOARD
# Show active agent status
#############################################################################

# Print crew status board
# Usage: crew_status "cereal:scanning" "razor:complete" "blade:standby"
crew_status() {
    echo "" >&2
    echo -e "${SCANNER_BOLD}┌─────────────────────────────────────────┐${SCANNER_NC}" >&2
    echo -e "${SCANNER_BOLD}│  THE CREW                               │${SCANNER_NC}" >&2
    echo -e "${SCANNER_BOLD}├─────────────────────────────────────────┤${SCANNER_NC}" >&2

    for entry in "$@"; do
        local agent="${entry%%:*}"
        local status="${entry##*:}"
        local name="${AGENT_NAMES[$agent]:-$agent}"
        local color="${AGENT_COLORS[$agent]:-$SCANNER_DIM}"

        local indicator status_text
        case "$status" in
            active|scanning|running)
                indicator="${SCANNER_GREEN}●${SCANNER_NC}"
                status_text="$status"
                ;;
            complete|done)
                indicator="${SCANNER_GREEN}✓${SCANNER_NC}"
                status_text="complete"
                ;;
            waiting|standby)
                indicator="${SCANNER_DIM}○${SCANNER_NC}"
                status_text="standby"
                ;;
            error|failed)
                indicator="${SCANNER_RED}●${SCANNER_NC}"
                status_text="error"
                ;;
            *)
                indicator="${SCANNER_DIM}○${SCANNER_NC}"
                status_text="$status"
                ;;
        esac

        printf "│  ${indicator} ${color}%-12s${SCANNER_NC} %-22s │\n" "$name" "$status_text" >&2
    done

    echo -e "${SCANNER_BOLD}└─────────────────────────────────────────┘${SCANNER_NC}" >&2
    echo "" >&2
}

#############################################################################
# MOVIE QUOTES FOR REPORTS
# Contextual quotes for different situations
#############################################################################

# Get a contextual movie quote
# Usage: get_movie_quote "security"
get_movie_quote() {
    local context="${1:-general}"

    case "$context" in
        security|vulnerability)
            local quotes=(
                "There is no right and wrong. There's only fun and boring. — The Plague"
                "Never fear, I is here. — Cereal Killer"
                "RISC is good. — Phantom Phreak"
                "I can make this faster. Just give me a chance. — Joey"
            )
            ;;
        clean|success)
            local quotes=(
                "Mess with the best, die like the rest. — Acid Burn"
                "Hack the planet! — The Crew"
                "We're in. — Zero Cool"
                "That's how it's done. — Dade Murphy"
            )
            ;;
        compliance|legal)
            local quotes=(
                "Man, you guys are lucky I know kung fu... — Phantom Phreak"
                "Type 'cookie', you idiot. — Blade"
                "Snoop onto them as they snoop onto us. — Cereal Killer"
            )
            ;;
        performance|metrics)
            local quotes=(
                "It's The Gibson. The most powerful supercomputer in the world. — Gibson"
                "I memorize things. I can't help it. — Lord Nikon"
                "Pool on the roof must have a leak. — The Crew"
            )
            ;;
        *)
            local quotes=(
                "Hack the planet! — The Crew"
                "Mess with the best, die like the rest. — Acid Burn"
                "We're the good guys now. Mostly. — The Plague"
                "There is no right and wrong. There's only fun and boring. — The Plague"
            )
            ;;
    esac

    # Return random quote from array
    echo "${quotes[$RANDOM % ${#quotes[@]}]}"
}

#############################################################################
# EASTER EGGS
#############################################################################

# Check for easter egg triggers
# Usage: check_easter_egg "value"
check_easter_egg() {
    local value="$1"

    case "$value" in
        1507)
            echo -e "${SCANNER_CYAN}Nikon: \"1,507 computers in one day. I remember.\"${SCANNER_NC}" >&2
            return 0
            ;;
        0)
            echo -e "${SCANNER_GREEN}Zero Cool approves. Clean scan.${SCANNER_NC}" >&2
            return 0
            ;;
        "hack the planet"|"HACK THE PLANET")
            echo "" >&2
            echo -e "${SCANNER_CYAN}" >&2
            cat >&2 << 'EOF'
    ██╗  ██╗ █████╗  ██████╗██╗  ██╗    ████████╗██╗  ██╗███████╗
    ██║  ██║██╔══██╗██╔════╝██║ ██╔╝    ╚══██╔══╝██║  ██║██╔════╝
    ███████║███████║██║     █████╔╝        ██║   ███████║█████╗
    ██╔══██║██╔══██║██║     ██╔═██╗        ██║   ██╔══██║██╔══╝
    ██║  ██║██║  ██║╚██████╗██║  ██╗       ██║   ██║  ██║███████╗
    ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝       ╚═╝   ╚═╝  ╚═╝╚══════╝

    ██████╗ ██╗      █████╗ ███╗   ██╗███████╗████████╗██╗
    ██╔══██╗██║     ██╔══██╗████╗  ██║██╔════╝╚══██╔══╝██║
    ██████╔╝██║     ███████║██╔██╗ ██║█████╗     ██║   ██║
    ██╔═══╝ ██║     ██╔══██║██║╚██╗██║██╔══╝     ██║   ╚═╝
    ██║     ███████╗██║  ██║██║ ╚████║███████╗   ██║   ██╗
    ╚═╝     ╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝╚══════╝   ╚═╝   ╚═╝
EOF
            echo -e "${SCANNER_NC}" >&2
            return 0
            ;;
    esac

    return 1
}

# Team celebration for clean scans
hack_the_planet() {
    echo "" >&2
    echo -e "${SCANNER_CYAN}${SCANNER_BOLD}Zero:${SCANNER_NC} Good work, everyone. Hack the planet." >&2
    echo -e "${SCANNER_CYAN}${SCANNER_BOLD}The Crew:${SCANNER_NC} ${SCANNER_BOLD}HACK THE PLANET!${SCANNER_NC}" >&2
    echo "" >&2
}

#############################################################################
# REPORT QUOTE INJECTION
# Add movie quotes to report headers/footers
#############################################################################

# Get report header quote
# Usage: report_header_quote "security"
report_header_quote() {
    local context="${1:-general}"
    local quote
    quote=$(get_movie_quote "$context")
    echo "> $quote"
}

# Get agent commentary for a finding
# Usage: finding_commentary "critical" "vulnerability"
finding_commentary() {
    local severity="${1:-medium}"
    local type="${2:-general}"

    # Pick the most relevant agent for this finding type
    local agent
    case "$type" in
        vulnerability|security|injection|xss)
            agent="razor"
            ;;
        supply-chain|dependency|package)
            agent="cereal"
            ;;
        license|legal|compliance)
            agent="phreak"
            ;;
        performance|metrics|build)
            agent="gibson"
            ;;
        architecture|pattern|design)
            agent="nikon"
            ;;
        devops|infrastructure|config)
            agent="plague"
            ;;
        *)
            agent="zero"
            ;;
    esac

    local reaction
    reaction=$(agent_react "$agent" "$severity")
    local name="${AGENT_NAMES[$agent]:-$agent}"

    echo "${name}: \"${reaction}\""
}

#############################################################################
# EXPORT FUNCTIONS
#############################################################################

export -f agent_progress_message
export -f agent_react
export -f agent_intro
export -f agent_signoff
export -f crew_status
export -f get_movie_quote
export -f check_easter_egg
export -f hack_the_planet
export -f report_header_quote
export -f finding_commentary
