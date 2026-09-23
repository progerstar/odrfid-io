#!/usr/bin/env bash

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repo_dir=$(cd -- "${script_dir}/.." && pwd)
app_name=odrfid-io
app_version=${APP_VERSION:-dev}

if [[ ! "${app_version}" =~ ^[A-Za-z0-9][A-Za-z0-9._-]*$ ]]; then
    printf 'Invalid APP_VERSION: %s\n' "${app_version}" >&2
    return 2
fi

write_checksum()
{
    local archive=$1
    local archive_dir archive_name
    archive_dir=$(dirname -- "${archive}")
    archive_name=$(basename -- "${archive}")
    (
        cd "${archive_dir}"
        if command -v sha256sum >/dev/null 2>&1; then
            sha256sum "${archive_name}"
        else
            shasum -a 256 "${archive_name}"
        fi
    ) > "${archive}.sha256"
}
