#!/usr/bin/env ash
# shellcheck shell=dash
# This custom entry point makes debugging image builds easier.

MAIN=/opt/memos/memos

set -eu
ln -snf "/usr/share/zoneinfo/$TZ" /etc/localtime || true
printf %s "$TZ" >/etc/timezone

release=$(grep PRETTY_NAME </etc/os-release | cut -d'"' -f2)
platform=$(grep TARGETPLATFORM </opt/memos/buildinfo | cut -d'=' -f2)
bindate=$(stat -c %y ${MAIN} | cut -d'.' -f1)
checksum=$(sha256sum ${MAIN} | cut -d' ' -f1)

cyan="\033[36m"
green="\033[32m"
magenta="\033[35m"
yellow="\033[33m"
reset="\033[0m"

printf "\n%bTimezone:             %b%s%b\n" "$magenta" "$green" "$TZ" "$reset"
printf "%bBase image:           %b%s%b\n" "$magenta" "$green" "$release" "$reset"
printf "%bTarget platform:      %b%s%b\n" "$yellow" "$green" "$platform" "$reset"
printf "%bHost Architecture:    %b%s%b\n" "$yellow" "$green" "$(uname -m)" "$reset"
printf "%bMain binary date:     %b%s%b\n" "$cyan" "$green" "$bindate" "$reset"
printf "%bMain binary checksum: %b%s%b\n" "$cyan" "$green" "$checksum" "$reset"

file_env() {
  var="$1"
  fileVar="${var}_FILE"

  val_var="$(printenv "$var")"
  val_fileVar="$(printenv "$fileVar")"

  if [ -n "$val_var" ] && [ -n "$val_fileVar" ]; then
    echo "error: both $var and $fileVar are set (but are exclusive)" >&2
    exit 1
  fi

  if [ -n "$val_var" ]; then
    val="$val_var"
  elif [ -n "$val_fileVar" ]; then
    val="$(cat "$val_fileVar")"
  fi

  export "$var"="$val"
  unset "$fileVar"
}

file_env "MEMOS_DSN"

exec ${MAIN} "$@"
