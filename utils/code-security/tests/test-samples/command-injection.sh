#!/bin/bash
# Test sample: Command injection vulnerabilities
# This file contains intentionally vulnerable code for testing

# VULNERABLE: User input directly in command
vulnerable_ping() {
    local host="$1"
    # Vulnerable - user input passed to shell
    ping -c 4 $host
}

# VULNERABLE: Using eval with user input
vulnerable_eval() {
    local command="$1"
    # Vulnerable - arbitrary command execution
    eval "$command"
}

# VULNERABLE: Backtick command substitution
vulnerable_backtick() {
    local filename="$1"
    # Vulnerable - command injection possible
    content=`cat $filename`
    echo "$content"
}

# VULNERABLE: User input in subshell
vulnerable_subshell() {
    local dir="$1"
    # Vulnerable - path traversal and command injection
    files=$(ls -la $dir)
    echo "$files"
}

# SECURE versions for comparison

# SECURE: Validate input before use
secure_ping() {
    local host="$1"
    # Validate hostname format
    if [[ ! "$host" =~ ^[a-zA-Z0-9.-]+$ ]]; then
        echo "Invalid hostname" >&2
        return 1
    fi
    ping -c 4 "$host"
}

# SECURE: Use arrays instead of string interpolation
secure_list() {
    local dir="$1"
    # Quote variables, use -- to prevent option injection
    ls -la -- "$dir"
}

# SECURE: Avoid eval entirely
secure_no_eval() {
    local action="$1"
    case "$action" in
        list) ls -la ;;
        status) git status ;;
        *) echo "Unknown action" >&2; return 1 ;;
    esac
}
