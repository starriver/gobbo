package cmds

import (
	"os"

	"github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/internal/opts"
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

var New = charli.Command{
	Name:        "new",
	Headline:    "Create a new project",
	Description: description,
	Options: []charli.Option{
		{
			Short:    'b',
			Long:     "bare",
			Flag:     true,
			Headline: "Generate gobbo.toml only",
		},

		// TODO: templates, examples.
	},

	Args: charli.Args{
		Count:    1,
		Metavars: []string{"PROJECT"},
	},

	Run: func(r *charli.Result) {
		opts.LogSetup(r)

		//store := opts.StoreSetup(r)

		godot := opts.GodotSetup(r, true)

		var path string
		if len(r.Args) == 1 {
			path = r.Args[0]
			_, err := os.Stat(path)
			if !os.IsNotExist(err) {
				if err != nil {
					r.Errorf("couldn't stat '%s': %s", path, err)
				} else {
					r.Errorf("project path already exists: '%s'", path)
				}
			}
		}

		bare := r.Options["b"].IsSet

		if r.Fail {
			return
		}

		template.Generate("", path, godot, bare)
	},
}
