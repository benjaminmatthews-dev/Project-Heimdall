#!/usr/bin/env bash

set -euo pipefail

if [[ ${1:-} == "--" ]]; then
    shift
fi

required_vars=(BIFROST_HOST_ADDR BIFROST_HOST_USER BIFROST_HOST_KEY_PATH)
missing_vars=()
for var in "${required_vars[@]}"; do
    if [[ -z "${!var:-}" ]]; then
        missing_vars+=("$var")
    fi
done

if [[ ${#missing_vars[@]} -gt 0 ]]; then
    echo "ERROR: Missing required env vars: ${missing_vars[*]}" >&2
    exit 1
fi

if [[ ! -f "$BIFROST_HOST_KEY_PATH" ]]; then
    echo "ERROR: SSH key not found at $BIFROST_HOST_KEY_PATH" >&2
    exit 1
fi

# Detect host platform from env vars passed through by docker compose.
# $OS is set to 'Windows_NT' on Windows; $SHELL is set on Unix hosts.
if [[ "${OS:-}" == "Windows_NT" ]]; then
    host_platform="windows"
else
    host_platform="unix"
fi
ssh_target="${BIFROST_HOST_USER}@${BIFROST_HOST_ADDR}"
ssh_opts=(
    -i "$BIFROST_HOST_KEY_PATH"
    -o StrictHostKeyChecking=no
    -o BatchMode=yes
)
host_working_dir="${BIFROST_HOST_WORKING_DIR:-}"

quote_powershell_arg() {
    local value=${1//\'/\'\'}
    printf "'%s'" "$value"
}

build_unix_command() {
    local remote_cmd prefix=""
    if [[ -n "$host_working_dir" ]]; then
        prefix="cd $(printf '%q' "$host_working_dir") && "
    fi
    printf -v remote_cmd '%q ' "$@"
    printf '%s%s' "$prefix" "${remote_cmd% }"
}

build_windows_command() {
    local quoted_args=() prefix=""
    local arg
    for arg in "$@"; do
        quoted_args+=("$(quote_powershell_arg "$arg")")
    done

    if [[ -n "$host_working_dir" ]]; then
        prefix="Set-Location -LiteralPath $(quote_powershell_arg "$host_working_dir"); "
    fi

    local joined_args="${quoted_args[*]}"
    printf '%s& %s' "$prefix" "$joined_args"
}

if [[ $# -eq 0 ]]; then
    if [[ "$host_platform" == "windows" ]]; then
        exec ssh -t "${ssh_opts[@]}" "$ssh_target" powershell.exe
    fi
    exec ssh -t "${ssh_opts[@]}" "$ssh_target"
fi

case "$host_platform" in
    unix)
        remote_cmd="$(build_unix_command "$@")"
        exec ssh "${ssh_opts[@]}" "$ssh_target" "$remote_cmd"
        ;;
    windows)
        remote_cmd="$(build_windows_command "$@")"
        exec ssh "${ssh_opts[@]}" "$ssh_target" powershell.exe -NoProfile -Command "$remote_cmd"
        ;;
esac
