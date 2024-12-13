package cmds

import (
	"os"
	"os/exec"

	"github.com/starriver/charli"
	"github.com/starriver/gobbo/internal/opts"
	"github.com/starriver/gobbo/pkg/export"
	"github.com/starriver/gobbo/pkg/glog"
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
		// compose up will require that we start all containers immediately, but
		// we can signal them to actually start their builds by eg. placing a
		// file.

		{
			Short:    'n',
			Long:     "no-install",
			Flag:     true,
			Headline: "Abort if dependencies are missing",
		},
		{
			Short:    'd',
			Long:     "debug",
			Flag:     true,
			Headline: "Build debug exports",
		},
		{
			Short:    'c',
			Long:     "compose",
			Flag:     true,
			Headline: "Output Docker Compose config then exit",
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

		err = export.CheckEnvironment()
		if err != nil {
			r.Error(err)
			return
		}

		debug := r.Options["d"].IsSet
		c := export.Configure(store, project, debug, r.Args)

		if r.Options["c"].IsSet {
			err = yaml.NewEncoder(os.Stdout).Encode(c)
			if err != nil {
				r.Error(err)
			}
			return
		}

		err = export.BuildImage()
		if err != nil {
			r.Error(err)
			return
		}

		glog.Info("Importing assets...")
		godotPath := store.Join("bin", godot.BinaryPath(&store.Platform))
		cmd := exec.Command(godotPath, "--no-header", "--headless", "--import", project.GodotConfigPath())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		glog.Debugf("%s %v", cmd.Path, cmd.Args)

		err = cmd.Run()
		if err != nil {
			r.Errorf("Import failed, aborting.")
			return
		}

		glog.Info("Starting exports...")
		export.Run(c)
	},
}
