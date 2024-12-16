![Gobbo](./.logo.svg)

**Gobbo** is a CLI toolchain for Godot 4.x.

This is heavily WIP.

## Current `gobbo.toml` schema

`godot` is the only required key. Defaults are listed here.

```toml
godot = "4.3"
src = "src"

[export]
only = []
dist = "dist"
zip = false
volumes = []
scripts.pre = ""
scripts.post = ""

[export.variant]
only = []
volumes = []
scripts.pre = ""
scripts.post = ""
elective = false
```

- `godot` is the Godot version to use. Pre-release versions can be accessed with a suffix, eg. `-beta1`. Currently only official builds are supported, build-from-source is planned.
- `src` is the path of the Godot project root (ie. containing `project.godot`).

### Export table

This configures the Gobbo exporter. You should still set up your export presets (in `src/export_presets.cfg`) as normal.

- `only` filters the presets to export.
- `dist` is the path of the finished exports. When exporting, it will be created if it doesn't exist.
- `zip` automatically zips all exports if true. Otherwise, exports will be in `dist` subdirectories.
- `volumes` allows for [short-form Docker volumes](https://docs.docker.com/reference/cli/docker/container/run/#volume) to be mounted for all build containers. Currently, only the `z`, `Z` and `ro` flags are supported.
- `scripts.*` specifies Bash script hooks to be executed before (`pre`) and after (`post`) the Godot export has executed.

#### Variants

Export **variants** can be specified, which will enable a two-dimensional build matrix (presets vs. variants). Variants are named after their tables.

For each variant's table:

- `only` and `volumes` are merged with their respective values in `[export]`.
- `scripts.*` will **override** their respective values in `[export]`.
- `elective`, if true, disables the variant unless it's explicitly specified in the `gobbo export` command.

For example, to produce a matrix of builds for Itch and Steam:

```toml
[export.itch]
# Simply having this table here is enough to add a row to the matrix for Itch
# builds. In this example, they're simply the unaltered Godot exports.

[export.steam]
only = ["windows", "macos", "linux"]
volumes = { "./steam" = "/opt/steam" }
scripts = {
	pre = "/opt/steam/prepare.sh"
	post = "/opt/steam/finalise.sh"
}
elective = true
```

---

## License

[ISC](./LICENSE)
