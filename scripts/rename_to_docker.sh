reset="\e\033[0m"
cyan="\e[36m"
green="\e[32m"

set -eu

echo -e "${cyan}Renaming builds to Docker format...${reset}"
for dir in ./build/backend/memos_linux_*; do
    if [[ ! -d "${dir}" ]]; then
        continue
    fi

    already_renamed=$(echo "${dir}" | grep -E "armv[5-7]$|amd64v[2-4]$|_386$|arm64$" || true)
    if [[ -n "${already_renamed}" ]]; then
        continue
    fi

    new_name="${dir}"
    new_name="${new_name//amd64_v1/amd64}"
    new_name=$(echo "${new_name}" | sh -c 'sed -E "s/(amd64_v)([2-4])/amd64v\2/"')
    new_name=$(echo "${new_name}" | sh -c 'sed -E "s/(arm_)([5-7])/armv\2/"')

    if [[ "${new_name}" == "${dir}" ]]; then
        continue
    fi

    dir=${dir//\\//}
    new_name=${new_name//\\//}

    echo -e "${green}Renaming ${dir} to ${new_name}${reset}"
    mv -f "$dir" "$new_name"
done
