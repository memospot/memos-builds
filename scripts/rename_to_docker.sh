#!/usr/bin/env bash

reset="\033[0m"
cyan="\033[36m"
green="\033[32m"

set -eu

printf "${cyan}Renaming builds to Docker format...${reset}\n"
for dir in ./build/backend/memos_linux_*; do
    if [ ! -d "${dir}" ]; then
        continue
    fi

    already_renamed=$(printf "${dir}" | grep -E "armv[5-7]$|amd64v[2-4]$|_386$|arm64$" || true)
    if [ -n "${already_renamed}" ]; then
        continue
    fi

    new_name="${dir}"
    new_name=$(printf "${new_name}" | sed 's/amd64_v1/amd64/g')
    new_name=$(printf "${new_name}" | sed -E "s/(amd64_v)([2-4])/amd64v\2/")
    new_name=$(printf "${new_name}" | sed -E "s/(arm_)([5-7])/armv\2/")

    if [ "${new_name}" = "${dir}" ]; then
        continue
    fi

    dir=$(echo "$dir" | sed 's/\\\\/\//g')
    new_name=$(echo "$new_name" | sed 's/\\\\/\//g')

    printf "${green}Renaming ${dir} to ${new_name}${reset}\n"
    mv -f "$dir" "$new_name"
done
