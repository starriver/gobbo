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
