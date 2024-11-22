![Gobbo](./.logo.svg)

**Gobbo** is a CLI toolchain for Godot 4.x.

This is heavily WIP.

## Current `gobbo.toml` schema

```toml
godot = "4.3"  # required
src = "src"    # optional
```

- `godot` is the Godot version to use. Pre-release versions can be accessed with a suffix, eg. `-beta1`. Currently only official builds are supported, build-from-source is planned.
- `src` is the path of the Godot project root (ie. containing `project.godot`). If omitted, defaults to `src/`.

## Planned export stuff

By default, running `gobbo export` will build all exports in parallel and output them to `dist/`. They'll default to release builds, but `--debug` will be able to be specified.

To implement export matrices, `export` table(s) will be used in `gobbo.toml`. Note that this is to be used *with* Godot's `export_presets.cfg`.

`[export]` will be used by itself to configure global options:

```toml
[export]
only = ["windows", "linux"] # Specify only some exports to be built
dist = "my-dist-dir/" # Change dist/ path
```

Subtables will be used to create a matrix. For example, to produce a matrix of builds for Itch and Steam:

```toml
[export.itch]
# Simply having this table here is enough to add a row to the matrix for Itch
# builds. In this example, they're simply the unaltered Godot exports.

[export.steam]
only = ["windows", "macos", "linux"]
volumes = { "./steam" = "/opt/steam" } # Mount files into the container
scripts = { # Run scripts before/after the Godot export.
	pre = "/opt/steam/prepare.sh"
	post = "/opt/steam/finalise.sh"
}
elective = true # If set, disables this matrix row unless it's explicitly
                # specified (eg. `gobbo export itch steam`)
```

## Planned package management stuff

To implement package management, a `dependencies` table will be used in `gobbo.toml`, in a similar format to `go.mod`:

```toml
[dependencies]
"gitlab.com/snoopdouglas/abfi" = "v3.0.0"
```

The above will pull the abfi plugin into `src/.gobbo/abfi`. Any of abfi's dependencies will also go into `src/.gobbo`.

Version resolution will likely take some cues from Terraform, as Godot plugins tend to do lots of global things - and it's quite rare for plugins to depend on one another right now. If two different major versions of the same package are required, just error.

We'll always use the highest-specified version of a package from the project and its dependencies.

## License

[ISC](./LICENSE)
