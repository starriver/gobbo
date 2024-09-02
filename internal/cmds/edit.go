package cmds

import (
	"github.com/starriver/charli"
	"gitlab.com/starriver/gobbo/internal/opts"
	"gitlab.com/starriver/gobbo/pkg/exec"
	"gitlab.com/starriver/gobbo/pkg/glog"
)

const editDesc = `
Runs the Godot editor for a project. To edit a project outside of the
current directory, use {-p}/{--project}.

Unless {-n}/{--no-install} is supplied, {gobbo install} will run first.

Extraneous arguments will be passed to Godot. Use {--} to prevent Gobbo
parsing flags.

Unless {-f}/{--foreground} is supplied, the editor will run in the background
and its logs will be silenced. However, if it errors out quickly, its logs
will be shown.
`

var Edit = charli.Command{
	Name:        "edit",
	Headline:    "Run Godot editor",
	Description: editDesc,
	Options: []charli.Option{
		opts.Project,
		{
			Short:    'f',
			Long:     "foreground",
			Flag:     true,
			Headline: "Run Godot in the foreground (don't detach)",
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
		args := append(
			[]string{"-e", project.GodotConfigPath()},
			r.Args...,
		)

		if r.Options["f"].IsSet {
			exec.Execv(bin, args)
			panic("Should be unreachable")
		}

		glog.Infof("Running Godot %s editor...", godot.String())
		err = exec.Runway(bin, args)
		if err != nil {
			r.Error(err)
		}
	},
}
