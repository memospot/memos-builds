#!/usr/bin/env bash
# This custom entry point makes debugging image builds easier.

MAIN=/opt/memos/memos

set -eu
ln -snf /usr/share/zoneinfo/$TZ /etc/localtime

machinearch=$(uname -m)
release=$(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)
platform=$(cat /opt/memos/buildinfo | grep TARGETPLATFORM | cut -d'=' -f2)
checksum=$(sha256sum ${MAIN} | cut -d' ' -f1)
bindate=$(stat -c %y ${MAIN} | cut -d'.' -f1)
baseimage=$(echo $release | cut -d' ' -f1)

print () { echo -e $@ ; }
if [ "${baseimage}" = "Debian" ]; then
    print () { echo $@ ; }
fi

reset="\e\033[0m"
green="\e[32m"
magenta="\e[35m"

print "${magenta}Timezone:          ${green}$TZ${reset}"
print "${magenta}Base image:        ${green}$release${reset}"
print "${magenta}Target platform:   ${green}$platform${reset}"
print "${magenta}Host Architecture: ${green}$machinearch${reset}"
echo ""
print "${magenta}Main binary date:     ${green}$bindate${reset}"
print "${magenta}Main binary checksum: ${green}$checksum${reset}"
echo ""

exec ${MAIN}
