#!/usr/bin/env bash
export TERM=xterm-256color
reset=$(tput sgr0)
green=$(tput setaf 2)
magenta=$(tput setaf 5)

ln -snf /usr/share/zoneinfo/$TZ /etc/localtime
echo $TZ > /etc/timezone

MAIN=/opt/memos/memos

arch=$(uname -m)
release=$(cat /etc/os-release | grep PRETTY_NAME | cut -d'"' -f2)
platform=$(cat /opt/memos/buildinfo | grep TARGETPLATFORM | cut -d'=' -f2)
checksum=$(sha256sum ${MAIN} | cut -d' ' -f1)
bindate=$(stat -c %y ${MAIN} | cut -d'.' -f1)

echo "${magenta}Timezone: ${green}$TZ${reset}"
echo "${magenta}Container base image: ${green}$release ($arch)${reset}"
echo "${magenta}Container target platform: ${green}$platform${reset}"
echo "${magenta}Main binary date: ${green}$bindate${reset}"
echo "${magenta}Main binary checksum: ${green}$checksum${reset}"

${MAIN}
