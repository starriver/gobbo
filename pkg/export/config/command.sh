#!/bin/bash
set -euo pipefail

# Gobbo export image command script

log() {
	echo "[gobbo] $@"
}

log 'Preparing environment'

# Create a writeable copy of the source.
# NOTE: -a is used because the import files' timestamps need to be later than
# the source files, or Godot will freeze during the export.
cp -a /srv/src-ro /srv/src
cd /srv/src

# Place editor settings.
(
	mkdir -p ~/.config/godot
	cd "$_"
	cp /opt/editor_settings.tres "editor_settings-$GODOT_SETTINGS_VERSION.tres"
)

if [ -n "$SCRIPT_PRE" ]; then
	log 'Running pre-script'
	echo "+ $SCRIPT_PRE"
	$SCRIPT_PRE
fi

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

log 'Exporting'
"$GODOT_PATH" --headless "$EXPORT_FLAG" "$EXPORT_PRESET" "/srv/dist/$FILENAME.$EXTENSION"
# exit 0

if [ "$ZIP" == 1 ] && [ "$EXTENSION" != zip ]; then
(
	log 'Zipping'
	cd /srv/dist
	zip -mr "$FILENAME.zip" *
)
fi

if [ -n "$SCRIPT_POST" ]; then
	log 'Running post-script'
	echo "+ $SCRIPT_POST"
	$SCRIPT_POST
fi
