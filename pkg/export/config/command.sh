#!/bin/bash
set -euxo pipefail

# Gobbo export image command script

# Create a writeable copy of the source.
cp -r /srv/src-ro /srv/src
cd /srv/src

$SCRIPT_PRE

EXPORT_FLAG='--export-release'
if [ "$EXPORT_DEBUG" == 1 ]; then
	EXPORT_FLAG='--export-debug'
fi

FILENAME="$PROJECT_NAME-$PROJECT_VERSION-"
if [ -n "$EXPORT_VARIANT" ]; then
	FILENAME="$FILENAME-$EXPORT_VARIANT"
fi
FILENAME="$FILENAME-$EXPORT_PRESET.zip"

# Do the export.
"$GODOT_PATH" "$EXPORT_FLAG" "$EXPORT_PRESET" "/srv/dist/$FILENAME"

$SCRIPT_POST
