#!/bin/bash
#
# Remove AIStore metadata from the given host paths.
# With CLEAN_DATA=true, also remove the bucket directories and the objects in them.
#
set -euo pipefail

: "${AIS_PATHS:?AIS_PATHS must be set to a space-separated list of host paths}"
clean_data="${CLEAN_DATA:-false}"

read -ra paths <<< "$AIS_PATHS"

# A path is unsafe unless it is absolute, below the root, and free of empty,
# "." and ".." segments.
for path in "${paths[@]}"; do
    if [[ "$path" != /* || "$path" == "/" || "$path" == *//* || "$path" =~ (^|/)\.\.?(/|$) ]]; then
        echo "refusing unsafe path '$path'" >&2
        exit 1
    fi
done

for path in "${paths[@]}"; do
    target="/host${path%/}"
    if [ ! -d "$target" ]; then
        echo "$path is not a directory on $(hostname), skipping"
        continue
    fi
    # Symlinks can lead back out of the host root.
    resolved=$(cd "$target" && pwd -P)
    if [[ "$resolved" != /host/?* ]]; then
        echo "refusing unsafe path '$path', it resolves outside the host root" >&2
        exit 1
    fi
    echo "removing AIS metadata from $path"
    rm -rf "$target"/.ais.*
    if [ "$clean_data" = true ]; then
        echo "removing bucket data from $path"
        rm -rf "$target"/@*
    fi
done
