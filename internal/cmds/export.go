package cmds

import (
	"os"

	"github.com/starriver/charli"
	"github.com/starriver/gobbo/internal/opts"
	"github.com/starriver/gobbo/pkg/export"
	"gopkg.in/yaml.v3"
)

const exportDesc = `
Builds project exports in parallel using Docker containers. All exports are
built unless {EXPORT}s are specified.

Gobbo builds its own Docker image for the containers, tagged {_gobbo:base}. You
can provide your own by creating a Dockerfile in your project's root directory,
which will tag the image {_gobbo:HASH}, where {HASH} is a hash of your project
path.

Use '{FROM _gobbo:base}' in your Dockerfile to ensure you have all of the
prerequisite dependencies available (unless you know what you're doing!). Note
that Godot will be bind-mounted into the build containers, so there's no need
to install it in your Dockerfile.

Docker Compose is required, and the Docker CLI must be in your {PATH}. If not,
the command will fail.
`

var Export = charli.Command{
	Name:        "export",
	Headline:    "Build project exports",
	Description: exportDesc,
	Options: []charli.Option{
		opts.Project,

		// TODO: -j/--jobs

		{
			Short:    'n',
			Long:     "no-install",
			Flag:     true,
			Headline: "Abort if dependencies are missing",
		},
	},
	Args: charli.Args{
		Varadic:  true,
		Metavars: []string{"EXPORT"},
	},

	Run: func(r *charli.Result) {
		opts.LogSetup(r)

		store := opts.StoreSetup(r)

		installMode := opts.IfAbsent
		if r.Options["n"].IsSet {
			installMode = opts.Never
		}

		project, godot := opts.ProjectGodotSetup(r, store, installMode, true)

		installed, err := store.IsGodotInstalled(godot)
		if err != nil {
			r.Error(err)
		} else if !installed {
			r.Errorf("Godot %s not installed", godot.String())
		}

		if r.Fail {
			return
		}

		// TODO remove
		if project != nil {
		}

		err = export.CheckEnvironment()
		if err != nil {
			r.Error(err)
			return
		}

		c := export.Configure(store, project, r.Args)
		yaml.NewEncoder(os.Stdout).Encode(c)
	},
}
