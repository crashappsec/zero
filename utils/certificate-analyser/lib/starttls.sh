#!/bin/bash
# Copyright (c) 2025 Crash Override Inc.
# https://crashoverride.com
#
# SPDX-License-Identifier: GPL-3.0

#############################################################################
# StartTLS Protocol Library
# Support for StartTLS certificate retrieval across various protocols
# Compatible with bash 3.2+ (macOS default)
#############################################################################

# Cross-platform timeout command
# macOS doesn't have timeout by default, use perl as fallback
_run_with_timeout() {
    local timeout_secs="$1"
    shift

    # Try GNU timeout first
    if command -v timeout >/dev/null 2>&1; then
        timeout "$timeout_secs" "$@"
    # Try gtimeout (homebrew coreutils on macOS)
    elif command -v gtimeout >/dev/null 2>&1; then
        gtimeout "$timeout_secs" "$@"
    else
        # Perl fallback for macOS
        perl -e '
            use strict;
            use warnings;
            my $timeout = shift;
            my @cmd = @ARGV;
            $SIG{ALRM} = sub { exit 124; };
            alarm($timeout);
            exec(@cmd) or exit(127);
        ' "$timeout_secs" "$@"
    fi
}

# Get default port for StartTLS protocol
# Usage: get_starttls_port <protocol>
get_starttls_port() {
    local proto="$1"
    proto=$(echo "$proto" | tr '[:upper:]' '[:lower:]')

    case "$proto" in
        smtp) echo "25" ;;
        smtps) echo "587" ;;
        pop3) echo "110" ;;
        imap) echo "143" ;;
        ftp) echo "21" ;;
        ldap) echo "389" ;;
        mysql) echo "3306" ;;
        postgres|postgresql) echo "5432" ;;
        xmpp) echo "5222" ;;
        xmpp-server) echo "5269" ;;
        sieve) echo "4190" ;;
        nntp) echo "119" ;;
        lmtp) echo "24" ;;
        *) echo "0" ;;
    esac
}

# Check if protocol supports StartTLS via openssl
# Usage: is_openssl_starttls_supported <protocol>
is_openssl_starttls_supported() {
    local proto="$1"
    proto=$(echo "$proto" | tr '[:upper:]' '[:lower:]')

    local param
    param=$(get_openssl_starttls_param "$proto")
    [[ -n "$param" ]]
}

# Get openssl starttls parameter for protocol
# Usage: get_openssl_starttls_param <protocol>
get_openssl_starttls_param() {
    local proto="$1"
    proto=$(echo "$proto" | tr '[:upper:]' '[:lower:]')

    case "$proto" in
        smtp|smtps) echo "smtp" ;;
        pop3) echo "pop3" ;;
        imap) echo "imap" ;;
        ftp) echo "ftp" ;;
        ldap) echo "ldap" ;;
        xmpp) echo "xmpp" ;;
        xmpp-server) echo "xmpp-server" ;;
        sieve) echo "sieve" ;;
        nntp) echo "nntp" ;;
        lmtp) echo "lmtp" ;;
        postgres|postgresql) echo "postgres" ;;
        mysql) echo "mysql" ;;
        *) echo "" ;;
    esac
}

# Retrieve certificates via StartTLS
# Usage: starttls_get_certs <protocol> <host> [port] <output_file>
starttls_get_certs() {
    local proto="$1"
    local host="$2"
    local port="${3:-}"
    local output_file="${4:-}"

    proto=$(echo "$proto" | tr '[:upper:]' '[:lower:]')

    # Get default port if not specified
    if [[ -z "$port" ]] || [[ "$port" == "-" ]]; then
        port=$(get_starttls_port "$proto")
    fi

    # Set output file if not specified
    if [[ -z "$output_file" ]]; then
        output_file=$(mktemp)
    fi

    # Handle different protocols
    case "$proto" in
        mysql)
            starttls_mysql_get_certs "$host" "$port" "$output_file"
            ;;
        postgres|postgresql)
            starttls_postgres_get_certs "$host" "$port" "$output_file"
            ;;
        *)
            # Use openssl s_client for supported protocols
            local starttls_param=$(get_openssl_starttls_param "$proto")
            if [[ -n "$starttls_param" ]]; then
                starttls_openssl_get_certs "$host" "$port" "$starttls_param" "$output_file"
            else
                echo "Error: Unsupported protocol: $proto" >&2
                return 1
            fi
            ;;
    esac
}

# Get certificates using openssl s_client with -starttls
# Usage: starttls_openssl_get_certs <host> <port> <starttls_param> <output_file>
starttls_openssl_get_certs() {
    local host="$1"
    local port="$2"
    local starttls_param="$3"
    local output_file="$4"

    echo | _run_with_timeout 10 openssl s_client \
        -connect "$host:$port" \
        -starttls "$starttls_param" \
        -showcerts \
        2>/dev/null | \
        sed -n '/-----BEGIN CERTIFICATE-----/,/-----END CERTIFICATE-----/p' > "$output_file"

    if [[ -s "$output_file" ]]; then
        return 0
    else
        return 1
    fi
}

# Get certificates from MySQL via StartTLS
# Usage: starttls_mysql_get_certs <host> <port> <output_file>
starttls_mysql_get_certs() {
    local host="$1"
    local port="${2:-3306}"
    local output_file="$3"

    # MySQL uses a custom protocol for StartTLS
    # We need to send a specific handshake

    echo | _run_with_timeout 10 openssl s_client \
        -connect "$host:$port" \
        -starttls mysql \
        -showcerts \
        2>/dev/null | \
        sed -n '/-----BEGIN CERTIFICATE-----/,/-----END CERTIFICATE-----/p' > "$output_file"

    if [[ -s "$output_file" ]]; then
        return 0
    fi

    # Fallback: try direct SSL connection (for MySQL with native SSL)
    echo | _run_with_timeout 10 openssl s_client \
        -connect "$host:$port" \
        -showcerts \
        2>/dev/null | \
        sed -n '/-----BEGIN CERTIFICATE-----/,/-----END CERTIFICATE-----/p' > "$output_file"

    [[ -s "$output_file" ]]
}

# Get certificates from PostgreSQL via StartTLS
# Usage: starttls_postgres_get_certs <host> <port> <output_file>
starttls_postgres_get_certs() {
    local host="$1"
    local port="${2:-5432}"
    local output_file="$3"

    echo | _run_with_timeout 10 openssl s_client \
        -connect "$host:$port" \
        -starttls postgres \
        -showcerts \
        2>/dev/null | \
        sed -n '/-----BEGIN CERTIFICATE-----/,/-----END CERTIFICATE-----/p' > "$output_file"

    if [[ -s "$output_file" ]]; then
        return 0
    fi

    # Fallback: try direct SSL connection
    echo | _run_with_timeout 10 openssl s_client \
        -connect "$host:$port" \
        -showcerts \
        2>/dev/null | \
        sed -n '/-----BEGIN CERTIFICATE-----/,/-----END CERTIFICATE-----/p' > "$output_file"

    [[ -s "$output_file" ]]
}

# Test StartTLS connection
# Usage: test_starttls <protocol> <host> [port]
test_starttls() {
    local proto="$1"
    local host="$2"
    local port="${3:-}"

    proto=$(echo "$proto" | tr '[:upper:]' '[:lower:]')

    if [[ -z "$port" ]]; then
        port=$(get_starttls_port "$proto")
    fi

    local temp_file=$(mktemp)
    local result

    if starttls_get_certs "$proto" "$host" "$port" "$temp_file"; then
        local cert_count=$(grep -c "BEGIN CERTIFICATE" "$temp_file" 2>/dev/null || echo "0")
        result="SUCCESS: Retrieved $cert_count certificate(s)"
        rm -f "$temp_file"
        echo "$result"
        return 0
    else
        result="FAILED: Could not establish StartTLS connection"
        rm -f "$temp_file"
        echo "$result"
        return 1
    fi
}

# Get certificate info via StartTLS
# Usage: starttls_cert_info <protocol> <host> [port]
starttls_cert_info() {
    local proto="$1"
    local host="$2"
    local port="${3:-}"

    proto=$(echo "$proto" | tr '[:upper:]' '[:lower:]')

    if [[ -z "$port" ]]; then
        port=$(get_starttls_port "$proto")
    fi

    local starttls_param=$(get_openssl_starttls_param "$proto")

    if [[ -n "$starttls_param" ]]; then
        echo | _run_with_timeout 10 openssl s_client \
            -connect "$host:$port" \
            -starttls "$starttls_param" \
            2>/dev/null | openssl x509 -noout -text 2>/dev/null
    else
        # For MySQL/PostgreSQL, get certs first then parse
        local temp_file=$(mktemp)
        if starttls_get_certs "$proto" "$host" "$port" "$temp_file"; then
            openssl x509 -in "$temp_file" -noout -text 2>/dev/null
        fi
        rm -f "$temp_file"
    fi
}

# List supported StartTLS protocols
# Usage: list_starttls_protocols
list_starttls_protocols() {
    echo "Supported StartTLS Protocols:"
    echo "============================="
    echo ""
    printf "%-15s %s\n" "Protocol" "Default Port"
    printf "%-15s %s\n" "--------" "------------"

    for proto in smtp smtps pop3 imap ftp ldap mysql postgres xmpp sieve nntp lmtp; do
        local port=$(get_starttls_port "$proto")
        printf "%-15s %s\n" "$proto" "$port"
    done
}

# Check TLS version support via StartTLS
# Usage: check_starttls_tls_versions <protocol> <host> [port]
check_starttls_tls_versions() {
    local proto="$1"
    local host="$2"
    local port="${3:-}"

    proto=$(echo "$proto" | tr '[:upper:]' '[:lower:]')

    if [[ -z "$port" ]]; then
        port=$(get_starttls_port "$proto")
    fi

    local starttls_param=$(get_openssl_starttls_param "$proto")

    echo "TLS Version Support for $host:$port ($proto)"
    echo "==========================================="
    echo ""

    for tls_version in tls1 tls1_1 tls1_2 tls1_3; do
        local version_flag="-$tls_version"
        local version_name

        case "$tls_version" in
            tls1) version_name="TLS 1.0" ;;
            tls1_1) version_name="TLS 1.1" ;;
            tls1_2) version_name="TLS 1.2" ;;
            tls1_3) version_name="TLS 1.3" ;;
        esac

        local result
        if [[ -n "$starttls_param" ]]; then
            result=$(echo | _run_with_timeout 5 openssl s_client \
                -connect "$host:$port" \
                -starttls "$starttls_param" \
                $version_flag \
                2>&1)
        else
            result=$(echo | _run_with_timeout 5 openssl s_client \
                -connect "$host:$port" \
                $version_flag \
                2>&1)
        fi

        if echo "$result" | grep -q "Cipher is"; then
            printf "%-10s: ✓ Supported\n" "$version_name"
        else
            printf "%-10s: ✗ Not supported\n" "$version_name"
        fi
    done
}

# Get cipher suites via StartTLS
# Usage: get_starttls_ciphers <protocol> <host> [port]
get_starttls_ciphers() {
    local proto="$1"
    local host="$2"
    local port="${3:-}"

    proto=$(echo "$proto" | tr '[:upper:]' '[:lower:]')

    if [[ -z "$port" ]]; then
        port=$(get_starttls_port "$proto")
    fi

    local starttls_param=$(get_openssl_starttls_param "$proto")

    local result
    if [[ -n "$starttls_param" ]]; then
        result=$(echo | _run_with_timeout 10 openssl s_client \
            -connect "$host:$port" \
            -starttls "$starttls_param" \
            2>/dev/null)
    else
        local temp_file=$(mktemp)
        starttls_get_certs "$proto" "$host" "$port" "$temp_file"
        rm -f "$temp_file"
        result=$(echo | _run_with_timeout 10 openssl s_client -connect "$host:$port" 2>/dev/null)
    fi

    echo "Connection Details for $host:$port ($proto)"
    echo "============================================"
    echo ""
    echo "$result" | grep -E "^(Protocol|Cipher|Server public key|Secure Renegotiation)" | head -10
}

# Export functions
export -f get_starttls_port
export -f is_openssl_starttls_supported
export -f get_openssl_starttls_param
export -f starttls_get_certs
export -f starttls_openssl_get_certs
export -f starttls_mysql_get_certs
export -f starttls_postgres_get_certs
export -f test_starttls
export -f starttls_cert_info
export -f list_starttls_protocols
export -f check_starttls_tls_versions
export -f get_starttls_ciphers
