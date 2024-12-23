package cmds

import (
	"github.com/starriver/charli"
	"github.com/starriver/gobbo/internal/opts"
	"github.com/starriver/gobbo/pkg/glog"
)

const installDesc = `
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
		opts.Project,
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

		_, godot := opts.ProjectGodotSetup(r, store, opts.Never, false)

		noCache := r.Options["n"].IsSet

		if r.Fail {
			return
		}

		installed, err := store.IsGodotInstalled(godot)
		if err != nil {
			r.Error(err)
			return
		}

		if installed && !noCache {
			glog.Infof("Godot %s already installed.", godot.String())
			return
		}

		err = store.InstallGodot(godot)
		if err != nil {
			r.Error(err)
		}
	},
}
