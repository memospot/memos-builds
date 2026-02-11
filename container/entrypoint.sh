#!/usr/bin/env ash
# shellcheck shell=ash
#
# This custom entry point does a few things:
# - Sets the timezone.
# - Fixes ownership of the data directory.
# - Loads environment variables from a file.
# - Loads DSN from a secret file.
# - Cleans up the demo database if running in demo mode.
# - Prints some information about the container.
#
# It's saved as `/init` inside containers.
#
#! Syntax must be compatible with the BusyBox ash shell.

cyan="\033[36m"
green="\033[32m"
magenta="\033[35m"
red="\033[31m"
yellow="\033[33m"
reset="\033[0m"

set_timezone() {
  # Set timezone if running as root (for system calls that read /etc/localtime).
  # Most applications use the TZ environment variable.
  if [ "$(id -u)" = "0" ]; then
    ln -snf "/usr/share/zoneinfo/$TZ" /etc/localtime || true
    printf "%s\n" "$TZ" >/etc/timezone
    printf "%s\n" "$TZ" >/etc/TZ
  fi
}
set_timezone

# Fix ownership of data directory for users upgrading
# from older versions where files were created as root.
fix_permissions() {
  script="$1"
  shift  # Remove first argument, leaving "$@" with remaining args.
  PUID=${PUID:-10001}
  PGID=${PGID:-10001}

  # Allow running as root if PUID is explicitly set as 0.
  # Otherwise, set up nonroot user and group.
  if [ "$PUID" != "0" ] && [ "$(id -u)" = "0" ]; then
    # Ensure nonroot user and group exist with correct IDs.
    # Delete and recreate to safely handle ID changes on Busybox.
    if id nonroot >/dev/null 2>&1; then
      current_user_uid=$(id -u nonroot)
      current_user_gid=$(id -g nonroot)
      if [ "$current_user_uid" != "$PUID" ] || [ "$current_user_gid" != "$PGID" ]; then
        printf "%bEntrypoint: Updating nonroot user ID & Group from %b%s%b to %b%s%b\n" "$cyan" "$green" "$current_user_uid:$current_user_gid" "$cyan" "$green" "$PUID:$PGID" "$reset"
        deluser nonroot || true
        delgroup nonroot || true
      fi
    fi

    if ! grep -q "^nonroot:" /etc/group; then
      addgroup -g "$PGID" -S nonroot
    fi

    if ! id nonroot >/dev/null 2>&1; then
      adduser -u "$PUID" -S -G nonroot -h "$MEMOS_DATA" nonroot
    fi
  fi

  # Check current ownership and only chown data directory if needed.
  if [ -d "$MEMOS_DATA" ]; then
    current_uid=$(stat -c %u "$MEMOS_DATA" 2>/dev/null || echo "")
    current_gid=$(stat -c %g "$MEMOS_DATA" 2>/dev/null || echo "")

    if [ "$current_uid" != "$PUID" ] || [ "$current_gid" != "$PGID" ]; then
      printf "%bFixing ownership of %b%s%b. %b%s%b -> %b%s%b\n" "$yellow" "$green" "$MEMOS_DATA" "$yellow" "$green" "$current_uid:$current_gid" "$cyan" "$green" "$PUID:$PGID" "$reset"
      chown -R "$PUID:$PGID" "$MEMOS_DATA" 2>/dev/null || true
    fi
  fi

  # Allow running as root if PUID is explicitly set as 0.
  # Otherwise, re-execute as non-root user.
  if [ "$PUID" != "0" ] && [ "$(id -u)" = "0" ]; then
    exec su-exec "$PUID:$PGID" "$script" "$@"
  fi
}
fix_permissions "$0" "$@"

# Clean up demo database if running in demo mode to prevent migration issues.
cleanup_demo_db() {
  if [ "$MEMOS_DEMO" = "true" ] || [ "$MEMOS_MODE" = "demo" ]; then
    for db_file in memos_demo.db memos_demo.db-shm memos_demo.db-wal; do
      if [ -f "$MEMOS_DATA/$db_file" ]; then
        printf "%bCleaning up demo database file: %b%s%b\n" "$yellow" "$magenta" "$MEMOS_DATA/$db_file" "$reset"
        rm -f "$MEMOS_DATA/$db_file"
      fi
    done
  fi
}
cleanup_demo_db

# Load environment variables from supplied env file, if present.
load_env() {
  env_file="$1"
  [ -f "$env_file" ] || return

  printf "%bLoading environment from %b%s%b. %bThis file can override environment variables passed to the container.%b\n" "$cyan" "$green" "$env_file" "$reset" "$magenta" "$reset"

  while IFS= read -r line || [ -n "$line" ]; do
    # Skip comments and empty lines
    case "$line" in
      \#*|"") continue ;;
    esac

    # Split into key and value
    key="${line%%=*}"
    val="${line#*=}"

    # Skip if key is empty
    [ -n "$key" ] || continue

    # Remove surrounding quotes if present
    val="${val%\"}"
    val="${val#\"}"
    val="${val%\'}"
    val="${val#\'}"

    export "$key"="$val"
  done < "$env_file"
}
load_env "${MEMOS_DATA}/memos.env"

# Load an env var from either the var itself or from a file path given by VAR_NAME_FILE.
file_env() {
  env_var_name="$1"
  file_path_var_name="${env_var_name}_FILE"

  value_from_env="$(printenv "$env_var_name" || true)"
  path_to_secret_file="$(printenv "$file_path_var_name" || true)"

  if [ -n "$value_from_env" ] && [ -n "$path_to_secret_file" ]; then
    echo "error: both $env_var_name and $file_path_var_name are set (but they are mutually exclusive)" >&2
    exit 1
  fi

  if [ -n "$value_from_env" ]; then
    resolved_value="$value_from_env"
  elif [ -n "$path_to_secret_file" ]; then
    set +x
    resolved_value="$(cat "$path_to_secret_file")"
  fi

  export "$env_var_name"="$resolved_value"
  unset "$file_path_var_name"
}

# Auto-load mounted DSN secret with the same ID as the env var.
if [ -f "/run/secrets/MEMOS_DSN" ]; then
  export MEMOS_DSN_FILE="/run/secrets/MEMOS_DSN"
fi
# This supports Docker secrets: set MEMOS_DSN_FILE=/run/secrets/dsn env var and
# mount the secret there instead of passing plain-text passwords in MEMOS_DSN.
file_env "MEMOS_DSN"

MAIN="$(realpath "$1")"
release=$(grep PRETTY_NAME </etc/os-release | cut -d'"' -f2)
platform=$(grep TARGETPLATFORM </usr/share/memos/buildinfo | cut -d'=' -f2)
bindate=$(stat -c %y "${MAIN}" | cut -d'.' -f1)
checksum=$(grep SHA256SUM </usr/share/memos/buildinfo | cut -d'=' -f2)

printf "\n%bTimezone:              %b%s%b\n" "$cyan" "$green" "$TZ" "$reset"
printf "%bBase image:            %b%s%b\n" "$cyan" "$green" "$release" "$reset"
printf "%bTarget platform:       %b%s%b\n" "$cyan" "$green" "$platform" "$reset"
printf "%bHost Architecture:     %b%s%b\n" "$cyan" "$green" "$(uname -m)" "$reset"
printf "%bMain binary date:      %b%s%b\n" "$cyan" "$green" "$bindate" "$reset"
printf "%bMain binary sha256sum: %b%s%b\n" "$cyan" "$green" "$checksum" "$reset"
printf "%bRunning as:            %b%s%b (PUID=%b%s%b, PGID=%b%s%b)\n" "$cyan" "$magenta" "$(id -un)" "$cyan" "$magenta" "$PUID" "$cyan" "$magenta" "$PGID" "$reset"

printf "%bStarting Memos…%b\n" "$green" "$reset"
exec "$@"
