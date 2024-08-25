package cmds

import (
	"os"

	"github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/internal/opts"
	"gitlab.com/starriver/gobbo/pkg/project"
)

const installDesc = `
Install Godot and packages

Installs the relevant Godot version and packages for the given project.

Alternatively, {-g}/{--godot} can also be used to install an arbitrary Godot
version without a project.

If {-c}/{--check} is supplied, installed dependencies will be verified. No
installations will occur, and the program will exit with an error code if any
dependencies are missing.
`

var Install = charli.Command{
	Name:        "install",
	Headline:    "Install Godot & packages",
	Description: installDesc,
	Options: []charli.Option{
		{
			Short:    'e',
			Long:     "export-templates",
			Flag:     true,
			Headline: "Install Godot export templates",
		},
		{
			Short:    'n',
			Long:     "no-cache",
			Flag:     true,
			Headline: "Disable caching and reinstall dependencies",
		},

		// TODO: --check, --only, --no-cache [godot|packages|all]
	},

	Run: func(r *charli.Result) {
		opts.LogSetup(r)

		store := opts.StoreSetup(r)

		godot := opts.GodotSetup(r, store, true)

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

		project.Generate("", path, godot, bare)
	},
}
