package cmds

import (
	"os"

	"github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/internal/opts"
	"gitlab.com/starriver/gobbo/pkg/exec"
	"gitlab.com/starriver/gobbo/pkg/glog"
)

const runDesc = `
Runs Godot in the foreground. If in a project directory, or {-p}/{--project} is
supplied, Godot will run the project.

Unless {-n}/{--no-install} is supplied, {gobbo install} will run first.

Extraneous arguments will be passed to Godot. Use {--} to prevent Gobbo
parsing flags.
`

var Run = charli.Command{
	Name:        "run",
	Headline:    "Run a project",
	Description: editDesc,
	Options: []charli.Option{
		opts.Project,
		{
			Short:    'b',
			Long:     "background",
			Flag:     true,
			Headline: "Run Godot in the background (don't attach)",
		},
		{
			Short:    'n',
			Long:     "no-install",
			Flag:     true,
			Headline: "Skip dependency check and installation",
		},
	},
	Args: charli.Args{
		Varadic:  true,
		Metavars: []string{"GODOT_ARGS"},
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

		bin := store.Join("bin", godot.BinaryPath(&store.Platform))

		// chdir first or Godot won't run the project without the editor.
		glog.Debugf("cd: '%s'", project.Src)
		os.Chdir(project.Src)

		if r.Options["b"].IsSet {
			glog.Infof(
				"Running Godot %s in the background...",
				godot.String(),
			)
			err = exec.Runway(bin, r.Args)
			if err != nil {
				r.Error(err)
			}
			return
		}

		exec.Execv(bin, r.Args)

		panic("Should be unreachable")
	},
}
