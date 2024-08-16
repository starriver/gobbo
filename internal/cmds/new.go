package cmds

import (
	"errors"
	"os"

	cli "github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/internal/opts"
	"gitlab.com/starriver/gobbo/pkg/godot"
	"gitlab.com/starriver/gobbo/pkg/template"
)

const description = `
Scaffolds a new project in {PROJECT}/. {PROJECT}/ must not already exist -
unless {-b/--bare} is supplied, in which case only a project config file will
be generated at {PROJECT}/gobbo.toml.

This defaults to checking for the current stable Godot version (like
{-g/--godot stable}). An internet connection is required, so if it isn't
available, supply a concrete Godot version.

When using a template or example, {@} can be used to specify a ref - eg.:

  gobbo new --template github.com/user/template{@v1} templated-project
`

var New = cli.Command{
	Name:        "new",
	Headline:    "Create a new project",
	Description: description,
	Options: []cli.Option{
		opts.Log,
		opts.Store,
		opts.Godot,
		{
			Short:    'b',
			Long:     "bare",
			Flag:     true,
			Headline: "Generate gobbo.toml only",
		},
	},

	Args: cli.Args{
		Count:    1,
		Metavars: []string{"PROJECT"},
	},

	Run: func(r *cli.Result) {
		opts.LogSetup(r)

		var path string
		if len(r.Args) == 1 {
			path = r.Args[0]
			if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
				r.Errorf("project path already exists: '%s'", path)
			}
		}

		// store := opts.StoreSetup(r)

		opt := r.Options["g"]
		var g godot.Official
		var err error
		if opt.IsSet {
			g, err = godot.Parse(opt.Value)
			if err != nil {
				r.Error(err)
			}
		} else {
			g, err = godot.CurrentRelease(false)
			if err != nil {
				r.Error(err)
			}
		}

		if r.Fail {
			return
		}

		template.Generate("", path, g, r.Options["b"].IsSet)
	},
}
