#!/usr/bin/env bash
set -euo pipefail
source "$(dirname -- "${BASH_SOURCE[0]}")/common.sh"

[[ "$(uname -s)" == Darwin ]] || {
    printf 'macOS package requires a macOS host.\n' >&2
    exit 2
}

dist_dir=${1:-"${repo_dir}/dist/macos"}
mkdir -p "${dist_dir}"
dist_dir=$(cd -- "${dist_dir}" && pwd)
stage_dir=$(mktemp -d)
trap 'rm -rf -- "${stage_dir}"' EXIT
package_dir="${stage_dir}/${app_name}"
mkdir -p "${package_dir}"

for arch in amd64 arm64; do
    clang_arch=x86_64
    [[ "${arch}" == arm64 ]] && clang_arch=arm64
    (
        cd "${repo_dir}"
        CGO_ENABLED=1 GOOS=darwin GOARCH="${arch}" CC="clang -arch ${clang_arch}" \
            go build -trimpath -buildvcs=false -ldflags='-s -w' \
            -o "${stage_dir}/${app_name}-${arch}" .
    )
done

lipo -create -output "${package_dir}/${app_name}" \
    "${stage_dir}/${app_name}-amd64" "${stage_dir}/${app_name}-arm64"
architectures=$(lipo -archs "${package_dir}/${app_name}")
[[ " ${architectures} " == *' x86_64 '* && " ${architectures} " == *' arm64 '* ]] || {
    printf 'Universal binary is missing an architecture: %s\n' "${architectures}" >&2
    exit 1
}
codesign --force --sign - "${package_dir}/${app_name}"
codesign --verify --strict "${package_dir}/${app_name}"
"${package_dir}/${app_name}" -h >/dev/null 2>&1

cp "${repo_dir}/README.md" "${repo_dir}/LICENSE" "${package_dir}/"
archive="${dist_dir}/${app_name}-${app_version}-macos-universal.tar.gz"
tar -C "${stage_dir}" -czf "${archive}" "${app_name}"
write_checksum "${archive}"
printf 'Created %s\n' "${archive}"
