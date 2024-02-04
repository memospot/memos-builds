#!/usr/bin/env bash
# This custom entry point makes debugging image builds easier.

MAIN=/opt/memos/memos

set -eu
ln -snf /usr/share/zoneinfo/$TZ /etc/localtime || true
printf "$TZ\n" > /etc/timezone

release=$(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)
platform=$(cat /opt/memos/buildinfo | grep TARGETPLATFORM | cut -d'=' -f2)
machinearch=$(uname -m)
bindate=$(stat -c %y ${MAIN} | cut -d'.' -f1)
checksum=$(sha256sum ${MAIN} | cut -d' ' -f1)

cyan="\033[36m"
green="\033[32m"
magenta="\033[35m"
yellow="\033[33m"
reset="\033[0m"

printf "${magenta}Timezone:          ${green}$TZ${reset}\n"
printf "${magenta}Base image:        ${green}$release${reset}\n"
printf "${yellow}Target platform:   ${green}$platform${reset}\n"
printf "${yellow}Host Architecture: ${green}$machinearch${reset}\n"
printf "\n"
printf "${cyan}Main binary date:     ${green}$bindate${reset}\n"
printf "${cyan}Main binary checksum: ${green}$checksum${reset}\n"
printf "\n"

exec ${MAIN}
