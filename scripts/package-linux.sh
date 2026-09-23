#!/usr/bin/env bash
set -euo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/common.sh"

[[ "$(uname -s)" == Linux && "$(uname -m)" == x86_64 ]] || {
    printf 'Linux package requires an x86-64 Linux host.\n' >&2
    exit 2
}

dist_dir=${1:-"${repo_dir}/dist/linux"}
mkdir -p "${dist_dir}"
dist_dir=$(cd -- "${dist_dir}" && pwd)
stage_dir=$(mktemp -d)
trap 'rm -rf -- "${stage_dir}"' EXIT
package_dir="${stage_dir}/${app_name}"
mkdir -p "${package_dir}"

(
    cd "${repo_dir}"
    CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
        -trimpath -buildvcs=false -ldflags='-s -w' \
        -o "${package_dir}/${app_name}" .
)

"${package_dir}/${app_name}" -h >/dev/null 2>&1
if ldd "${package_dir}/${app_name}" | grep -q 'not found'; then
    printf 'The Linux binary has unresolved shared libraries.\n' >&2
    exit 1
fi

cp "${repo_dir}/README.md" "${repo_dir}/LICENSE" "${package_dir}/"
archive="${dist_dir}/${app_name}-${app_version}-linux-x86_64.tar.gz"
tar -C "${stage_dir}" -czf "${archive}" "${app_name}"
write_checksum "${archive}"
printf 'Created %s\n' "${archive}"
