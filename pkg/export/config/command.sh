#!/bin/bash
set -euxo pipefail

# Gobbo export image command script

# Create a writeable copy of the source.
cp -r /srv/src-ro /srv/src
cd /srv/src

# Place editor settings.
cp /opt/editor_settings.tres \
	"$HOME/.local/share/godot/editor_settings-$GODOT_SETTINGS_VERSION.tres"

# Run pre-script.
$SCRIPT_PRE

# Prepare flags.

EXPORT_FLAG='--export-release'
if [ "$EXPORT_DEBUG" == 1 ]; then
	EXPORT_FLAG='--export-debug'
fi

FILENAME="$PROJECT_NAME-$PROJECT_VERSION"
if [ -n "$EXPORT_VARIANT" ]; then
	FILENAME="$FILENAME-$EXPORT_VARIANT"
fi
FILENAME="$FILENAME-$EXPORT_PRESET"

# Do the export.
"$GODOT_PATH" "$EXPORT_FLAG" "$EXPORT_PRESET" "/srv/dist/$FILENAME.$EXTENSION"

# Zip output if specified.
if [ "$ZIP" == 1 ] && [ "$EXTENSION" != zip ]; then
(
	cd /srv/dist
	zip -mr "$FILENAME.zip" *
)
fi

# Run post-script.
$SCRIPT_POST
